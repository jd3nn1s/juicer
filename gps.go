package juicer

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

type gpsRetryable struct {
	c GPS
	sendChan chan<- gpsData
}

func (g *gpsRetryable) Open() error {
	c, err := gpsConnect(gpsPortName)
	g.c = c
	return err
}

func (g *gpsRetryable) Close() error {
	if g.c == nil {
		return nil
	}
	return g.c.Close()
}

func (g *gpsRetryable) Start(ctx context.Context) error {
	return g.c.Start(ctx, skytraq.Callbacks{
		SoftwareVersion: func(version skytraq.SoftwareVersion) {
			log.Infof("software version: %v", version)
		},
		NavData: g.navDataFn,
	})
}

func (g *gpsRetryable) Name() string {
	return "gps"
}

func (g*gpsRetryable) navDataFn(navData skytraq.NavData) {
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
	err := retry(ctx, &gpsRetryable{
		sendChan: sendChan,
	})
	if err != nil {
		log.Errorf("gps done: %v", err)
	}
}