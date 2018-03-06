package importsrv

import "github.com/stripe/veneur/trace"

func WithTraceClient(c *trace.Client) Option {
	return func(opts *options) {
		opts.traceClient = c
	}
}
