package main

import (
	"context"
	"github.com/jd3nn1s/juicer"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetLevel(log.InfoLevel)

	ctx := context.Background()

	jc := juicer.NewJuicer()
	jc.Start(ctx)

	for {
		changed := jc.CheckChannels()
		if changed {
			jc.TelemetryUpdate()
		}
	}
}
