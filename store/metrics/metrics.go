package metrics

import (
	"time"

	"github.com/hashicorp/go-metrics"
)

var (
	_ StoreMetrics = Metrics{}
	_ StoreMetrics = NoOpMetrics{}
)

// StoreMetrics defines the set of supported metric APIs for the store package.
type StoreMetrics interface {
	MeasureSince(keys ...string)
}

// Metrics defines a default StoreMetrics implementation.
type Metrics struct {
	Labels []metrics.Label
}

// NewMetrics returns a new instance of the Metrics with labels set by the node
// operator.
func NewMetrics(labels [][]string) Metrics {
	m := Metrics{}

	if numGlobalLabels := len(labels); numGlobalLabels > 0 {
		parsedGlobalLabels := make([]metrics.Label, numGlobalLabels)
		for i, label := range labels {
			parsedGlobalLabels[i] = metrics.Label{Name: label[0], Value: label[1]}
		}

		m.Labels = parsedGlobalLabels
	}

	return m
}

// MeasureSince provides a wrapper functionality for emitting a time measure
// metric with global labels (if any).
func (m Metrics) MeasureSince(keys ...string) {
	start := time.Now()
	metrics.MeasureSinceWithLabels(keys, start.UTC(), m.Labels)
}

// NoOpMetrics is a no-op implementation of the StoreMetrics interface
type NoOpMetrics struct{}

// NewNoOpMetrics returns a new instance of the NoOpMetrics
func NewNoOpMetrics() NoOpMetrics {
	return NoOpMetrics{}
}

// MeasureSince is a no-op implementation of the StoreMetrics interface to avoid
// time.Now() calls.
func (NoOpMetrics) MeasureSince(_ ...string) {}
