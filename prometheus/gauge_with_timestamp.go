package prometheus

import (
	"math"
	"sync/atomic"

	"github.com/golang/protobuf/proto"

	dto "github.com/prometheus/client_model/go"
)

type GaugeWithTimestamp interface {

	// SetTimestamp sets the timestamp as given value
	SetWithTimestamp(float64, float64)
}

type gaugeWithTimestamp struct {
	Gauge
	time    float64
}

func NewGaugeWithTimestamp(opts GaugeOpts) GaugeWithTimestamp {
	desc := NewDesc(
		BuildFQName(opts.Namespace, opts.Subsystem, opts.Name),
		opts.Help,
		nil,
		opts.ConstLabels,
	)
	g := NewGauge(opts)
	result := &gaugeWithTimestamp{Gauge: g}
	return result
}

func (g *gaugeWithTimestamp) Desc() *Desc {
	return g.Gauge.Desc()
}

func (g *gaugeWithTimestamp) SetWithTimestamp(val float64, time float64) {
	g.Gauge.Set(val)
	g.time = time
}

func (g *gaugeWithTimestamp) Write(out *dto.Metric) error {
	m := g.Gauge.Write(out)
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
