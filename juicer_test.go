package juicer

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCheckChannelsGPS(t *testing.T) {
	jc := NewJuicer()

	gps := gpsData{
		Latitude:  1,
		Longitude: 2,
		Altitude:  3,
		Track:     4,
		Speed:     5,
	}

	jc.gpsChan <- gps
	assert.True(t, jc.CheckChannels())
	assert.Equal(t, 0.0000001, jc.telemetry.Latitude)
	assert.Equal(t, 0.0000002, jc.telemetry.Longitude)
	assert.Equal(t, float32(0.03), jc.telemetry.Altitude)
	assert.Equal(t, float32(4.0), jc.telemetry.Track)
	assert.Equal(t, float32(5.0), jc.telemetry.GPSSpeed)

	// send the same data
	jc.gpsChan <- gps
	prevTelem := jc.telemetry
	assert.False(t, jc.CheckChannels())
	assert.Equal(t, prevTelem, jc.telemetry)

	// send different data
	jc.gpsChan <- gpsData{
		Latitude:  6,
		Longitude: 7,
		Altitude:  8,
		Track:     9,
		Speed:     10,
	}
	assert.True(t, jc.CheckChannels())
	assert.Equal(t, 0.0000006, jc.telemetry.Latitude)
	assert.Equal(t, 0.0000007, jc.telemetry.Longitude)
	assert.Equal(t, float32(0.08), jc.telemetry.Altitude)
	assert.Equal(t, float32(9.0), jc.telemetry.Track)
	assert.Equal(t, float32(10.0), jc.telemetry.GPSSpeed)
}

func TestCheckChannelECU(t *testing.T) {
	jc := NewJuicer()

	ecu := ecuData{
		GasPedalAngle:  1,
		RPM:            2,
		OilPressure:    3,
		Speed:          4,
		CoolantTemp:    5,
		AirIntakeTemp:  6,
		BatteryVoltage: 7,
	}

	jc.ecuChan <- ecu
	assert.True(t, jc.CheckChannels())
	assert.Equal(t, uint8(1), jc.telemetry.GasPedalAngle)
	assert.Equal(t, float32(2), jc.telemetry.RPM)
	assert.Equal(t, float32(3), jc.telemetry.OilPressure)
	assert.Equal(t, float32(4), jc.telemetry.Speed)
	assert.Equal(t, float32(0), jc.telemetry.CoolantTemp, "coolant temp should not be from ECU")
	assert.Equal(t, float32(6), jc.telemetry.AirIntakeTemp)
	assert.Equal(t, float32(7), jc.telemetry.BatteryVoltage)

	// send the same data
	jc.ecuChan <- ecu
	prevTelem := jc.telemetry
	assert.False(t, jc.CheckChannels())
	assert.Equal(t, prevTelem, jc.telemetry)

	jc.ecuChan <- ecuData{
		GasPedalAngle:  8,
		RPM:            9,
		OilPressure:    10,
		Speed:          11,
		CoolantTemp:    12,
		AirIntakeTemp:  13,
		BatteryVoltage: 14,
	}
	assert.True(t, jc.CheckChannels())
	assert.Equal(t, uint8(8), jc.telemetry.GasPedalAngle)
	assert.Equal(t, float32(9), jc.telemetry.RPM)
	assert.Equal(t, float32(10), jc.telemetry.OilPressure)
	assert.Equal(t, float32(11), jc.telemetry.Speed)
	assert.Equal(t, float32(0), jc.telemetry.CoolantTemp, "coolant temp should not be from ECU")
	assert.Equal(t, float32(13), jc.telemetry.AirIntakeTemp)
	assert.Equal(t, float32(14), jc.telemetry.BatteryVoltage)
}

func TestCheckChannelCAN(t *testing.T) {
	jc := NewJuicer()

	can := canSensorData{
		FuelRemaining: 1,
		FuelLevel:     2,
		CoolantTemp:   3,
		OilTemp:       4,
	}
	jc.canSensorChan <- can
	assert.True(t, jc.CheckChannels())
	assert.Equal(t, float32(1), jc.telemetry.FuelRemaining)
	assert.Equal(t, uint8(2), jc.telemetry.FuelLevel)
	assert.Equal(t, float32(3), jc.telemetry.CoolantTemp)
	assert.Equal(t, float32(4), jc.telemetry.OilTemp)

	jc.canSensorChan <- can
	prevTelem := jc.telemetry
	assert.False(t, jc.CheckChannels())
	assert.Equal(t, prevTelem, jc.telemetry)

	jc.canSensorChan <- canSensorData{
		FuelRemaining: 5,
		FuelLevel:     6,
		CoolantTemp:   7,
		OilTemp:       8,
	}
	assert.True(t, jc.CheckChannels())
	assert.Equal(t, float32(5), jc.telemetry.FuelRemaining)
	assert.Equal(t, uint8(6), jc.telemetry.FuelLevel)
	assert.Equal(t, float32(7), jc.telemetry.CoolantTemp)
	assert.Equal(t, float32(8), jc.telemetry.OilTemp)
}

func TestInterleaved(t *testing.T) {
	jc := NewJuicer()

	jc.canSensorChan <- canSensorData{
		FuelRemaining: 1,
	}
	assert.True(t, jc.CheckChannels())
	assert.Equal(t, float32(1), jc.telemetry.FuelRemaining)

	jc.ecuChan <- ecuData{
		GasPedalAngle:  8,
	}
	assert.True(t, jc.CheckChannels())
	assert.Equal(t, float32(1), jc.telemetry.FuelRemaining)
	assert.Equal(t, uint8(8), jc.telemetry.GasPedalAngle)

}

func TestCastToFloat32(t *testing.T) {
	assert.Equal(t, float32(1), castToFloat32(int(1)))
	assert.Equal(t, float32(0), castToFloat32("hah"))
}

func TestSendSpeed(t *testing.T) {
	canStub := canBusStub{}
	fwder := &CANForwarder{
		canSensorBus: &canBusRetryable{
			c:        &canStub,
		},
	}

	prevT := Telemetry{}
	newT := Telemetry{}
	newT.Speed = 100
	assert.NoError(t, fwder.Forward(&prevT, &newT))
	assert.Equal(t, 100, canStub.speed)
	assert.Equal(t, 1, canStub.speedCallCount)

	prevT = newT
	assert.NoError(t, fwder.Forward(&prevT, &newT))
	assert.Equal(t, 1, canStub.speedCallCount, "unexpected call after unchanged telemetry")

	newT.Speed = 200
	assert.NoError(t, fwder.Forward(&prevT, &newT))
	assert.Equal(t, 200, canStub.speed)
	assert.Equal(t, 2, canStub.speedCallCount)
}

func TestAddForwarder(t *testing.T) {
	jc := NewJuicer()
	fwder := forwarderStub{}
	jc.AddForwarder(&fwder)
}