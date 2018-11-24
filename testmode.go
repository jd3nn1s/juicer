package juicer

import (
	"context"
	"time"
)

func (jc *Juicer) runTestMode(ctx context.Context) {
	gps := gpsData{}
	ecu := ecuData{}
	can := canSensorData{}

	go func() {
		down := false
		for {
			select {
			case <-time.Tick(time.Millisecond * 20):
			case <-ctx.Done():
				return
			}
			jc.gpsChan <- gps

			if down {
				gps.Speed -= 0.01
				gps.Longitude -= 100
				gps.Latitude -= 100
			} else {
				gps.Speed += 0.01
				gps.Longitude += 100
				gps.Latitude += 100
			}

			if gps.Speed == 0 {
				down = false
			} else if gps.Speed == 100 {
				down = true
			}
		}
	}()

	go func() {
		down := false
		for {
			select {
			case <-time.Tick(time.Millisecond * 250):
			case <-ctx.Done():
				return
			}
			jc.ecuChan <- ecu

			if down {
				ecu.RPM -= 100
			} else {
				ecu.RPM += 100
			}

			if ecu.RPM == 1800 {
				down = true
			} else if ecu.RPM == 0 {
				down = false
			}
		}
	}()

	go func() {
		down := false
		for {
			select {
			case <-time.Tick(time.Second):
			case <-ctx.Done():
				return
			}
			jc.canSensorChan <- can

			if down {
				can.CoolantTemp -= 5
				can.FuelLevel -= 1
			} else {
				can.CoolantTemp += 5
				can.FuelLevel += 1
			}

			if can.CoolantTemp == 120 {
				down = true
			} else if can.CoolantTemp == 0 {
				down = false
			}
		}
	}()
}
