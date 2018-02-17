package samplers

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/axiomhq/hyperloglog"
	"github.com/stripe/veneur/tdigest"
)

// MetricType defines what kind of metric this is, so that we or our upstream
// sinks can do the right thing with it.
type MetricType int

const (
	// CounterMetric is a counter
	CounterMetric MetricType = iota
	// GaugeMetric is a gauge
	GaugeMetric
)

// RouteInformation is a key-only map indicating sink names that are
// supposed to receive a metric. A nil RouteInformation value
// corresponds to the "every sink" value; an entry in a non-nil
// RouteInformation means that the key should receive the metric.
type RouteInformation map[string]struct{}

// RouteTo returns true if the named sink should receive a metric
// according to the route table. A nil route table causes any sink to
// be eligible for the metric.
func (ri RouteInformation) RouteTo(name string) bool {
	if ri == nil {
		return true
	}
	_, ok := ri[name]
	return ok
}

// InterMetric represents a metric that has been completed and is ready for
// flushing by sinks.
type InterMetric struct {
	Name      string
	Timestamp int64
	Value     float64
	Tags      []string
	Type      MetricType

	// Sinks, if non-nil, indicates which metric sinks a metric
	// should be inserted into. If nil, that means the metric is
	// meant to go to every sink.
	Sinks RouteInformation
}

type Aggregate int

const (
	AggregateMin Aggregate = 1 << iota
	AggregateMax
	AggregateMedian
	AggregateAverage
	AggregateCount
	AggregateSum
	AggregateHarmonicMean
)

var AggregatesLookup = map[string]Aggregate{
	"min":    AggregateMin,
	"max":    AggregateMax,
	"median": AggregateMedian,
	"avg":    AggregateAverage,
	"count":  AggregateCount,
	"sum":    AggregateSum,
	"hmean":  AggregateHarmonicMean,
}

type HistogramAggregates struct {
	Value Aggregate
	Count int
}

var aggregates = [...]string{
	AggregateMin:          "min",
	AggregateMax:          "max",
	AggregateMedian:       "median",
	AggregateAverage:      "avg",
	AggregateCount:        "count",
	AggregateSum:          "sum",
	AggregateHarmonicMean: "hmean",
}

// JSONMetric is used to represent a metric that can be remarshaled with its
// internal state intact. It is used to send metrics from one Veneur to another.
type JSONMetric struct {
	MetricKey
	Tags []string `json:"tags"`
	// the Value is an internal representation of the metric's contents, eg a
	// gob-encoded histogram or hyperloglog.
	Value []byte      `json:"value"`
	Scope MetricScope `json:"scope"`
}

const sinkPrefix string = "veneursinkonly:"

func routeInfo(tags []string) RouteInformation {
	var info RouteInformation
	for _, tag := range tags {
		if !strings.HasPrefix(tag, sinkPrefix) {
			continue
		}
		if info == nil {
			info = make(RouteInformation)
		}
		// Take the tag suffix (the part after the ':' in
		// "veneursinkonly:", and make that the key in our
		// route information map:
		info[tag[len(sinkPrefix):]] = struct{}{}
	}
	return info
}

// Counter is an accumulator
type Counter struct {
	Name  string
	Tags  []string
	value int64
}

// Sample adds a sample to the counter.
func (c *Counter) Sample(sample float64, sampleRate float32) {
	c.value += int64(sample) * int64(1/sampleRate)
}

// Flush generates an InterMetric from the current state of this Counter.
func (c *Counter) Flush(interval time.Duration) []InterMetric {
	tags := make([]string, len(c.Tags))
	copy(tags, c.Tags)
	return []InterMetric{{
		Name:      c.Name,
		Timestamp: time.Now().Unix(),
		Value:     float64(c.value),
		Tags:      tags,
		Type:      CounterMetric,
		Sinks:     routeInfo(tags),
	}}
}

// Export converts a Counter into a JSONMetric which reports the rate.
func (c *Counter) Export() (JSONMetric, error) {
	buf := new(bytes.Buffer)

	err := binary.Write(buf, binary.LittleEndian, c.value)
	if err != nil {
		return JSONMetric{}, err
	}

	return JSONMetric{
		MetricKey: MetricKey{
			Name:       c.Name,
			Type:       "counter",
			JoinedTags: strings.Join(c.Tags, ","),
		},
		Tags:  c.Tags,
		Value: buf.Bytes(),
	}, nil
}

