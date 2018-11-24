package juicer

import (
	"context"
	"github.com/jd3nn1s/juicer/lemoncan"
	"github.com/jd3nn1s/kw1281"
	"github.com/jd3nn1s/skytraq"
)

type KW1281 interface {
	Close() error
	Start(context.Context, kw1281.Callbacks) error
}

type GPS interface {
	Close() error
	Start(context.Context, skytraq.Callbacks) error
}

type CANBus interface {
	Close() error
	Start(context.Context, lemoncan.Callbacks) error
	SendSpeed(int) error
}

type MetricSender interface {
	SendSpeed(int) error
}