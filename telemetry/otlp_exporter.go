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

func StartOtlpExporter(ctx context.Context, logger log.Logger, cfg Config) error {
	exporter, err := otlpmetrichttp.New(ctx,
		otlpmetrichttp.WithEndpoint(cfg.OtlpCollectorEndpoint),
		otlpmetrichttp.WithURLPath(cfg.OtlpCollectorMetricsURLPath),
		otlpmetrichttp.WithHeaders(map[string]string{
			"Authorization": "Basic " + formatBasicAuth(cfg.OtlpUser, cfg.OtlpToken),
		}),
	)
	if err != nil {
		return fmt.Errorf("OTLP exporter setup failed: %w", err)
	}

	res, err := resource.New(ctx, resource.WithAttributes(
		semconv.ServiceName(cfg.OtlpServiceName),
	))
	if err != nil {
		return fmt.Errorf("OTLP resource creation failed: %w", err)
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(exporter,
			metric.WithInterval(cfg.OtlpPushInterval))),
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
		ticker := time.NewTicker(cfg.OtlpPushInterval)
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

func scrapePrometheusMetrics(
	ctx context.Context,
	meter otmetric.Meter,
	logger log.Logger,
	gauges map[string]otmetric.Float64Gauge,
	histograms map[string]otmetric.Float64Histogram,
	counters map[string]otmetric.Float64Counter,
) error {
	metricFamilies, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		logger.Debug("failed to gather prometheus metrics: %v", err)
		return err
	}

	for _, mf := range metricFamilies {
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

func recordCounter(
	ctx context.Context,
	meter otmetric.Meter,
	logger log.Logger,
	counters map[string]otmetric.Float64Counter,
	name, help string,
	val float64,
	attrs []attribute.KeyValue,
) {
	g, ok := counters[name]
	if !ok {
		var err error
		g, err = meter.Float64Counter(name, otmetric.WithDescription(help))
		if err != nil {
			logger.Debug("failed to create counter %q: %v", name, err)
			return
		}
		counters[name] = g
	}
	g.Add(ctx, val, otmetric.WithAttributes(attrs...))
}

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

func recordHistogram(
	ctx context.Context,
	meter otmetric.Meter,
	logger log.Logger,
	histograms map[string]otmetric.Float64Histogram,
	name, help string,
	h *dto.Histogram,
) {
	boundaries := make([]float64, 0, len(h.Bucket)-1) // excluding +Inf
	bucketCounts := make([]uint64, 0, len(h.Bucket))

	for _, bucket := range h.Bucket {
		if math.IsInf(bucket.GetUpperBound(), +1) {
			continue // Skip +Inf bucket boundary explicitly
		}
		boundaries = append(boundaries, bucket.GetUpperBound())
		bucketCounts = append(bucketCounts, bucket.GetCumulativeCount())
	}

	hist, ok := histograms[name]
	if !ok {
		var err error
		hist, err = meter.Float64Histogram(
			name,
			otmetric.WithDescription(help),
			otmetric.WithExplicitBucketBoundaries(boundaries...),
		)
		if err != nil {
			logger.Debug("failed to create histogram %s: %v", name, err)
			return
		}
		histograms[name] = hist
	}

	prevCount := uint64(0)
	for i, count := range bucketCounts {
		countInBucket := count - prevCount
		prevCount = count

		// Explicitly record the mid-point of the bucket as approximation:
		var value float64
		if i == 0 {
			value = boundaries[0] / 2.0
		} else {
			value = (boundaries[i-1] + boundaries[i]) / 2.0
		}

		// Record `countInBucket` number of observations explicitly (approximation):
		for j := uint64(0); j < countInBucket; j++ {
			hist.Record(ctx, value)
		}
	}
}

func recordSummary(ctx context.Context, meter otmetric.Meter, logger log.Logger, gauges map[string]otmetric.Float64Gauge, name, help string, s *dto.Summary) {
	recordGauge(ctx, meter, logger, gauges, name+"_sum", help+" (summary sum)", s.GetSampleSum(), nil)
	recordGauge(ctx, meter, logger, gauges, name+"_count", help+" (summary count)", float64(s.GetSampleCount()), nil)

	for _, q := range s.Quantile {
		attrs := []attribute.KeyValue{
			attribute.String("quantile", fmt.Sprintf("%v", q.GetQuantile())),
		}
		recordGauge(ctx, meter, logger, gauges, name, help+" (summary quantile)", q.GetValue(), attrs)
	}
}

func formatBasicAuth(username, token string) string {
	auth := username + ":" + token
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
