package juicer

import (
	"context"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func TestRunCANBus(t *testing.T) {
	jc := NewJuicer()

	origCanBusConnect := canBusConnect
	defer func() {
		canBusConnect = origCanBusConnect
	}()

	stub := createCANBusStub()
	canBusConnect = func(p string) (CANBus, error) {
		return stub, nil
	}

	canBusRetryable := &canBusRetryable{
		sendChan: jc.canSensorChan,
	}

	// close before opening
	assert.NoError(t, canBusRetryable.Close())
	assert.NoError(t, canBusRetryable.Open())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		_ = canBusRetryable.Start(ctx)
		wg.Done()
	}()
	<-stub.startChan

	expectedData := canSensorData{}

	stub.fnChan <- func() {
		stub.callbacks.CoolantTemp(1)
	}
	data := <-jc.canSensorChan
	expectedData.CoolantTemp = 1
	assert.Equal(t, expectedData, data)

	stub.fnChan <- func() {
		stub.callbacks.OilTemp(2)
	}
	data = <-jc.canSensorChan
	expectedData.OilTemp = 2
	assert.Equal(t, expectedData, data)

	stub.fnChan <- func() {
		stub.callbacks.Fuel(3)
	}
	data = <-jc.canSensorChan
	expectedData.FuelLevel = 3
	assert.Equal(t, expectedData, data)

	cancel()
	wg.Wait()
}