package juicer

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
	PrevTelemetry Telemetry
	Telemetry     Telemetry

	gpsChan       chan gpsData
	ecuChan       chan ecuData
	canSensorChan chan canSensorData

	canSensorBus *canBusRetryable

	forwarders []Forwarder
	testMode   bool
}

func NewJuicer() *Juicer {
	jc := &Juicer{
		forwarders: make([]Forwarder, 0),
	}
	jc.mkChannels()

	canSensorBus := newCANBus(jc.canSensorChan)
	jc.canSensorBus = canSensorBus

	jc.AddForwarder(&CANForwarder{
		canSensorBus: canSensorBus,
	})
	return jc
}

func (jc *Juicer) AddForwarder(fwder Forwarder) {
	jc.forwarders = append(jc.forwarders, fwder)
}

func (jc *Juicer) Start(ctx context.Context) {
	if jc.testMode {
		log.Warn("starting in test mode")
		jc.runTestMode(ctx)
		return
	}
	go jc.canSensorBus.runCAN(ctx)
	go runECU(ctx, jc.ecuChan)
	go runGPS(ctx, jc.gpsChan)
}

func (jc *Juicer) SetTestMode(testMode bool) {
	jc.testMode = testMode
}

func (jc *Juicer) TelemetryUpdate() {
	for _, fwder := range jc.forwarders {
		if err := fwder.Forward(&jc.PrevTelemetry, &jc.Telemetry); err != nil {
			log.Errorf("unable to send to forwarder %v %v", fwder, err)
		}
	}
}

func (jc *Juicer) mkChannels() {
	jc.gpsChan = make(chan gpsData, channelBufferSize)
	jc.ecuChan = make(chan ecuData, channelBufferSize)
	jc.canSensorChan = make(chan canSensorData, channelBufferSize)
}

func (jc *Juicer) CheckChannels() (changed bool) {
	newTelemetry := jc.Telemetry
	select {
	case gpsData := <-jc.gpsChan:
		newTelemetry.Latitude = float64(gpsData.Latitude) / math.Pow(10, 7)
		newTelemetry.Longitude = float64(gpsData.Longitude) / math.Pow(10, 7)
		newTelemetry.Altitude = float32(gpsData.Altitude) / 100.0
		newTelemetry.Track = float32(gpsData.Track)
		newTelemetry.GPSSpeed = float32(gpsData.Speed)
	case ecuData := <-jc.ecuChan:
		newTelemetry.GasPedalAngle = uint8(ecuData.GasPedalAngle)
		newTelemetry.RPM = ecuData.RPM
		newTelemetry.OilPressure = ecuData.OilPressure
		newTelemetry.Speed = float32(ecuData.Speed)
		newTelemetry.AirIntakeTemp = ecuData.AirIntakeTemp
		newTelemetry.BatteryVoltage = ecuData.BatteryVoltage
	case canSensorData := <-jc.canSensorChan:
		newTelemetry.FuelRemaining = canSensorData.FuelRemaining
		newTelemetry.FuelLevel = uint8(canSensorData.FuelLevel)
		newTelemetry.CoolantTemp = float32(canSensorData.CoolantTemp)
		newTelemetry.OilTemp = float32(canSensorData.OilTemp)
	}
	if jc.Telemetry != newTelemetry {
		jc.PrevTelemetry = jc.Telemetry
		jc.Telemetry = newTelemetry
		return true
	}
	return false
}

func castToFloat32(val interface{}) float32 {
	switch v := val.(type) {
	case int:
		return float32(v)
	case float64:
		return float32(v)
	case float32:
		return v
	default:
		log.WithField("v", val).Error("unable to cast to float32")
	}
	return 0
}
