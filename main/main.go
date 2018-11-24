package main

import (
	"context"
	"github.com/jd3nn1s/juicer"
	"github.com/jd3nn1s/juicer/forwarder"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetLevel(log.InfoLevel)

	ctx := context.Background()

	jc := juicer.NewJuicer()
	fwder, err := forwarder.NewUDPForwarder("udpforwarder.toml")
	if err != nil {
		log.Fatal("unable to load UDP forwarder: ", err)
	}
	jc.AddForwarder(fwder)
	jc.Start(ctx)

	for {
		changed := jc.CheckChannels()
		if changed {
			jc.TelemetryUpdate()
		}
	}
}
