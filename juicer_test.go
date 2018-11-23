package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCheckChannelsGPS(t *testing.T) {
	gpsChan, ecuChan, canSensorChan := mkChannels()

	gps := gpsData{
		Latitude:  1,
		Longitude: 2,
		Altitude:  3,
		Track:     4,
		Speed:     5,
	}

	gpsChan <- gps
	telem := Telemetry{}
	assert.True(t, checkChannels(&telem, gpsChan, ecuChan, canSensorChan))
	assert.Equal(t, 0.0000001, telem.Latitude)
	assert.Equal(t, 0.0000002, telem.Longitude)
	assert.Equal(t, float32(0.03), telem.Altitude)
	assert.Equal(t, float32(4.0), telem.Track)
	assert.Equal(t, float32(5.0), telem.GPSSpeed)

	// send the same data
	gpsChan <- gps
	prevTelem := telem
	assert.False(t, checkChannels(&telem, gpsChan, ecuChan, canSensorChan))
	assert.Equal(t, prevTelem, telem)

	// send different data
	gpsChan <- gpsData{
		Latitude:  6,
		Longitude: 7,
		Altitude:  8,
		Track:     9,
		Speed:     10,
	}
	assert.True(t, checkChannels(&telem, gpsChan, ecuChan, canSensorChan))
	assert.Equal(t, 0.0000006, telem.Latitude)
	assert.Equal(t, 0.0000007, telem.Longitude)
	assert.Equal(t, float32(0.08), telem.Altitude)
	assert.Equal(t, float32(9.0), telem.Track)
	assert.Equal(t, float32(10.0), telem.GPSSpeed)
}

func TestCheckChannelECU(t *testing.T) {
	gpsChan, ecuChan, canSensorChan := mkChannels()

	ecu := ecuData{
		GasPedalAngle:  1,
		RPM:            2,
		OilPressure:    3,
		Speed:          4,
		CoolantTemp:    5,
		AirIntakeTemp:  6,
		BatteryVoltage: 7,
	}

	ecuChan <- ecu
	telem := Telemetry{}
	assert.True(t, checkChannels(&telem, gpsChan, ecuChan, canSensorChan))
	assert.Equal(t, uint8(1), telem.GasPedalAngle)
	assert.Equal(t, float32(2), telem.RPM)
	assert.Equal(t, float32(3), telem.OilPressure)
	assert.Equal(t, float32(4), telem.Speed)
	assert.Equal(t, float32(5), telem.CoolantTemp)
	assert.Equal(t, float32(6), telem.AirIntakeTemp)
	assert.Equal(t, float32(7), telem.BatteryVoltage)

	// send the same data
	ecuChan <- ecu
	prevTelem := telem
	assert.False(t, checkChannels(&telem, gpsChan, ecuChan, canSensorChan))
	assert.Equal(t, prevTelem, telem)

	ecuChan <- ecuData{
		GasPedalAngle:  8,
		RPM:            9,
		OilPressure:    10,
		Speed:          11,
		CoolantTemp:    12,
		AirIntakeTemp:  13,
		BatteryVoltage: 14,
	}
	assert.True(t, checkChannels(&telem, gpsChan, ecuChan, canSensorChan))
	assert.Equal(t, uint8(8), telem.GasPedalAngle)
	assert.Equal(t, float32(9), telem.RPM)
	assert.Equal(t, float32(10), telem.OilPressure)
	assert.Equal(t, float32(11), telem.Speed)
	assert.Equal(t, float32(12), telem.CoolantTemp)
	assert.Equal(t, float32(13), telem.AirIntakeTemp)
	assert.Equal(t, float32(14), telem.BatteryVoltage)
}

func TestCheckChannelCAN(t *testing.T) {
	gpsChan, ecuChan, canSensorChan := mkChannels()

	can := canSensorData{
		FuelRemaining: 1,
		FuelLevel:     2,
		CoolantTemp:   3,
		OilTemp:       4,
	}
	canSensorChan <- can
	telem := Telemetry{}
	assert.True(t, checkChannels(&telem, gpsChan, ecuChan, canSensorChan))
	assert.Equal(t, float32(1), telem.FuelRemaining)
	assert.Equal(t, uint8(2), telem.FuelLevel)
	assert.Equal(t, float32(3), telem.CoolantTemp)
	assert.Equal(t, float32(4), telem.OilTemp)

	canSensorChan <- can
	prevTelem := telem
	assert.False(t, checkChannels(&telem, gpsChan, ecuChan, canSensorChan))
	assert.Equal(t, prevTelem, telem)

	canSensorChan <- canSensorData{
		FuelRemaining: 5,
		FuelLevel:     6,
		CoolantTemp:   7,
		OilTemp:       8,
	}
	assert.True(t, checkChannels(&telem, gpsChan, ecuChan, canSensorChan))
	assert.Equal(t, float32(5), telem.FuelRemaining)
	assert.Equal(t, uint8(6), telem.FuelLevel)
	assert.Equal(t, float32(7), telem.CoolantTemp)
	assert.Equal(t, float32(8), telem.OilTemp)
}

func TestCastToFloat32(t *testing.T) {
	assert.Equal(t, float32(1), castToFloat32(int(1)))
	assert.Equal(t, float32(0), castToFloat32("hah"))
}
