package importsrv

import (
	"context"
	"fmt"
	"hash/fnv"
	"net"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/stripe/veneur/forwardrpc"
	"github.com/stripe/veneur/samplers"
	"github.com/stripe/veneur/samplers/metricpb"
	"github.com/stripe/veneur/ssf"
	"github.com/stripe/veneur/trace"
	"google.golang.org/grpc"
)

// MetricIngester reads metrics from protobufs
type MetricIngester interface {
	IngestMetric(*metricpb.Metric)
}

// Server wraps a gRPC server and implements the forwardrpc.Forward service.
// It reads a list of metrics, and based on the provided key chooses a
// MetricIngester to send it to.  A unique metric (name, tags, and type)
// should always be routed to the same MetricIngester.
type Server struct {
	*grpc.Server
	metricOuts []MetricIngester
	opts       *options
}

type options struct {
	traceClient *trace.Client
}

// Option is returned by functions that serve as options to New, like
// "With..."
type Option func(*options)

// New creates a unstarted Server with the input MetricIngester's to send
// output to.
func New(metricOuts []MetricIngester, opts ...Option) *Server {
	res := &Server{
		Server:     grpc.NewServer(),
		metricOuts: metricOuts,
		opts:       &options{},
	}

	for _, opt := range opts {
		opt(res.opts)
	}

	if res.opts.traceClient == nil {
		res.opts.traceClient = trace.DefaultClient
	}

	forwardrpc.RegisterForwardServer(res.Server, res)

	return res
}

func (s *Server) Serve(addr string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to bind the import server to '%s': %v",
			addr, err)
	}

	return s.Server.Serve(ln)
}

func (s *Server) SendMetrics(ctx context.Context, mlist *forwardrpc.MetricList) (*empty.Empty, error) {
	span, _ := trace.StartSpanFromContext(ctx, "veneur.opentracing.importsrv.handle_send_metrics")
	span.SetTag("protocol", "grpc")
	defer span.ClientFinish(s.opts.traceClient)

	h := fnv.New32a()

	for _, m := range mlist.Metrics {
		h.Reset()
		h.Write([]byte(samplers.NewMetricKeyFromMetric(m).String()))
		workerIdx := h.Sum32() % uint32(len(s.metricOuts))
		s.metricOuts[workerIdx].IngestMetric(m)
	}

	span.Add(
		ssf.Timing("import.response_duration_ns", time.Since(span.Start),
			time.Nanosecond, map[string]string{"part": "merge"}),
		ssf.Count("import.metrics_total", float32(len(mlist.Metrics)), nil),
	)

	return &empty.Empty{}, nil
}
