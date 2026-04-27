package metrics

import (
	"sync"
	"time"
)

// MetricType represents the type of a metric.
type MetricType int

const (
	Counter   MetricType = iota // Counter is a monotonically increasing value.
	Gauge                       // Gauge is a value that can go up and down.
	Histogram                   // Histogram tracks value distributions.
)

// String returns the string representation of the MetricType.
func (mt MetricType) String() string {
	switch mt {
	case Counter:
		return "counter"
	case Gauge:
		return "gauge"
	case Histogram:
		return "histogram"
	default:
		return "unknown"
	}
}

// Metric represents a single metric data point.
type Metric struct {
	Name      string
	Type      MetricType
	Value     float64
	Labels    map[string]string
	Timestamp time.Time
}

// MetricFamily groups related metrics under a common name and type.
type MetricFamily struct {
	Name    string
	Help    string
	Type    MetricType
	Metrics []Metric
}

// MetricsConfig holds configuration for the metrics subsystem.
type MetricsConfig struct {
	EnablePrometheus   bool
	Namespace          string
	Subsystem          string
	CollectionInterval time.Duration
}

// DefaultMetricsConfig returns a MetricsConfig with sensible defaults.
func DefaultMetricsConfig() MetricsConfig {
	return MetricsConfig{
		EnablePrometheus:   true,
		Namespace:          "cosmos",
		Subsystem:          "sdk",
		CollectionInterval: 10 * time.Second,
	}
}

// MetricCollector defines the interface for collecting metrics.
type MetricCollector interface {
	Collect() []Metric
}

// MetricRegistry holds registered metric collectors.
type MetricRegistry struct {
	mu         sync.RWMutex
	collectors map[string]MetricCollector
}

// NewMetricRegistry creates a new MetricRegistry.
func NewMetricRegistry() *MetricRegistry {
	return &MetricRegistry{
		collectors: make(map[string]MetricCollector),
	}
}

// Register adds a collector to the registry under the given name.
func (r *MetricRegistry) Register(name string, collector MetricCollector) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.collectors[name] = collector
}

// Unregister removes a collector from the registry.
func (r *MetricRegistry) Unregister(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.collectors, name)
}

// CollectAll gathers metrics from all registered collectors.
func (r *MetricRegistry) CollectAll() []Metric {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var all []Metric
	for _, c := range r.collectors {
		all = append(all, c.Collect()...)
	}
	return all
}
