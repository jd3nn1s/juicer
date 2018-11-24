package juicer

import (
	"context"
	"github.com/jd3nn1s/juicer/lemoncan"
	"github.com/jd3nn1s/kw1281"
	"github.com/jd3nn1s/skytraq"
)

type sensorStub struct {
	startChan chan struct{}
	errChan   chan error
	fnChan    chan func()
}

type kw1281Stub struct {
	sensorStub
	callbacks kw1281.Callbacks
}

type skytraqStub struct {
	sensorStub
	callbacks skytraq.Callbacks
}

type canBusStub struct {
	sensorStub
	speed int
	speedCallCount int
	callbacks lemoncan.Callbacks
}

func createSensorStub() *sensorStub {
	ret := sensorStub{
		startChan: make(chan struct{}),
		errChan:   make(chan error),
		fnChan:    make(chan func()),
	}
	return &ret
}

func (s *sensorStub) Close() error {
	return nil
}

func (s *sensorStub) start(ctx context.Context) error {
	select {
	case s.startChan <- struct{}{}:
	default:
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-s.errChan:
			return err
		case fn := <-s.fnChan:
			fn()
		}
	}
}

func createECUStub() *kw1281Stub {
	return &kw1281Stub{
		sensorStub: *createSensorStub(),
	}
}

func (k *kw1281Stub) Start(ctx context.Context, callbacks kw1281.Callbacks) error {
	k.callbacks = callbacks
	return k.sensorStub.start(ctx)
}

func createGPSStub() *skytraqStub {
	return &skytraqStub{
		sensorStub: *createSensorStub(),
	}
}

func (k *skytraqStub) Start(ctx context.Context, callbacks skytraq.Callbacks) error {
	k.callbacks = callbacks
	return k.sensorStub.start(ctx)
}

func createCANBusStub() *canBusStub {
	return &canBusStub{
		sensorStub: *createSensorStub(),
	}
}

func (c *canBusStub) Start(ctx context.Context, callbacks lemoncan.Callbacks) error {
	c.callbacks = callbacks
	return c.sensorStub.start(ctx)
}

func (c *canBusStub) SendSpeed(speed int) error {
	c.speedCallCount++
	c.speed = speed
	return nil
}

type forwarderStub struct {
	telemetry *Telemetry
}

func (fwd* forwarderStub) Forward(newTelemetry *Telemetry, prevTelemetry *Telemetry) error {
	fwd.telemetry = newTelemetry
	return nil
}