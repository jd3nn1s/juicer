package main

import (
	"context"
	"flag"
	"github.com/jd3nn1s/juicer"
	"github.com/jd3nn1s/juicer/forwarder"
	log "github.com/sirupsen/logrus"
)

var testMode = flag.Bool("testmode", false, "generate test data")

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
			jc.TelemetryUpdate()
		}
	}
}
