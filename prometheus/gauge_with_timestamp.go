package prometheus

import (
	"math"
	"sync/atomic"

	"github.com/golang/protobuf/proto"

	dto "github.com/prometheus/client_model/go"
)

type GaugeWithTimestamp interface {
	Metric
	Collector

	// SetTimestamp sets the timestamp as given value
	SetWithTimestamp(float64, float64)
}

type gaugeWithTimestamp struct {
	// valBits contains the bits of the represented float64 value. It has
	// to go first in the struct to guarantee alignment for atomic
	// operations.  http://golang.org/pkg/sync/atomic/#pkg-note-BUG
	valBits uint64
	time    float64

	selfCollector

	desc       *Desc
	labelPairs []*dto.LabelPair
}

func NewGaugeWithTimestamp(opts GaugeOpts) GaugeWithTimestamp {
	desc := NewDesc(
		BuildFQName(opts.Namespace, opts.Subsystem, opts.Name),
		opts.Help,
		nil,
		opts.ConstLabels,
	)
	result := &gaugeWithTimestamp{desc: desc, labelPairs: desc.constLabelPairs}
	result.init(result) // Init self-collection.
	return result
}

func (g *gaugeWithTimestamp) Desc() *Desc {
	return g.desc
}

func (g *gaugeWithTimestamp) SetWithTimestamp(val float64, time float64) {
	atomic.StoreUint64(&g.valBits, math.Float64bits(val))
	g.time = time
}

func (g *gaugeWithTimestamp) Write(out *dto.Metric) error {
	val := math.Float64frombits(atomic.LoadUint64(&g.valBits))
	m := populateMetric(GaugeValue, val, g.labelPairs, nil, out)
	out.TimestampMs = proto.Int64(int64(g.time * 1000))
	return m
}

type GaugeWithTimestampVec struct {
	*metricVec
}

func NewGaugeWithTimestampVec(opts GaugeOpts, labelNames []string) *GaugeWithTimestampVec {
	desc := NewDesc(
		BuildFQName(opts.Namespace, opts.Subsystem, opts.Name),
		opts.Help,
		labelNames,
		opts.ConstLabels,
	)
	return &GaugeWithTimestampVec{
		metricVec: newMetricVec(desc, func(lvs ...string) Metric {
			if len(lvs) != len(desc.variableLabels) {
				panic(makeInconsistentCardinalityError(desc.fqName, desc.variableLabels, lvs))
			}
			result := &gaugeWithTimestamp{desc: desc, labelPairs: makeLabelPairs(desc, lvs)}
			result.init(result) // Init self-collection.
			return result
		}),
	}
}

func (v *GaugeWithTimestampVec) GetMetricWith(labels Labels) (GaugeWithTimestamp, error) {
	metric, err := v.metricVec.getMetricWith(labels)
	if metric != nil {
		return metric.(GaugeWithTimestamp), err
	}
	return nil, err
}

func (v *GaugeWithTimestampVec) With(labels Labels) GaugeWithTimestamp {
	g, err := v.GetMetricWith(labels)
	if err != nil {
		panic(err)
	}
	return g
}
