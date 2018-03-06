package proxyserver

import (
	"context"
	"fmt"
	"io/ioutil"
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"
	"github.com/stripe/veneur/forwardrpc"
	"github.com/stripe/veneur/samplers"
	"github.com/stripe/veneur/samplers/metricpb"

	"github.com/stripe/veneur/ssf"
	"github.com/stripe/veneur/trace"
	"google.golang.org/grpc"
	"stathat.com/c/consistent"
)

type Server struct {
	destinations *consistent.Consistent
	opts         *Options
}

type Options struct {
	Log         *logrus.Entry
	Timeout     time.Duration
	TraceClient *trace.Client
}

func (o *Options) setDefaults() {
	if o.Log == nil {
		log := logrus.New()
		log.Out = ioutil.Discard
		o.Log = logrus.NewEntry(log)
	}
}

func New(destinations *consistent.Consistent, opts *Options) *Server {
	if opts == nil {
		opts = &Options{}
	}
	opts.setDefaults()

	return &Server{
		destinations: destinations,
		opts:         opts,
	}
}

func (s *Server) SendMetrics(ctx context.Context, mlist *forwardrpc.MetricList) (*empty.Empty, error) {
	go s.sendMetrics(context.Background(), mlist)
	return &empty.Empty{}, nil
}

func (s *Server) sendMetrics(ctx context.Context, mlist *forwardrpc.MetricList) (res error) {
	span, _ := trace.StartSpanFromContext(ctx, "veneur.opentracing.proxy.forward_metrics")
	defer span.ClientFinish(s.opts.TraceClient)

	if s.opts.Timeout > 0 {
		s.opts.Log.WithField("timeout", s.opts.Timeout).Info("Setting timeout")
		var cancel func()
		ctx, cancel = context.WithTimeout(ctx, s.opts.Timeout)
		defer cancel()
	}
	metrics := mlist.Metrics
	span.Add(ssf.Count("import.metrics_total", float32(len(metrics)), map[string]string{
		"veneurglobalonly": "",
		"input":            "grpc",
	}))

	dests := make(map[string][]*metricpb.Metric)
	for _, metric := range metrics {
		dest, err := s.destForMetric(metric)
		if err != nil {
			res = multierror.Append(res, s.recordError(span, err, "no-destination",
				"failed to get a destination for a metric", 1))
		} else {
			if _, ok := dests[dest]; !ok {
				dests[dest] = make([]*metricpb.Metric, 0, 1)
			}
			dests[dest] = append(dests[dest], metric)
		}
	}

	// Wait for all of the forward to finish
	wg := sync.WaitGroup{}
	wg.Add(len(dests))

	for dest, batch := range dests {
		go func(dest string, batch []*metricpb.Metric) {
			defer wg.Done()
			if err := s.forward(ctx, dest, batch); err != nil {
				res = multierror.Append(res, s.recordError(span, err, "forward",
					fmt.Sprintf("failed to forward %d metrics to the host '%s'",
						len(batch), dest),
					len(batch)))
			}
		}(dest, batch)
	}

	wg.Wait() // Wait for all the above goroutines to complete

	span.Add(ssf.Timing("proxy.duration_ns", time.Since(span.Start), time.Nanosecond, map[string]string{
		"protocol": "grpc",
	}))
	span.Add(ssf.Count("proxy.proxied_metrics_total", float32(len(metrics)), nil))

	return res
}

func (s *Server) recordError(
	span *trace.Span,
	err error,
	cause string,
	message string,
	numMetrics int,
) error {
	tags := map[string]string{
		"cause":    cause,
		"protocol": "grpc",
	}
	span.Add(ssf.Count("proxy.proxied_metrics_failed", float32(numMetrics), tags))
	span.Add(ssf.Count("proxy.forward_errors", 1, tags))
	s.opts.Log.WithError(err).WithFields(logrus.Fields{
		"cause": cause,
	}).Error(message)

	return fmt.Errorf("%s: %v", message, err)
}

func (s *Server) destForMetric(m *metricpb.Metric) (string, error) {
	key := samplers.NewMetricKeyFromMetric(m)
	dest, err := s.destinations.Get(key.String())
	if err != nil {
		return "", fmt.Errorf("failed to hash the MetricKey '%s' to a "+
			"destination: %v", key.String(), err)
	}

	return dest, nil
}

func (s *Server) forward(ctx context.Context, dest string, metrics []*metricpb.Metric) error {
	conn, err := grpc.Dial(dest, grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("failed to create a gRPC connection: %v", err)
	}
	defer conn.Close()

	c := forwardrpc.NewForwardClient(conn)
	_, err = c.SendMetrics(ctx, &forwardrpc.MetricList{Metrics: metrics})
	if err != nil {
		return fmt.Errorf("failed to send %d metrics over gRPC: %v",
			len(metrics), err)
	}

	return nil
}
