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
	assert.Equal(t, 0.0000001, jc.Telemetry.Latitude)
	assert.Equal(t, 0.0000002, jc.Telemetry.Longitude)
	assert.Equal(t, float32(0.03), jc.Telemetry.Altitude)
	assert.Equal(t, float32(4.0), jc.Telemetry.Track)
	assert.Equal(t, float32(5.0), jc.Telemetry.GPSSpeed)

	// send the same data
	jc.gpsChan <- gps
	prevTelem := jc.Telemetry
	assert.False(t, jc.CheckChannels())
	assert.Equal(t, prevTelem, jc.Telemetry)

	// send different data
	jc.gpsChan <- gpsData{
		Latitude:  6,
		Longitude: 7,
		Altitude:  8,
		Track:     9,
		Speed:     10,
	}
	assert.True(t, jc.CheckChannels())
	assert.Equal(t, 0.0000006, jc.Telemetry.Latitude)
	assert.Equal(t, 0.0000007, jc.Telemetry.Longitude)
	assert.Equal(t, float32(0.08), jc.Telemetry.Altitude)
	assert.Equal(t, float32(9.0), jc.Telemetry.Track)
	assert.Equal(t, float32(10.0), jc.Telemetry.GPSSpeed)
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
	assert.Equal(t, uint8(1), jc.Telemetry.GasPedalAngle)
	assert.Equal(t, float32(2), jc.Telemetry.RPM)
	assert.Equal(t, float32(3), jc.Telemetry.OilPressure)
	assert.Equal(t, float32(4), jc.Telemetry.Speed)
	assert.Equal(t, float32(0), jc.Telemetry.CoolantTemp, "coolant temp should not be from ECU")
	assert.Equal(t, float32(6), jc.Telemetry.AirIntakeTemp)
	assert.Equal(t, float32(7), jc.Telemetry.BatteryVoltage)

	// send the same data
	jc.ecuChan <- ecu
	prevTelem := jc.Telemetry
	assert.False(t, jc.CheckChannels())
	assert.Equal(t, prevTelem, jc.Telemetry)

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
	assert.Equal(t, uint8(8), jc.Telemetry.GasPedalAngle)
	assert.Equal(t, float32(9), jc.Telemetry.RPM)
	assert.Equal(t, float32(10), jc.Telemetry.OilPressure)
	assert.Equal(t, float32(11), jc.Telemetry.Speed)
	assert.Equal(t, float32(0), jc.Telemetry.CoolantTemp, "coolant temp should not be from ECU")
	assert.Equal(t, float32(13), jc.Telemetry.AirIntakeTemp)
	assert.Equal(t, float32(14), jc.Telemetry.BatteryVoltage)
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
	assert.Equal(t, float32(1), jc.Telemetry.FuelRemaining)
	assert.Equal(t, uint8(2), jc.Telemetry.FuelLevel)
	assert.Equal(t, float32(3), jc.Telemetry.CoolantTemp)
	assert.Equal(t, float32(4), jc.Telemetry.OilTemp)

	jc.canSensorChan <- can
	prevTelem := jc.Telemetry
	assert.False(t, jc.CheckChannels())
	assert.Equal(t, prevTelem, jc.Telemetry)

	jc.canSensorChan <- canSensorData{
		FuelRemaining: 5,
		FuelLevel:     6,
		CoolantTemp:   7,
		OilTemp:       8,
	}
	assert.True(t, jc.CheckChannels())
	assert.Equal(t, float32(5), jc.Telemetry.FuelRemaining)
	assert.Equal(t, uint8(6), jc.Telemetry.FuelLevel)
	assert.Equal(t, float32(7), jc.Telemetry.CoolantTemp)
	assert.Equal(t, float32(8), jc.Telemetry.OilTemp)
}

func TestInterleaved(t *testing.T) {
	jc := NewJuicer()

	jc.canSensorChan <- canSensorData{
		FuelRemaining: 1,
	}
	assert.True(t, jc.CheckChannels())
	assert.Equal(t, float32(1), jc.Telemetry.FuelRemaining)

	jc.ecuChan <- ecuData{
		GasPedalAngle:  8,
	}
	assert.True(t, jc.CheckChannels())
	assert.Equal(t, float32(1), jc.Telemetry.FuelRemaining)
	assert.Equal(t, uint8(8), jc.Telemetry.GasPedalAngle)

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