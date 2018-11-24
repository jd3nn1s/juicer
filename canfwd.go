package juicer

import (
	"github.com/pkg/errors"
)

type CANForwarder struct {
	canSensorBus CANBus
}

func (fwd *CANForwarder) Forward(prevTelemetry *Telemetry, newTelemetry *Telemetry) error {
	if prevTelemetry.Speed != newTelemetry.Speed {
		if err := fwd.canSensorBus.SendSpeed(int(newTelemetry.Speed)); err != nil {
			return errors.Wrapf(err, "unable to send speed to CAN bus")
		}
	}
	return nil
}
