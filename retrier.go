package main

import (
	"context"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"time"
)

var retrySleep = time.Second

type Retryable interface {
	Open() error
	Close() error
	Start(ctx context.Context) error
	Name() string
}

func retry(ctx context.Context, r Retryable) error {
	errStarting := errors.New("starting")
	err := errStarting
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if err != nil {
			if err != errStarting {
				log.WithField("err", err).Errorf("%s: reconnecting due to error", r.Name())
				if err = r.Close(); err != nil {
					log.WithField("err", err).Warnf("%s: unable to close", r.Name())
				}
				time.Sleep(retrySleep)
			}
			err = r.Open()
			if err != nil {
				continue
			}
		}
		err = r.Start(ctx)
	}
}