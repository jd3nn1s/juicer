package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/jd3nn1s/juicer"
	"github.com/jd3nn1s/juicer/forwarder"
	log "github.com/sirupsen/logrus"
)

var testMode = flag.Bool("testmode", false, "generate test data")
var printTelemetry = flag.Bool("print-telemetry", false, "print telemetry to stdout")

func main() {
	log.SetLevel(log.InfoLevel)
	flag.Parse()

	ctx := context.Background()

	jc := juicer.NewJuicer()
	fwder, err := forwarder.NewUDPForwarder("udpforwarder.toml")
	if err != nil {
		log.Fatal("unable to load UDP forwarder: ", err)
	}
	go fwder.Start(ctx)
	jc.AddForwarder(fwder)
	jc.SetTestMode(*testMode)
	jc.Start(ctx)

	for {
		changed := jc.CheckChannels()
		if changed {
			if *printTelemetry {
				fmt.Printf("%+v\n", jc.Telemetry)
			}
			jc.TelemetryUpdate()
		}
	}
}