// Combine merges the values seen with another set (marshalled as a byte slice)
func (c *Counter) Combine(other []byte) error {
	var otherCounts int64
	buf := bytes.NewReader(other)
	err := binary.Read(buf, binary.LittleEndian, &otherCounts)

	if err != nil {
		return err
	}

	c.value += otherCounts

	return nil
}

// GetName returns the Name of the counter.
func (c *Counter) GetName() string {
	return c.Name
}

// NewCounter generates and returns a new Counter.
func NewCounter(Name string, Tags []string) *Counter {
	return &Counter{Name: Name, Tags: Tags}
}

// Gauge retains whatever the last value was.
type Gauge struct {
	Name  string
	Tags  []string
	value float64
}

// Sample takes on whatever value is passed in as a sample.
func (g *Gauge) Sample(sample float64, sampleRate float32) {
	g.value = sample
}

// Flush generates an InterMetric from the current state of this gauge.
func (g *Gauge) Flush() []InterMetric {
	tags := make([]string, len(g.Tags))
	copy(tags, g.Tags)
	return []InterMetric{{
		Name:      g.Name,
		Timestamp: time.Now().Unix(),
		Value:     float64(g.value),
		Tags:      tags,
		Type:      GaugeMetric,
		Sinks:     routeInfo(tags),
	}}
}

// Export converts a Gauge into a JSONMetric.
func (g *Gauge) Export() (JSONMetric, error) {
	var buf bytes.Buffer

	err := binary.Write(&buf, binary.LittleEndian, g.value)
	if err != nil {
		return JSONMetric{}, err
	}

	return JSONMetric{
		MetricKey: MetricKey{
			Name:       g.Name,
			Type:       "gauge",
			JoinedTags: strings.Join(g.Tags, ","),
		},
		Tags:  g.Tags,
		Value: buf.Bytes(),
	}, nil
}

// Combine is pretty naïve for Gauges, as it just overwrites the value.
func (g *Gauge) Combine(other []byte) error {
	var otherValue float64
	buf := bytes.NewReader(other)
	err := binary.Read(buf, binary.LittleEndian, &otherValue)

	if err != nil {
		return err
	}

	g.value = otherValue

	return nil
}

// GetName returns the Name of the Gauge
func (g *Gauge) GetName() string {
	return g.Name
}

// NewGauge genearaaaa who am I kidding just getting rid of the warning.
func NewGauge(Name string, Tags []string) *Gauge {
	return &Gauge{Name: Name, Tags: Tags}
}

// Set is a list of unique values seen.
type Set struct {
	Name string
	Tags []string
	Hll  *hyperloglog.Sketch
}

// Sample checks if the supplied value has is already in the filter. If not, it increments
// the counter!
func (s *Set) Sample(sample string, sampleRate float32) {
	s.Hll.Insert([]byte(sample))
}

// NewSet generates a new Set and returns it
func NewSet(Name string, Tags []string) *Set {
	// error is only returned if precision is outside the 4-18 range
	// TODO: this is the maximum precision, should it be configurable?
	Hll := hyperloglog.New()
	return &Set{
		Name: Name,
		Tags: Tags,
		Hll:  Hll,
	}
}

// Flush generates an InterMetric for the state of this Set.
func (s *Set) Flush() []InterMetric {
	tags := make([]string, len(s.Tags))
	copy(tags, s.Tags)
	return []InterMetric{{
		Name:      s.Name,
		Timestamp: time.Now().Unix(),
		Value:     float64(s.Hll.Estimate()),
		Tags:      tags,
		Type:      GaugeMetric,
		Sinks:     routeInfo(tags),
	}}
}

// Export converts a Set into a JSONMetric which reports the Tags in the set.
func (s *Set) Export() (JSONMetric, error) {
	val, err := s.Hll.MarshalBinary()
	if err != nil {
		return JSONMetric{}, err
	}
	return JSONMetric{
		MetricKey: MetricKey{
			Name:       s.Name,
			Type:       "set",
			JoinedTags: strings.Join(s.Tags, ","),
		},
		Tags:  s.Tags,
		Value: val,
	}, nil
}

