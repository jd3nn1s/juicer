package juicer

import (
	"context"
	"github.com/jd3nn1s/juicer/lemoncan"
	log "github.com/sirupsen/logrus"
)

type canBusRetryable struct {
	c        CANBus
	sendChan chan<- canSensorData
	data     canSensorData
}

func (bus *canBusRetryable) Open() error {
	c, err := canBusConnect(canBusPortName)
	bus.c = c
	return err
}

func (bus *canBusRetryable) Close() error {
	if bus.c == nil {
		return nil
	}
	return bus.c.Close()
}

func (bus *canBusRetryable) Start(ctx context.Context) error {
	return bus.c.Start(ctx, lemoncan.Callbacks{
		Fuel: func(v int) {
			bus.data.FuelLevel = v
			bus.send()
		},
		CoolantTemp: func(v int) {
			bus.data.CoolantTemp = v
			bus.send()
		},
		OilTemp: func(v int) {
			bus.data.OilTemp = v
			bus.send()
		},
	})
}

func (bus *canBusRetryable) send() {
	select {
	case bus.sendChan <- bus.data:
	default:
	}
}

func (bus *canBusRetryable) Name() string {
	return "canbus"
}

var canBusConnect = func(p string) (CANBus, error) {
	return lemoncan.Connect(p)
}

func newCANBus(sendChan chan<- canSensorData) *canBusRetryable {
	return &canBusRetryable{
		sendChan: sendChan,
	}
}

func (bus *canBusRetryable) runCAN(ctx context.Context) {
	err := retry(ctx, bus)
	if err != nil {
		log.Errorf("canbus done: %v", err)
	}
}

func (bus *canBusRetryable) CANBus() CANBus {
	return bus.c
}