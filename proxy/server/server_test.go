package proxyserver

import (
	"context"
	"math/rand"
	"net"
	"strconv"
	"sync"
	"testing"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/stretchr/testify/assert"
	"github.com/stripe/veneur/forwardrpc"
	"github.com/stripe/veneur/samplers/metricpb"
	"google.golang.org/grpc"
	"stathat.com/c/consistent"
)

type testForwardServer struct {
	mtx     *sync.Mutex
	metrics []*metricpb.Metric
	lis     net.Listener
	server  *grpc.Server
}

func newTestForwardServer() *testForwardServer {
	return &testForwardServer{
		mtx: &sync.Mutex{},
	}
}

func (s *testForwardServer) start(t *testing.T) {
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

func (s *testForwardServer) stop() {
	s.server.Stop()
}

func (s *testForwardServer) SendMetrics(ctx context.Context, mlist *forwardrpc.MetricList) (*empty.Empty, error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.metrics = append(s.metrics, mlist.Metrics...)
	return &empty.Empty{}, nil
}

func createTestForwardServers(t *testing.T, n int) []*testForwardServer {
	t.Helper()

	res := make([]*testForwardServer, n)
	for i := range res {
		res[i] = newTestForwardServer()
		res[i].start(t)
	}

	return res
}

func stopTestForwardServers(ss []*testForwardServer) {
	for _, s := range ss {
		s.stop()
	}
}

func makeRandomForwardMetric() *metricpb.Metric {
	return &metricpb.Metric{
		Name:  strconv.Itoa(rand.Int()),
		Tags:  []string{strconv.Itoa(rand.Int())},
		Value: &metricpb.Metric_Counter{&metricpb.CounterValue{rand.Int63()}},
	}
}

func makeForwardMetrics(n int) []*metricpb.Metric {
	res := make([]*metricpb.Metric, 100)
	for i := range res {
		res[i] = makeRandomForwardMetric()
	}
	return res
}

// Test that it forwards a decent number of input metrics to many different
// destinations
func TestManyDestinations(t *testing.T) {
	// Test with many different numbers of forwarding destinations
	for numDests := 1; numDests < 10; numDests++ {
		dests := createTestForwardServers(t, numDests)
		defer stopTestForwardServers(dests)

		ring := consistent.New()
		for _, dest := range dests {
			ring.Add(dest.lis.Addr().String())
		}

		expected := makeForwardMetrics(100)

		server := New(ring, nil)
		err := server.sendMetrics(context.Background(), &forwardrpc.MetricList{expected})
		assert.NoError(t, err, "sendMetrics shouldn't have failed")

		// Get all of the metrics from the destination servers
		var actual []*metricpb.Metric
		for _, dest := range dests {
			assert.NotEmpty(t, dest.metrics, "The server at '%s' got zero metrics. "+
				"The chunking logic may be incorrect", dest.lis.Addr().String())
			actual = append(actual, dest.metrics...)
		}

		assert.ElementsMatch(t, expected, actual)
	}
}

func TestNoDestinations(t *testing.T) {
	server := New(consistent.New(), nil)
	err := server.sendMetrics(context.Background(),
		&forwardrpc.MetricList{makeForwardMetrics(10)})
	assert.Error(t, err, "sendMetrics should have returned an error when there "+
		"are no valid destinations")
}

func TestUnreachableDestinations(t *testing.T) {
	ring := consistent.New()
	ring.Add("not-a-real-host:9001")
	ring.Add("another-bad-host:9001")

	server := New(ring, nil)
	err := server.sendMetrics(context.Background(),
		&forwardrpc.MetricList{makeForwardMetrics(10)})
	assert.Error(t, err, "sendMetrics should have returned an error when all "+
		"of the destinations are unreachable")
}

func TestTimeout(t *testing.T) {
	dests := createTestForwardServers(t, 3)
	defer stopTestForwardServers(dests)

	ring := consistent.New()
	for _, dest := range dests {
		ring.Add(dest.lis.Addr().String())
	}

	server := New(ring, &Options{Timeout: 1})
	err := server.sendMetrics(context.Background(),
		&forwardrpc.MetricList{makeForwardMetrics(10)})
	assert.Error(t, err, "sendMetrics should have returned an error when the "+
		"timeout was set to effectively zero")
}
