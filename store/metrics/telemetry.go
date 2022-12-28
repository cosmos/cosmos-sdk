package metrics

import (
	"time"

	"github.com/armon/go-metrics"
)

// StoreMetrics defines the set of metrics for the store package
type StoreMetrics interface {
	MeasureSince(keys ...string)
}

var (
	_ StoreMetrics = Metrics{}
	_ StoreMetrics = NoOpMetrics{}
)

// Metrics defines the metrics wrapper for the store package
type Metrics struct {
	Labels []metrics.Label
}

func NewMetrics(labels ...metrics.Label) Metrics {
	return Metrics{
		Labels: labels,
	}
}

// MeasureSince provides a wrapper functionality for emitting a a time measure
// metric with global labels (if any).
func (m Metrics) MeasureSince(keys ...string) {
	start := time.Now()
	metrics.MeasureSinceWithLabels(keys, start.UTC(), m.Labels)
}

// NoOpMetrics is a no-op implementation of the StoreMetrics interface
type NoOpMetrics struct{}

func NewNoOpMetrics() NoOpMetrics {
	return NoOpMetrics{}
}

// MeasureSince is a no-op implementation of the StoreMetrics interface to avoid time.Now() calls
func (m NoOpMetrics) MeasureSince(keys ...string) {}
