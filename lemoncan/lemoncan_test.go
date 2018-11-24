package lemoncan

import (
	"context"
	"encoding/binary"
	"github.com/brutella/can"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

type busStub struct {
	disconnected bool
	subscribed   bool
	stopChan     chan struct{}
	startedChan  chan struct{}
	publishChan chan *can.Frame
}

func (bus *busStub) SubscribeFunc(can.HandlerFunc) {
	bus.subscribed = true
}

func (bus *busStub) ConnectAndPublish() error {
	bus.startedChan <- struct{}{}
	<-bus.stopChan
	return nil
}

func (bus *busStub) Disconnect() error {
	bus.disconnected = true
	bus.stopChan <- struct{}{}
	return nil
}

func (bus *busStub) Publish(f can.Frame) error {
	bus.publishChan <- &f
	return nil
}

func TestConnect(t *testing.T) {
	origNewBus := newBus
	bus := &busStub{
		stopChan: make(chan struct{}, 1),
	}
	newBus = func(string) (CANBus, error) {
		return bus, nil
	}
	defer func() {
		newBus = origNewBus
	}()

	c, err := Connect("fakeport")
	assert.NotNil(t, c)
	assert.NoError(t, err)
	assert.IsType(t, &busStub{}, c.bus)

	assert.NoError(t, c.Close())
	assert.True(t, bus.disconnected)
}

func TestStart(t *testing.T) {
	bus := &busStub{
		stopChan:    make(chan struct{}),
		startedChan: make(chan struct{}),
	}

	c := &Connection{
		bus: bus,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cb := Callbacks{}
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		assert.NoError(t, c.Start(ctx, cb))
		wg.Done()
	}()
	<-bus.startedChan
	assert.True(t, bus.subscribed)
	assert.NotNil(t, c.cb)
	cancel()
	wg.Wait()
}

func TestSendSpeed(t *testing.T) {
	bus := &busStub{
		publishChan: make(chan *can.Frame, 1),
	}

	c := &Connection{
		bus: bus,
	}

	assert.NoError(t, c.SendSpeed(100))
	f := <-bus.publishChan
	assert.Equal(t, uint32(frameSpeed), f.ID)
}

func TestHandleFrame(t *testing.T) {
	bus := &busStub{
		publishChan: make(chan *can.Frame, 1),
	}

	data := struct{
		OilTemp int
		CoolantTemp int
		Fuel int
	}{}

	c := &Connection{
		cb: &Callbacks{
			OilTemp:     func(v int) {
				data.OilTemp = v
			},
			CoolantTemp: func(v int) {
				data.CoolantTemp = v
			},
			Fuel:        func(v int) {
				data.Fuel = v
			},
		},
		bus: bus,
	}
	expectedData := data

	buf := [8]byte{}
	binary.LittleEndian.PutUint16(buf[0:2], 1)
	c.handleFrame(can.Frame{
		ID: frameOilTemp,
		Length: 2,
		Data: buf,
	})
	expectedData.OilTemp = 1
	assert.Equal(t, expectedData, data)

	binary.LittleEndian.PutUint16(buf[0:2], 2)
	c.handleFrame(can.Frame{
		ID: frameCoolantTemp,
		Length: 2,
		Data: buf,
	})
	expectedData.CoolantTemp = 2
	assert.Equal(t, expectedData, data)

	binary.LittleEndian.PutUint16(buf[0:2], 3)
	c.handleFrame(can.Frame{
		ID: frameFuel,
		Length: 2,
		Data: buf,
	})
	expectedData.Fuel = 3
	assert.Equal(t, expectedData, data)

	// send unknown CAN frame
	c.handleFrame(can.Frame{
		ID: 400,
	})
	// no change to data
	assert.Equal(t, expectedData, data)

	// send too short a frame
	c.handleFrame(can.Frame{
		ID: frameOilTemp,
	})
	// no change to data
	assert.Equal(t, expectedData, data)
}

func TestUint16Result(t *testing.T) {
	_, err := uint16Result(can.Frame{})
	assert.Error(t, err)
	_, err = uint16Result(can.Frame{
		Length: 3,
	})
	assert.Error(t, err)

	buf := [8]byte{}
	binary.LittleEndian.PutUint16(buf[0:2], 300)
	n, err := uint16Result(can.Frame{
		Length: 2,
		Data: buf,
	})
	assert.NoError(t, err)
	assert.Equal(t, 300, n)
}