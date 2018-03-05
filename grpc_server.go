package veneur

import (
	"context"
	"hash/fnv"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/stripe/veneur/forwardrpc"
	"github.com/stripe/veneur/samplers"
	"github.com/stripe/veneur/ssf"
	"github.com/stripe/veneur/trace"
)

func (s *Server) SendMetrics(ctx context.Context, mlist *forwardrpc.MetricList) (*empty.Empty, error) {
	span, _ := trace.StartSpanFromContext(ctx, "veneur.opentracing.import.import_metrics")
	span.SetTag("protocol", "grpc")
	defer span.ClientFinish(s.TraceClient)

	// TODO profile how sending each metric to the channel compares to sending
	// chunks.  The chunked method is how the HTTP import is implemented
	for _, m := range mlist.Metrics {
		h := fnv.New32a()
		h.Write([]byte(samplers.NewMetricKeyFromMetric(m).String()))
		i := h.Sum32() % uint32(len(s.Workers))
		s.Workers[i].ImportMetricChan <- m
	}

	span.Add(ssf.Timing("import.response_duration_ns", time.Since(span.Start),
		time.Nanosecond, map[string]string{"part": "merge"}))

	return &empty.Empty{}, nil
}
