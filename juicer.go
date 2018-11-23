package main

import (
	"context"
	"github.com/jd3nn1s/juicer/lemoncan"
	"github.com/jd3nn1s/skytraq"
	log "github.com/sirupsen/logrus"
	"math"
	"time"
)

const (
	ecuPortName    = "/dev/obd"
	gpsPortName    = "/dev/ttyAMA0"
	canBusPortName = "can0"

	channelBufferSize = 1
)

var gpsConnect = func(p string) (GPS, error) {
	return skytraq.Connect(p)
}
var canBusConnect = func(p string) (CANBus, error) {
	return lemoncan.Connect(p)
}

func main() {
	log.SetLevel(log.InfoLevel)

	ctx := context.Background()

	gpsChan, ecuChan, canSensorChan := mkChannels()
	go runCAN(ctx, canSensorChan)
	go runECU(ctx, ecuChan)
	go runGPS(ctx, gpsChan)

	telemetry := Telemetry{}
	for {
		changed := checkChannels(&telemetry, gpsChan, ecuChan, canSensorChan)
		if changed {
			// send to UDP channel
		}
	}
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

func runGPS(ctx context.Context, sendChan chan<- gpsData) {
	var err error
	var c GPS
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		if err != nil {
			log.Error("reconnecting due to error ", err)
			if c != nil {
				if err = c.Close(); err != nil {
					log.WithField("err", err).Warn("unable to close gps connection")
				}
			}
			c = nil
			time.Sleep(time.Second)
		}
		if c == nil {
			c, err = gpsConnect(gpsPortName)
			if err != nil {
				continue
			}
		}
		err = c.Start(ctx, skytraq.Callbacks{
			SoftwareVersion: func(version skytraq.SoftwareVersion) {
				log.Infof("software version: %v", version)
			},
			NavData: func(navData skytraq.NavData) {
				if navData.Fix == skytraq.FixNone {
					log.Warnf("no satellite fix")
					return
				}
				if navData.HDOP > 500 {
					log.WithField("HDOP", navData.HDOP).Warn("poor resolution")
				}
				speed := math.Sqrt(math.Pow(float64(navData.VX), 2) +
					math.Pow(float64(navData.VY), 2))
				track := math.Atan(float64(navData.VX) / float64(navData.VY))
				if math.IsNaN(track) {
					track = 0
				}

				select {
				case sendChan <- gpsData{
					Latitude:  navData.Latitude,
					Longitude: navData.Longitude,
					Altitude:  navData.Altitude,
					Speed:     speed,
					Track:     track,
				}:
				default:
				}
			},
		})
	}
}

func runCAN(ctx context.Context, sendChan chan<- canSensorData) {
	var err error
	var c CANBus
	var data canSensorData

	send := func() {
		select {
		case sendChan <- data:
		default:
		}
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		if err != nil {
			log.Error("reconnecting due to error ", err)
			if c != nil {
				if err = c.Close(); err != nil {
					log.WithField("err", err).Warn("unable to close canbus connection")
				}
			}
			c = nil
			time.Sleep(time.Second)
		}
		if c == nil {
			c, err = canBusConnect(canBusPortName)
			continue
		}
		err = c.Start(ctx, lemoncan.Callbacks{
			Fuel: func(v int) {
				data.FuelLevel = v
				send()
			},
			CoolantTemp: func(v int) {
				data.CoolantTemp = v
				send()
			},
			OilTemp: func(v int) {
				data.OilTemp = v
				send()
			},
		})
	}
}
