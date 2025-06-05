package telemetry

import (
	"context"
	"encoding/base64"
	"fmt"
	"math"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	otmetric "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"

	"cosmossdk.io/log"
)

const meterName = "cosmos-sdk-otlp-exporter"

// StartOtlpExporter sets up and runs the OTLP exporter.
func StartOtlpExporter(ctx context.Context, logger log.Logger, cfg OtlpConfig) error {
	exporter, err := otlpmetrichttp.New(ctx,
		otlpmetrichttp.WithEndpoint(cfg.CollectorEndpoint),
		otlpmetrichttp.WithURLPath(cfg.CollectorMetricsURLPath),
		otlpmetrichttp.WithHeaders(map[string]string{
			"Authorization": "Basic " + formatBasicAuth(cfg.User, cfg.Token),
		}),
	)
	if err != nil {
		return fmt.Errorf("OTLP exporter setup failed: %w", err)
	}

	res, err := resource.New(ctx, resource.WithAttributes(
		semconv.ServiceName(cfg.ServiceName),
	))
	if err != nil {
		return fmt.Errorf("OTLP resource creation failed: %w", err)
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(exporter,
			metric.WithInterval(cfg.PushInterval))),
		metric.WithResource(res),
	)
	otel.SetMeterProvider(meterProvider)
	meter := otel.Meter(meterName)

	defer func() {
		if err := meterProvider.Shutdown(ctx); err != nil {
			logger.Error("failed to shut down OTLP MeterProvider: %v", err)
		}
	}()

	gauges := make(map[string]otmetric.Float64Gauge)
	histograms := make(map[string]otmetric.Float64Histogram)
	counters := make(map[string]otmetric.Float64Counter)

	go func() {
		ticker := time.NewTicker(cfg.PushInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := scrapePrometheusMetrics(ctx, meter, logger, gauges, histograms, counters); err != nil {
					logger.Debug("error scraping metrics: %v", err)
				}
			}
		}
	}()

	return nil
}

// scrapePrometheusMetrics gathers and forwards Prometheus metrics to OTLP.
func scrapePrometheusMetrics(
	ctx context.Context,
	meter otmetric.Meter,
	logger log.Logger,
	gauges map[string]otmetric.Float64Gauge,
	histograms map[string]otmetric.Float64Histogram,
	counters map[string]otmetric.Float64Counter,
) error {
	mfs, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		return fmt.Errorf("failed to gather prometheus metrics: %w", err)
	}

	for _, mf := range mfs {
		name := mf.GetName()
		for _, m := range mf.Metric {
			attrs := make([]attribute.KeyValue, len(m.Label))
			for i, label := range m.Label {
				attrs[i] = attribute.String(label.GetName(), label.GetValue())
			}

			switch mf.GetType() {
			case dto.MetricType_GAUGE:
				recordGauge(ctx, meter, logger, gauges, name, mf.GetHelp(), m.Gauge.GetValue(), attrs)
			case dto.MetricType_COUNTER:
				recordCounter(ctx, meter, logger, counters, name, mf.GetHelp(), m.Counter.GetValue(), attrs)
			case dto.MetricType_HISTOGRAM:
				recordHistogram(ctx, meter, logger, histograms, name, mf.GetHelp(), m.Histogram)
			case dto.MetricType_SUMMARY:
				recordSummary(ctx, meter, logger, gauges, name, mf.GetHelp(), m.Summary)
			default:
				continue
			}
		}
	}

	return nil
}

// recordCounter sends a Prometheus counter as an OTLP counter.
func recordCounter(
	ctx context.Context,
	meter otmetric.Meter,
	logger log.Logger,
	counters map[string]otmetric.Float64Counter,
	name, help string,
	val float64,
	attrs []attribute.KeyValue,
) {
	c, ok := counters[name]
	if !ok {
		var err error
		c, err = meter.Float64Counter(name, otmetric.WithDescription(help))
		if err != nil {
			logger.Debug("failed to create counter %q: %v", name, err)
			return
		}
		counters[name] = c
	}
	c.Add(ctx, val, otmetric.WithAttributes(attrs...))
}

// recordGauge sends a Prometheus gauge as an OTLP gauge.
func recordGauge(
	ctx context.Context,
	meter otmetric.Meter,
	logger log.Logger,
	gauges map[string]otmetric.Float64Gauge,
	name, help string,
	val float64,
	attrs []attribute.KeyValue,
) {
	g, ok := gauges[name]
	if !ok {
		var err error
		g, err = meter.Float64Gauge(name, otmetric.WithDescription(help))
		if err != nil {
			logger.Debug("failed to create gauge %q: %v", name, err)
			return
		}
		gauges[name] = g
	}
	g.Record(ctx, val, otmetric.WithAttributes(attrs...))
}

// recordHistogram sends a Prometheus histogram as an OTLP histogram.
func recordHistogram(
	ctx context.Context,
	meter otmetric.Meter,
	logger log.Logger,
	histograms map[string]otmetric.Float64Histogram,
	name, help string,
	h *dto.Histogram,
) {
	bounds := make([]float64, 0, len(h.Bucket)-1)
	counts := make([]uint64, 0, len(h.Bucket))

	for _, bucket := range h.Bucket {
		if math.IsInf(bucket.GetUpperBound(), +1) {
			continue
		}
		bounds = append(bounds, bucket.GetUpperBound())
		counts = append(counts, bucket.GetCumulativeCount())
	}

	hist, ok := histograms[name]
	if !ok {
		var err error
		hist, err = meter.Float64Histogram(
			name,
			otmetric.WithDescription(help),
			otmetric.WithExplicitBucketBoundaries(bounds...),
		)
		if err != nil {
			logger.Debug("failed to create histogram %s: %v", name, err)
			return
		}
		histograms[name] = hist
	}

	var prev uint64
	for i, count := range counts {
		n := count - prev
		prev = count

		var value float64
		if i == 0 {
			value = bounds[0] / 2.0
		} else {
			value = (bounds[i-1] + bounds[i]) / 2.0
		}

		for j := uint64(0); j < n; j++ {
			hist.Record(ctx, value)
		}
	}
}

// recordSummary sends a Prometheus summary as OTLP gauges.
func recordSummary(
	ctx context.Context,
	meter otmetric.Meter,
	logger log.Logger,
	gauges map[string]otmetric.Float64Gauge,
	name, help string,
	s *dto.Summary,
) {
	recordGauge(ctx, meter, logger, gauges, name+"_sum", help+" sum", s.GetSampleSum(), nil)
	recordGauge(ctx, meter, logger, gauges, name+"_count", help+" count", float64(s.GetSampleCount()), nil)

	for _, q := range s.Quantile {
		attrs := []attribute.KeyValue{
			attribute.String("quantile", fmt.Sprintf("%v", q.GetQuantile())),
		}
		recordGauge(ctx, meter, logger, gauges, name, help+" quantile", q.GetValue(), attrs)
	}
}

// formatBasicAuth returns a base64-encoded Basic Auth token.
func formatBasicAuth(username, token string) string {
	auth := username + ":" + token
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
