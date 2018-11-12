package canlemon

import (
	"github.com/brutella/can"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	canBusName = "can0"
)

const (
	frameOilTemp     uint32 = 0x100
	frameCoolantTemp        = 0x101
	frameFuel               = 0x102
	frameSpeed              = 0x103
)

type IntResultFn func(v int) int

var (
	bus      *can.Bus
	intFnMap map[uint32]IntResultFn
)

func Connect(oilTemp, coolantTemp, fuel IntResultFn) error {
	if bus != nil {
		return errors.New("can bus already connected")
	}
	newBus, err := can.NewBusForInterfaceWithName(canBusName)
	if err != nil {
		return err
	}

	intFnMap[frameOilTemp] = oilTemp
	intFnMap[frameCoolantTemp] = coolantTemp
	intFnMap[frameFuel] = fuel

	newBus.SubscribeFunc(handleFrame)

	bus = newBus
	return err
}

func Disconnect() error {
	if bus == nil {
		return errors.New("can bus not connected")
	}
	return bus.Disconnect()
}

func SendSpeed(speed int) error {
	if bus == nil {
		return errors.New("can bus not connected")
	}
	log.WithField("speed", speed).Debug("sending speed over canbus")
	return bus.Publish(can.Frame{
		ID:     frameSpeed,
		Length: 1,
		Data:   [8]uint8{uint8(speed)},
	})
}

func handleFrame(frame can.Frame) {
	log.WithField("canID", frame.ID).
		WithField("length", frame.Length).
		Debug("received canbus frame")

	fn, ok := intFnMap[frame.ID]
	if !ok {
		log.WithField("canID", frame.ID).
			Error("unknown canID")
		return
	}

	v, err := uint16Result(frame)
	if err != nil {
		log.Error("unable to convert oil temp", err)
	}
	log.WithField("canID", frame.ID).
		WithField("intValue", v).
		Debug("calling callback function")
	fn(v)
}

func uint16Result(frame can.Frame) (int, error) {
	if frame.Length != 2 {
		return 0, errors.Errorf("incorrect frame size for uint16: %v", frame.Length)
	}
	// Little endian
	return int(uint16(frame.Data[0]) + uint16(frame.Data[1])<<8), nil
}
