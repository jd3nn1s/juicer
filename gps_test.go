package main

import (
	"context"
	"github.com/jd3nn1s/skytraq"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func TestRunGPS(t *testing.T) {
	gpsChan, _, _ := mkChannels()

	origGPSConnect := gpsConnect
	defer func() {
		gpsConnect = origGPSConnect
	}()

	stub := createGPSStub()
	gpsConnect = func(p string) (GPS, error) {
		return stub, nil
	}

	gpsRetryable := &gpsRetryable{
		sendChan: gpsChan,
	}

	// close before opening
	assert.NoError(t, gpsRetryable.Close())
	assert.NoError(t, gpsRetryable.Open())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		_ = gpsRetryable.Start(ctx)
		wg.Done()
	}()
	<-stub.startChan

	stub.fnChan <- func() {
		stub.callbacks.SoftwareVersion(skytraq.SoftwareVersion{
			Kernel:   skytraq.Version{1, 2, 3},
			ODM:      skytraq.Version{4, 5, 6},
			Revision: skytraq.Version{7, 8, 9},
		})
	}

	navData := skytraq.NavData{
		Fix:            skytraq.Fix3D,
		SatelliteCount: 1,
		Latitude:       2,
		Longitude:      3,
		Altitude:       4,
		VX:             5,
		VY:             7,
		VZ:             8,
		HDOP:           9,
	}

	navData.Fix = skytraq.Fix3D
	stub.fnChan <- func() {
		stub.callbacks.NavData(navData)
	}

	// read some data
	<-gpsChan

	cancel()
	wg.Wait()
}

func TestNavDataFn(t *testing.T) {
	gpsChan, _, _ := mkChannels()
	gpsRetryable := gpsRetryable{
		sendChan: gpsChan,
	}

	navData := skytraq.NavData{
		Fix:            skytraq.FixNone,
		SatelliteCount: 1,
		Latitude:       2,
		Longitude:      3,
		Altitude:       4,
		VX:             5,
		VY:             7,
		VZ:             8,
		HDOP:           9,
	}

	gpsRetryable.navDataFn(navData)
	assertNoData(t, gpsChan,"unexpected data on channel as there is no fix")

	navData.Fix = skytraq.Fix3D
	gpsRetryable.navDataFn(navData)
	telem := <-gpsChan
	assert.Equal(t, 2, telem.Latitude)
	assert.Equal(t, 3, telem.Longitude)
	assert.Equal(t, 4, telem.Altitude)
	assert.Equal(t, float64(8.602325267042627), telem.Speed)
	assert.Equal(t, float64(0.6202494859828215), telem.Track)

	navData.HDOP = maxHDOP + 1
	gpsRetryable.navDataFn(navData)
	assertNoData(t, gpsChan,"unexpected data on channel as there is high HDOP")

	// no VY or VX should return 0 track
	navData.HDOP = 0
	navData.VY = 0
	navData.VX = 0
	gpsRetryable.navDataFn(navData)
	telem = <-gpsChan
	assert.Equal(t, float64(0), telem.Track)
}

func assertNoData(t *testing.T, gpsChan <-chan gpsData, msg string) {
	select {
	case <-gpsChan:
		assert.Fail(t, msg)
	default:
	}
}