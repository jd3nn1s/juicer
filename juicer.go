package main

import (
	"context"
	log "github.com/sirupsen/logrus"
	"math"
)

const (
	ecuPortName    = "/dev/obd"
	gpsPortName    = "/dev/ttyAMA0"
	canBusPortName = "can0"

	channelBufferSize = 1
)

type Juicer struct {
	metricSender MetricSender
	prevTelemetry Telemetry
	telemetry Telemetry
}

func main() {
	log.SetLevel(log.InfoLevel)

	ctx := context.Background()

	gpsChan, ecuChan, canSensorChan := mkChannels()
	canSensorBus := newCANBus(canSensorChan)
	go canSensorBus.runCAN(ctx)
	go runECU(ctx, ecuChan)
	go runGPS(ctx, gpsChan)

	juicer := Juicer{
		metricSender: canSensorBus.CANBus(),
	}

	for {
		changed := checkChannels(&juicer.telemetry, gpsChan, ecuChan, canSensorChan)
		if changed {
			juicer.telemetryUpdate()
		}
	}
}

func (jc *Juicer) telemetryUpdate() {
	// send to UDP channel
	if jc.prevTelemetry.Speed != jc.telemetry.Speed {
		if err := jc.metricSender.SendSpeed(int(jc.telemetry.Speed)); err != nil {
			log.WithField("speed", jc.telemetry.Speed).Error("unable to send speed to CAN bus")
		}
	}
	jc.prevTelemetry = jc.telemetry
}

func mkChannels() (gpsChan chan gpsData, ecuChan chan ecuData, canSensorChan chan canSensorData) {
	gpsChan = make(chan gpsData, channelBufferSize)
	ecuChan = make(chan ecuData, channelBufferSize)
	canSensorChan = make(chan canSensorData, channelBufferSize)
	return
}

func checkChannels(curTelemetry *Telemetry, gpsChan chan gpsData, ecuChan chan ecuData, canSensorChan chan canSensorData) (changed bool) {
	newTelemetry := Telemetry{}
	select {
	case gpsData := <-gpsChan:
		newTelemetry.Latitude = float64(gpsData.Latitude) / math.Pow(10, 7)
		newTelemetry.Longitude = float64(gpsData.Longitude) / math.Pow(10, 7)
		newTelemetry.Altitude = float32(gpsData.Altitude) / 100.0
		newTelemetry.Track = float32(gpsData.Track)
		newTelemetry.GPSSpeed = float32(gpsData.Speed)
	case ecuData := <-ecuChan:
		newTelemetry.GasPedalAngle = uint8(ecuData.GasPedalAngle)
		newTelemetry.RPM = ecuData.RPM
		newTelemetry.OilPressure = ecuData.OilPressure
		newTelemetry.Speed = float32(ecuData.Speed)
		newTelemetry.CoolantTemp = ecuData.CoolantTemp
		newTelemetry.AirIntakeTemp = ecuData.AirIntakeTemp
		newTelemetry.BatteryVoltage = ecuData.BatteryVoltage
	case canSensorData := <-canSensorChan:
		newTelemetry.FuelRemaining = canSensorData.FuelRemaining
		newTelemetry.FuelLevel = uint8(canSensorData.FuelLevel)
		newTelemetry.CoolantTemp = float32(canSensorData.CoolantTemp)
		newTelemetry.OilTemp = float32(canSensorData.OilTemp)
	}
	if *curTelemetry != newTelemetry {
		*curTelemetry = newTelemetry
		return true
	}
	return false
}

func castToFloat32(val interface{}) float32 {
	switch v := val.(type) {
	case int:
		return float32(v)
	case float32:
		return v
	}
	return 0
}
