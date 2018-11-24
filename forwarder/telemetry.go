package forwarder

import "github.com/jd3nn1s/juicer"

type Header struct {
	Type uint8
}

const (
	TypeTelemetry = 1
	TypeTiming    = 2
)

type UDPForwarder struct {

}

func (udp *UDPForwarder) Forward(newTelemetry *juicer.Telemetry, prevTelemetry *juicer.Telemetry) error {
	return nil
}