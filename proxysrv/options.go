package proxysrv

import (
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stripe/veneur/trace"
)

func WithForwardTimeout(d time.Duration) Option {
	return func(opts *options) {
		opts.forwardTimeout = d
	}
}

func WithLog(e *logrus.Entry) Option {
	return func(opts *options) {
		opts.log = e
	}
}

func WithTraceClient(c *trace.Client) Option {
	return func(opts *options) {
		opts.traceClient = c
	}
}
