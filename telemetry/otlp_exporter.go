package telemetry

import (
	"context"
	"log"
	"math"
	"net/http"
	"time"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	otmetric "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

func StartOtlpExporter(cfg Config) {
	ctx := context.Background()

	exporter, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithEndpoint(cfg.OtlpCollectorGrpcAddr),
		otlpmetricgrpc.WithInsecure(),
	)
	if err != nil {
		log.Fatalf("OTLP exporter setup failed: %v", err)
	}

	res, _ := resource.New(ctx, resource.WithAttributes(
		semconv.ServiceName(cfg.OtlpServiceName),
	))

	meterProvider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(exporter,
			metric.WithInterval(15*time.Second))),
		metric.WithResource(res),
	)
	otel.SetMeterProvider(meterProvider)
	meter := otel.Meter("cosmos-sdk-otlp-exporter")

	gauges := make(map[string]otmetric.Float64Gauge)
	histograms := make(map[string]otmetric.Float64Histogram)

	go func() {
		for {
			if err := scrapeAndPushMetrics(ctx, cfg.PrometheusEndpoint, meter, gauges, histograms); err != nil {
				log.Printf("error scraping metrics: %v", err)
			}
			time.Sleep(15 * time.Second)
		}
	}()
}

func scrapeAndPushMetrics(ctx context.Context, promEndpoint string, meter otmetric.Meter, gauges map[string]otmetric.Float64Gauge, histograms map[string]otmetric.Float64Histogram) error {
	resp, err := http.Get(promEndpoint)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	parser := expfmt.TextParser{}
	metricsFamilies, err := parser.TextToMetricFamilies(resp.Body)
	if err != nil {
		return err
	}

	for name, mf := range metricsFamilies {
		for _, m := range mf.GetMetric() {
			switch {
			case m.Gauge != nil:
				recordGauge(ctx, meter, gauges, name, mf.GetHelp(), m.Gauge.GetValue())

			case m.Counter != nil:
				recordGauge(ctx, meter, gauges, name, mf.GetHelp(), m.Counter.GetValue())

			case m.Histogram != nil:
				recordHistogram(ctx, meter, histograms, name, mf.GetHelp(), m.Histogram)

			case m.Summary != nil:
				continue // TODO: decide whether to support

			default:
				continue
			}
		}
	}

	return nil
}

func recordGauge(ctx context.Context, meter otmetric.Meter, gauges map[string]otmetric.Float64Gauge, name, help string, val float64) {
	g, ok := gauges[name]
	if !ok {
		var err error
		g, err = meter.Float64Gauge(name, otmetric.WithDescription(help))
		if err != nil {
			log.Printf("failed to create gauge %q: %v", name, err)
			return
		}
		gauges[name] = g
	}
	g.Record(ctx, val)
}

func recordHistogram(ctx context.Context, meter otmetric.Meter, histograms map[string]otmetric.Float64Histogram, name, help string, h *dto.Histogram) {
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
			log.Printf("failed to create histogram %s: %v", name, err)
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