// Combine merges the values seen with another set (marshalled as a byte slice)
func (s *Set) Combine(other []byte) error {
	otherHLL := hyperloglog.New()
	if err := otherHLL.UnmarshalBinary(other); err != nil {
		return err
	}
	if err := s.Hll.Merge(otherHLL); err != nil {
		// does not error unless compressions are different
		// however, decoding the other Hll causes us to use its compression
		// parameter, which might be different from ours
		return err
	}
	return nil
}

// GetName returns the Name of the Set
func (s *Set) GetName() string {
	return s.Name
}

// Histo is a collection of values that generates max, min, count, and
// percentiles over time.
type Histo struct {
	Name    string
	Tags    []string
	tDigest *tdigest.MergingDigest
	// these values are computed from only the samples that came through this
	// veneur instance, ignoring any histograms merged from elsewhere
	// we separate them because they're easy to aggregate on the backend without
	// loss of granularity, and having host-local information on them might be
	// useful
	weight        float64
	min           float64
	max           float64
	sum           float64
	reciprocalSum float64
}

// Sample adds the supplied value to the histogram.
func (h *Histo) Sample(sample float64, sampleRate float32) {
	weight := float64(1 / sampleRate)
	h.tDigest.Add(sample, weight)

	h.weight += weight
	h.min = math.Min(h.min, sample)
	h.max = math.Max(h.max, sample)
	h.sum += sample * weight

	h.reciprocalSum += (1 / sample) * weight
}

// NewHist generates a new Histo and returns it.
func NewHist(Name string, Tags []string) *Histo {
	return &Histo{
		Name: Name,
		Tags: Tags,
		// we're going to allocate a lot of these, so we don't want them to be huge
		tDigest: tdigest.NewMerging(100, false),
		min:     math.Inf(+1),
		max:     math.Inf(-1),
		sum:     0,
	}
}

// Flush generates InterMetrics for the current state of the Histo. percentiles
// indicates what percentiles should be exported from the histogram.
func (h *Histo) Flush(interval time.Duration, percentiles []float64, aggregates HistogramAggregates) []InterMetric {
	now := time.Now().Unix()
	metrics := make([]InterMetric, 0, aggregates.Count+len(percentiles))
	sinks := routeInfo(h.Tags)

	if (aggregates.Value&AggregateMax) == AggregateMax && !math.IsInf(h.max, 0) {
		// Defensively recopy tags to avoid aliasing bugs in case multiple InterMetrics share the same
		// tag array in the future
		tags := make([]string, len(h.Tags))
		copy(tags, h.Tags)
		metrics = append(metrics, InterMetric{
			Name:      fmt.Sprintf("%s.max", h.Name),
			Timestamp: now,
			Value:     float64(h.max),
			Tags:      tags,
			Type:      GaugeMetric,
			Sinks:     sinks,
		})
	}
	if (aggregates.Value&AggregateMin) == AggregateMin && !math.IsInf(h.min, 0) {
		tags := make([]string, len(h.Tags))
		copy(tags, h.Tags)
		metrics = append(metrics, InterMetric{
			Name:      fmt.Sprintf("%s.min", h.Name),
			Timestamp: now,
			Value:     float64(h.min),
			Tags:      tags,
			Type:      GaugeMetric,
			Sinks:     sinks,
		})
	}

	if (aggregates.Value&AggregateSum) == AggregateSum && h.sum != 0 {
		tags := make([]string, len(h.Tags))
		copy(tags, h.Tags)
		metrics = append(metrics, InterMetric{
			Name:      fmt.Sprintf("%s.sum", h.Name),
			Timestamp: now,
			Value:     float64(h.sum),
			Tags:      tags,
			Type:      GaugeMetric,
			Sinks:     sinks,
		})
	}

	if (aggregates.Value&AggregateAverage) == AggregateAverage && h.sum != 0 && h.weight != 0 {
		// we need both a rate and a non-zero sum before it will make sense
		// to submit an average
		tags := make([]string, len(h.Tags))
		copy(tags, h.Tags)
		metrics = append(metrics, InterMetric{
			Name:      fmt.Sprintf("%s.avg", h.Name),
			Timestamp: now,
			Value:     float64(h.sum / h.weight),
			Tags:      tags,
			Type:      GaugeMetric,
			Sinks:     sinks,
		})
	}

	if (aggregates.Value&AggregateCount) == AggregateCount && h.weight != 0 {
		// if we haven't received any local samples, then leave this sparse,
		// otherwise it can lead to some misleading zeroes in between the
		// flushes of downstream instances
		tags := make([]string, len(h.Tags))
		copy(tags, h.Tags)
		metrics = append(metrics, InterMetric{
			Name:      fmt.Sprintf("%s.count", h.Name),
			Timestamp: now,
			Value:     float64(h.weight),
			Tags:      tags,
			Type:      CounterMetric,
			Sinks:     sinks,
		})
	}

	if (aggregates.Value & AggregateMedian) == AggregateMedian {
		tags := make([]string, len(h.Tags))
		copy(tags, h.Tags)
		metrics = append(
			metrics,
			InterMetric{
				Name:      fmt.Sprintf("%s.median", h.Name),
				Timestamp: now,
				Value:     float64(h.tDigest.Quantile(0.5)),
				Tags:      tags,
				Type:      GaugeMetric,
				Sinks:     sinks,
			},
		)
	}

	if (aggregates.Value&AggregateHarmonicMean) == AggregateHarmonicMean && h.reciprocalSum != 0 && h.weight != 0 {
		// we need both a rate and a non-zero sum before it will make sense
		// to submit an average
		tags := make([]string, len(h.Tags))
		copy(tags, h.Tags)
		metrics = append(metrics, InterMetric{
			Name:      fmt.Sprintf("%s.hmean", h.Name),
			Timestamp: now,
			Value:     float64(h.weight / h.reciprocalSum),
			Tags:      tags,
			Type:      GaugeMetric,
			Sinks:     sinks,
		})
	}

	for _, p := range percentiles {
		tags := make([]string, len(h.Tags))
		copy(tags, h.Tags)
		metrics = append(
			metrics,
			// TODO Fix to allow for p999, etc
			InterMetric{
				Name:      fmt.Sprintf("%s.%dpercentile", h.Name, int(p*100)),
				Timestamp: now,
				Value:     float64(h.tDigest.Quantile(p)),
				Tags:      tags,
				Type:      GaugeMetric,
				Sinks:     sinks,
			},
		)
	}

	return metrics
}

