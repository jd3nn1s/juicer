package lemoncan

import (
	"context"
	"encoding/binary"
	"github.com/brutella/can"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	frameOilTemp     uint32 = 0x100
	frameCoolantTemp        = 0x101
	frameFuel               = 0x102
	frameSpeed              = 0x103
)

type IntResultFn func(v int)

type Callbacks struct {
	OilTemp     IntResultFn
	CoolantTemp IntResultFn
	Fuel        IntResultFn
}

type CANBus interface {
	SubscribeFunc(can.HandlerFunc)
	ConnectAndPublish() error
	Disconnect() error
	Publish(can.Frame) error
}

type Connection struct {
	bus CANBus
	cb  Callbacks
}


func Connect(portName string) (*Connection, error) {
	bus, err := can.NewBusForInterfaceWithName(portName)
	if err != nil {
		return nil, err
	}

	c := &Connection{
		bus: bus,
	}
	return c, nil
}

func (c *Connection) Start(ctx context.Context, cb Callbacks) error {
	c.cb = cb
	c.bus.SubscribeFunc(c.handleFrame)
	log.Info("CAN bus opened and subscribed")

	go func() {
		select {
		case <-ctx.Done():
			log.Info("stopping can bus: %v", ctx.Err())
			if err := c.bus.Disconnect(); err != nil {
				log.WithField("err", err).Warn("unable to disconnect canbus after context")
			}
		}
	}()

	return c.bus.ConnectAndPublish()
}

func (c *Connection) Close() error {
	if c.bus == nil {
		return errors.New("can bus not connected")
	}
	return c.bus.Disconnect()
}

func (c *Connection) SendSpeed(speed int) error {
	if c.bus == nil {
		return errors.New("can bus not connected")
	}
	log.WithField("speed", speed).Debug("sending speed over canbus")
	return c.bus.Publish(can.Frame{
		ID:     frameSpeed,
		Length: 1,
		Data:   [8]uint8{uint8(speed)},
	})
}

func (c *Connection) handleFrame(frame can.Frame) {
	log.WithField("canID", frame.ID).
		WithField("length", frame.Length).
		Debug("received canbus frame")

	var cb IntResultFn
	switch frame.ID {
	case frameOilTemp:
		cb = c.cb.OilTemp
	case frameCoolantTemp:
		cb = c.cb.CoolantTemp
	case frameFuel:
		cb = c.cb.Fuel
	default:
		log.WithField("canID", frame.ID).
			Error("unknown canID")
		return
	}

	if cb == nil {
		log.WithField("canID", frame.ID).Debug("no callback registered")
	}

	v, err := uint16Result(frame)
	if err != nil {
		log.Error("unable to convert to uint16", err)
	}
	log.WithField("canID", frame.ID).
		WithField("intValue", v).
		Debug("calling callback function")
	cb(v)
}

func uint16Result(frame can.Frame) (int, error) {
	if frame.Length != 2 {
		return 0, errors.Errorf("incorrect frame size for uint16: %v", frame.Length)
	}
	return int(binary.LittleEndian.Uint16(frame.Data[0:2])), nil
}
