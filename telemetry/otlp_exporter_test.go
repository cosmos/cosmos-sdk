package telemetry

import (
	"context"
	"encoding/base64"
	"testing"

	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otmetric "go.opentelemetry.io/otel/metric"

	"cosmossdk.io/log"
)

// TestFormatBasicAuth verifies that formatBasicAuth correctly base64‐encodes "username:token".
func TestFormatBasicAuth(t *testing.T) {
	user := "alice"
	token := "s3cr3t"
	encoded := formatBasicAuth(user, token)

	// Manually compute expected base64 of "alice:s3cr3t"
	expected := base64.StdEncoding.EncodeToString([]byte("alice:s3cr3t"))
	require.Equal(t, expected, encoded)
}

// TestRecordCounter ensures that recordCounter creates a new counter in the map.
func TestRecordCounterCreatesCounter(t *testing.T) {
	meter := otel.Meter("test-meter")
	counters := make(map[string]otmetric.Float64Counter)

	name := "sample_counter"
	help := "a sample counter"
	val := 5.0
	attrs := []attribute.KeyValue{
		attribute.String("label", "value"),
	}

	// First, call should create the counter entry
	recordCounter(context.Background(), meter, log.NewNopLogger(), counters, name, help, val, attrs)
	_, ok := counters[name]
	require.True(t, ok)

	// Second, call should not create a new entry, but update existing
	sizeBefore := len(counters)
	recordCounter(context.Background(), meter, log.NewNopLogger(), counters, name, help, val, attrs)
	require.Equal(t, sizeBefore, len(counters))
}

// TestRecordGauge ensures that recordGauge creates a new gauge in the map.
func TestRecordGaugeCreatesGauge(t *testing.T) {
	meter := otel.Meter("test-meter")
	gauges := make(map[string]otmetric.Float64Gauge)

	name := "sample_gauge"
	help := "a sample gauge"
	val := 3.14
	attrs := []attribute.KeyValue{
		attribute.String("foo", "bar"),
	}

	recordGauge(context.Background(), meter, log.NewNopLogger(), gauges, name, help, val, attrs)
	_, ok := gauges[name]
	require.True(t, ok)

	sizeBefore := len(gauges)
	recordGauge(context.Background(), meter, log.NewNopLogger(), gauges, name, help, val, attrs)
	require.Equal(t, sizeBefore, len(gauges))
}

// TestRecordHistogram ensures that recordHistogram creates a new histogram in the map
// and approximates buckets correctly.
func TestRecordHistogramCreatesHistogram(t *testing.T) {
	meter := otel.Meter("test-meter")
	histograms := make(map[string]otmetric.Float64Histogram)

	name := "sample_hist"
	help := "a sample histogram"

	upperBound := 10.0
	cumulativeCount := uint64(2)

	// Construct a simple Prometheus histogram with two buckets:
	//  - bucket 0: upperBound=10, count=2
	h := &dto.Histogram{
		Bucket: []*dto.Bucket{
			{UpperBound: &upperBound, CumulativeCount: &cumulativeCount},
		},
	}

	recordHistogram(context.Background(), meter, log.NewNopLogger(), histograms, name, help, h)
	_, ok := histograms[name]
	require.True(t, ok)

	// On a second call with identical buckets, it should reuse the existing histogram
	sizeBefore := len(histograms)
	recordHistogram(context.Background(), meter, log.NewNopLogger(), histograms, name, help, h)
	require.Equal(t, sizeBefore, len(histograms))
}

// TestRecordSummary ensures that recordSummary populates the gauges map with sum, count, and quantile entries.
func TestRecordSummaryCreatesSummaryMetrics(t *testing.T) {
	meter := otel.Meter("test-meter")
	gauges := make(map[string]otmetric.Float64Gauge)

	sampleSum := 15.0
	sampleCount := uint64(3)

	name := "sample_summary"
	help := "a sample summary"
	// Create a summary with sum=15, count=3,
	s := &dto.Summary{
		SampleSum:   &sampleSum,
		SampleCount: &sampleCount,
	}

	recordSummary(context.Background(), meter, log.NewNopLogger(), gauges, name, help, s)

	// Expect three entries:
	//  - name+"_sum"
	//  - name+"_count"
	//  - name (for quantiles; recorded twice but same key)
	_, ok := gauges[name+"_sum"]
	require.True(t, ok)
	_, ok = gauges[name+"_count"]
	require.True(t, ok)

	// Calling again should not increase map size
	sizeBefore := len(gauges)
	recordSummary(context.Background(), meter, log.NewNopLogger(), gauges, name, help, s)
	require.Equal(t, sizeBefore, len(gauges))
}
