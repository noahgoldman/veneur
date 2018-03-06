package forwardtest

import (
	"context"
	"net"
	"testing"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/stripe/veneur/forwardrpc"
	"github.com/stripe/veneur/samplers/metricpb"
	"google.golang.org/grpc"
)

type SendMetricHandler func([]*metricpb.Metric)

type Server struct {
	lis     net.Listener
	server  *grpc.Server
	handler SendMetricHandler
}

func NewServer(handler SendMetricHandler) *Server {
	return &Server{
		handler: handler,
	}
}

func (s *Server) Start(t *testing.T) {
	var err error
	s.lis, err = net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatalf("failed to create a TCP connection for a test GRPC "+
			"server: %v", err)
	}

	s.server = grpc.NewServer()
	forwardrpc.RegisterForwardServer(s.server, s)
	go func() {
		s.server.Serve(s.lis)
	}()
}

func (s *Server) Stop() {
	s.server.Stop()
}

func (s *Server) Addr() net.Addr {
	return s.lis.Addr()
}

func (s *Server) SendMetrics(ctx context.Context, mlist *forwardrpc.MetricList) (*empty.Empty, error) {
	s.handler(mlist.Metrics)
	return &empty.Empty{}, nil
}
