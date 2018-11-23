package main

import (
	"context"
	"github.com/jd3nn1s/skytraq"
	log "github.com/sirupsen/logrus"
	"math"
)

const (
	// maximum horizontal dilution of precision
	maxHDOP = 500
)

type gps struct {
	c GPS
	sendChan chan<- gpsData
}

func (g *gps) Open() error {
	c, err := gpsConnect(gpsPortName)
	g.c = c
	return err
}

func (g *gps) Close() error {
	if g.c == nil {
		return nil
	}
	return g.c.Close()
}

func (g *gps) Start(ctx context.Context) error {
	return g.c.Start(ctx, skytraq.Callbacks{
		SoftwareVersion: func(version skytraq.SoftwareVersion) {
			log.Infof("software version: %v", version)
		},
		NavData: g.navDataFn,
	})
}

func (g *gps) Name() string {
	return "gps"
}

func (g* gps) navDataFn(navData skytraq.NavData) {
	if navData.Fix == skytraq.FixNone {
		log.Warnf("no satellite fix")
		return
	}
	if navData.HDOP > maxHDOP {
		log.WithField("HDOP", navData.HDOP).Warn("poor resolution")
		return
	}
	speed := math.Sqrt(math.Pow(float64(navData.VX), 2) +
		math.Pow(float64(navData.VY), 2))
	track := math.Atan(float64(navData.VX) / float64(navData.VY))
	if math.IsNaN(track) {
		track = 0
	}

	select {
	case g.sendChan <- gpsData{
		Latitude:  navData.Latitude,
		Longitude: navData.Longitude,
		Altitude:  navData.Altitude,
		Speed:     speed,
		Track:     track,
	}:
	default:
	}
}

var gpsConnect = func(p string) (GPS, error) {
	return skytraq.Connect(p)
}

func runGPS(ctx context.Context, sendChan chan<- gpsData) {
	err := retry(ctx, &gps{
		sendChan: sendChan,
	})
	if err != nil {
		log.Errorf("gps done: %v", err)
	}
}