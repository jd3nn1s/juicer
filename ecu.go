package main

import (
	"context"
	"github.com/jd3nn1s/kw1281"
	log "github.com/sirupsen/logrus"
)

type ecuRetryable struct{
	c KW1281
	sendChan chan<- ecuData
}

// to allow testing
var ecuConnect = func(p string) (KW1281, error) {
	return kw1281.Connect(p)
}

func (e *ecuRetryable) Name() string {
	return "ecu"
}

func (e *ecuRetryable) Open() error {
	c, err := ecuConnect(ecuPortName)
	e.c = c
	return err
}

func (e *ecuRetryable) Close() error {
	if e.c == nil {
		return nil
	}
	return e.c.Close()
}

func (e *ecuRetryable) Start(ctx context.Context) error {
	data := ecuData{}
	return e.c.Start(ctx, kw1281.Callbacks{
		ECUDetails: func(details *kw1281.ECUDetails) {
			log.WithField("partNumber", details.PartNumber).Info()
			for _, line := range details.Details {
				log.Infof("ECU: %s", line)
			}
		},
		Measurement: func(group kw1281.MeasurementGroup, measurements []*kw1281.Measurement) {
			for _, m := range measurements {
				switch m.Metric {
				case kw1281.MetricRPM:
					data.RPM = castToFloat32(m.Value)
				case kw1281.MetricBatteryVoltage:
					data.BatteryVoltage = castToFloat32(m.Value)
				case kw1281.MetricThrottleAngle:
					data.GasPedalAngle = m.Value.(int)
				case kw1281.MetricAirIntakeTemp:
					data.AirIntakeTemp = castToFloat32(m.Value)
				case kw1281.MetricSpeed:
					data.Speed = m.Value.(int)
				}
			}
			select {
			case e.sendChan <- data:
			default:
			}
		},
	})
}

func runECU(ctx context.Context, sendChan chan<- ecuData) {
	err := retry(ctx, &ecuRetryable{
		sendChan: sendChan,
	})
	if err != nil {
		log.Errorf("ecu done: %v", err)
	}
}