// HistoValue is a serializable version of a Histo that will be sent as the
// Value of a JSONMetric, gob-encoded.
type HistoValue struct {
	TDigest       *tdigest.MergingDigest
	Weight        float64
	Min           float64
	Max           float64
	Sum           float64
	ReciprocalSum float64
}

// Export converts a Histogram into a JSONMetric
func (h *Histo) Export() (JSONMetric, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	hval := HistoValue{
		TDigest:       h.tDigest,
		Weight:        h.weight,
		Min:           h.min,
		Max:           h.max,
		Sum:           h.sum,
		ReciprocalSum: h.reciprocalSum,
	}
	if err := enc.Encode(hval); err != nil {
		return JSONMetric{}, err
	}
	return JSONMetric{
		MetricKey: MetricKey{
			Name:       h.Name,
			Type:       "histogram",
			JoinedTags: strings.Join(h.Tags, ","),
		},
		Tags:  h.Tags,
		Value: buf.Bytes(),
	}, nil
}

// Combine merges the values of a histogram with another histogram
// (marshalled as a byte slice)
func (h *Histo) Combine(other []byte) error {
	var val HistoValue
	dec := gob.NewDecoder(bytes.NewReader(other))

	if err := dec.Decode(&val); err != nil {
		// To support the old binary format, try to directly decode just
		// a tdigest.  This should probably be removed in a future breaking
		// release. This allows upgrading global instances first, while still
		// being compatible with non-updated local instances.
		otherHistogram := tdigest.NewMerging(100, false)
		if err := otherHistogram.GobDecode(other); err != nil {
			return fmt.Errorf("failed to unmarshal the Histo value: %v", err)
		}

		h.tDigest.Merge(otherHistogram)
		return nil
	}

	h.tDigest.Merge(val.TDigest)
	h.weight += val.Weight
	h.min = math.Min(h.min, val.Min)
	h.max = math.Max(h.max, val.Max)
	h.sum += val.Sum
	h.reciprocalSum += val.ReciprocalSum
	return nil
}

// GetName returns the Name of the Histo
func (h *Histo) GetName() string {
	return h.Name
}
