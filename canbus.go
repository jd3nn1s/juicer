package main

import (
	"context"
	"github.com/jd3nn1s/juicer/lemoncan"
	log "github.com/sirupsen/logrus"
)

type canBus struct {
	c        CANBus
	sendChan chan<- canSensorData
	data     canSensorData
}

func (bus *canBus) Open() error {
	c, err := canBusConnect(canBusPortName)
	bus.c = c
	return err
}

func (bus *canBus) Close() error {
	if bus.c == nil {
		return nil
	}
	return bus.c.Close()
}

func (bus *canBus) Start(ctx context.Context) error {
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

func (bus *canBus) send() {
	select {
	case bus.sendChan <- bus.data:
	default:
	}
}

func (bus *canBus) Name() string {
	return "canbus"
}

var canBusConnect = func(p string) (CANBus, error) {
	return lemoncan.Connect(p)
}

func runCAN(ctx context.Context, sendChan chan<- canSensorData) {
	err := retry(ctx, &canBus{
		sendChan: sendChan,
	})
	if err != nil {
		log.Errorf("canbus done: %v", err)
	}
}