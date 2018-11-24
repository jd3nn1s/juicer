package main

import (
	"context"
	"github.com/jd3nn1s/kw1281"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func TestRunECU(t *testing.T) {
	_, ecuChan, _ := mkChannels()

	origECUConnect := ecuConnect
	defer func() {
		ecuConnect = origECUConnect
	}()

	stub := createECUStub()
	ecuConnect = func(p string) (KW1281, error) {
		return stub, nil
	}

	ecuRetryable := &ecuRetryable{
		sendChan: ecuChan,
	}

	// close before opening
	assert.NoError(t, ecuRetryable.Close())
	assert.NoError(t, ecuRetryable.Open())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		_ = ecuRetryable.Start(ctx)
		wg.Done()
	}()
	<-stub.startChan

	stub.fnChan <- func() {
		stub.callbacks.ECUDetails(&kw1281.ECUDetails{
			PartNumber: "fakeecu",
			Details: []string{
				"line 1",
				"line 2",
			},
		})
	}

	measurements := []*kw1281.Measurement{
		{
			Metric: kw1281.MetricRPM,
			MeasurementValue: &kw1281.MeasurementValue{
				Value: 3200,
				Units: "RPM",
			},
		}, {
			Metric: kw1281.MetricCoolantTemp,
			MeasurementValue: &kw1281.MeasurementValue{
				Value: 53,
				Units: "Deg",
			},
		}, {
			Metric:           0,
			MeasurementValue: nil,
		},
	}
	stub.fnChan <- func() {
		stub.callbacks.Measurement(kw1281.GroupRPMCoolantTemp, measurements)
	}
	data := <-ecuChan
	assert.Equal(t, float32(3200), data.RPM)
	cancel()
	wg.Wait()
}
