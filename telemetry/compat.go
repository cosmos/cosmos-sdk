package telemetry

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"sync"

	gometrics "github.com/hashicorp/go-metrics"
	"go.opentelemetry.io/otel/attribute"
	otelmetric "go.opentelemetry.io/otel/metric"
)

type otelGoMetricsSink struct {
	meter      otelmetric.Meter
	ctx        context.Context
	counters   sync.Map
	gauges     sync.Map
	histograms sync.Map
}

func newOtelGoMetricsSink(ctx context.Context, meter otelmetric.Meter) gometrics.MetricSink {
	return &otelGoMetricsSink{
		meter: meter,
		ctx:   ctx,
	}
}

func (o *otelGoMetricsSink) getGauge(key string) otelmetric.Float64Gauge {
	entry, ok := o.gauges.Load(key)
	if ok {
		return entry.(otelmetric.Float64Gauge)
	}
	gauge, err := o.meter.Float64Gauge(key)
	if err != nil {
		panic(fmt.Sprintf("failed to create gauge metric %s: %v", key, err))
	}
	o.gauges.Store(key, gauge)
	return gauge
}

func (o *otelGoMetricsSink) SetGauge(key []string, val float32) {
	o.getGauge(flattenKey(key)).Record(o.ctx, float64(val))
}

func (o *otelGoMetricsSink) SetGaugeWithLabels(key []string, val float32, labels []gometrics.Label) {
	o.getGauge(flattenKey(key)).Record(o.ctx, float64(val), toOtelAttrs(labels))
}

func (o *otelGoMetricsSink) getCounter(key string) otelmetric.Float64Counter {
	entry, ok := o.counters.Load(key)
	if ok {
		return entry.(otelmetric.Float64Counter)
	}
	counter, err := o.meter.Float64Counter(key)
	if err != nil {
		panic(fmt.Sprintf("failed to create counter metric %s: %v", key, err))
	}
	o.counters.Store(key, counter)
	return counter
}

func (o *otelGoMetricsSink) IncrCounter(key []string, val float32) {
	o.getCounter(flattenKey(key)).Add(o.ctx, float64(val))
}

func (o *otelGoMetricsSink) IncrCounterWithLabels(key []string, val float32, labels []gometrics.Label) {
	o.getCounter(flattenKey(key)).Add(o.ctx, float64(val), toOtelAttrs(labels))
}

func (o *otelGoMetricsSink) getHistogram(key string) otelmetric.Float64Histogram {
	entry, ok := o.histograms.Load(key)
	if ok {
		return entry.(otelmetric.Float64Histogram)
	}
	// Otel doesn't have a histogram summary like go-metrics does, and the default bucket boundaries are not suitable for
	// the main processes that use it, so we will set some reasonable default here.
	// If these defaults are not suitable the user can override them by supplying their own values in the otel yaml config.
	// See example here: https://github.com/open-telemetry/opentelemetry-configuration/blob/65274d8cde98640e91932903e35287b330e482f4/examples/kitchen-sink.yaml#L177
	hist, err := o.meter.Float64Histogram(key, otelmetric.WithExplicitBucketBoundaries(0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10))
	if err != nil {
		panic(fmt.Sprintf("failed to create histogram metric %s: %v", key, err))
	}
	o.histograms.Store(key, hist)
	return hist
}

func (o *otelGoMetricsSink) EmitKey(key []string, val float32) {
	o.getHistogram(flattenKey(key)).Record(o.ctx, float64(val))
}

func (o *otelGoMetricsSink) AddSample(key []string, val float32) {
	o.getHistogram(flattenKey(key)).Record(o.ctx, float64(val))
}

func (o *otelGoMetricsSink) AddSampleWithLabels(key []string, val float32, labels []gometrics.Label) {
	o.getHistogram(flattenKey(key)).Record(o.ctx, float64(val), toOtelAttrs(labels))
}

func (o *otelGoMetricsSink) SetPrecisionGauge(key []string, val float64) {
	o.getGauge(flattenKey(key)).Record(o.ctx, val)
}

func (o *otelGoMetricsSink) SetPrecisionGaugeWithLabels(key []string, val float64, labels []gometrics.Label) {
	o.getGauge(flattenKey(key)).Record(o.ctx, val, toOtelAttrs(labels))
}

var (
	_ gometrics.MetricSink               = &otelGoMetricsSink{}
	_ gometrics.PrecisionGaugeMetricSink = &otelGoMetricsSink{}
)

var spaceReplacer = strings.NewReplacer(" ", "_")

// NOTE: this code was copied from https://github.com/hashicorp/go-metrics/blob/v0.5.4/inmem.go
func flattenKey(parts []string) string {
	buf := &bytes.Buffer{}

	joined := strings.Join(parts, ".")

	spaceReplacer.WriteString(buf, joined) //nolint: errcheck // unlikely and non-critical.

	return buf.String()
}

func toOtelAttrs(labels []gometrics.Label) otelmetric.MeasurementOption {
	attrs := make([]attribute.KeyValue, len(labels))
	for i, l := range labels {
		attrs[i] = attribute.String(l.Name, l.Value)
	}
	return otelmetric.WithAttributes(attrs...)
}
