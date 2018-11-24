package juicer

import (
	"github.com/pkg/errors"
)

type CANForwarder struct {
	canSensorBus *canBusRetryable
}

func (fwd *CANForwarder) Forward(prevTelemetry *Telemetry, newTelemetry *Telemetry) error {
	if prevTelemetry.Speed != newTelemetry.Speed {
		canBus := fwd.canSensorBus.CANBus()
		if canBus == nil {
			return errors.New("canbus is not initialized")
		}
		if err := canBus.SendSpeed(int(newTelemetry.Speed)); err != nil {
			return errors.Wrapf(err, "unable to send speed to CAN bus")
		}
	}
	return nil
}
