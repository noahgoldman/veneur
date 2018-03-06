package testutils

import (
	"math/rand"
	"strconv"

	"github.com/stripe/veneur/samplers/metricpb"
)

func RandomForwardMetric() *metricpb.Metric {
	return &metricpb.Metric{
		Name:  strconv.Itoa(rand.Int()),
		Type:  metricpb.Type_Counter,
		Tags:  []string{strconv.Itoa(rand.Int())},
		Value: &metricpb.Metric_Counter{&metricpb.CounterValue{rand.Int63()}},
	}
}

func RandomForwardMetrics(n int) []*metricpb.Metric {
	res := make([]*metricpb.Metric, n)
	for i := range res {
		res[i] = RandomForwardMetric()
	}
	return res
}
