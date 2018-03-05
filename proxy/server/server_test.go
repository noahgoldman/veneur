package proxyserver

import (
	"context"
	"math/rand"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stripe/veneur/forwardrpc"
	"github.com/stripe/veneur/forwardrpc/forwardtest"
	"github.com/stripe/veneur/samplers/metricpb"
	"stathat.com/c/consistent"
)

func createTestForwardServers(t *testing.T, n int) []*forwardtest.Server {
	t.Helper()

	res := make([]*forwardtest.Server, n)
	for i := range res {
		res[i] = forwardtest.NewServer()
		res[i].Start(t)
	}

	return res
}

func stopTestForwardServers(ss []*forwardtest.Server) {
	for _, s := range ss {
		s.Stop()
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
			ring.Add(dest.Addr().String())
		}

		expected := makeForwardMetrics(100)

		server := New(ring, nil)
		err := server.sendMetrics(context.Background(), &forwardrpc.MetricList{expected})
		assert.NoError(t, err, "sendMetrics shouldn't have failed")

		// Get all of the metrics from the destination servers
		var actual []*metricpb.Metric
		for _, dest := range dests {
			assert.NotEmpty(t, dest.Metrics(), "The server at '%s' got zero metrics. "+
				"The chunking logic may be incorrect", dest.Addr().String())
			actual = append(actual, dest.Metrics()...)
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
		ring.Add(dest.Addr().String())
	}

	server := New(ring, &Options{Timeout: 1})
	err := server.sendMetrics(context.Background(),
		&forwardrpc.MetricList{makeForwardMetrics(10)})
	assert.Error(t, err, "sendMetrics should have returned an error when the "+
		"timeout was set to effectively zero")
}
