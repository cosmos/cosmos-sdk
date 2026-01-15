package diskio

import (
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

// config contains optional settings for reporting disk I/O metrics.
type config struct {
	// MinimumReadInterval sets the minimum interval between calls to
	// disk.IOCounters(). Negative values are ignored.
	MinimumReadInterval time.Duration

	// MeterProvider sets the metric.MeterProvider. If nil, the global
	// Provider will be used.
	MeterProvider metric.MeterProvider
}

// Option supports configuring optional settings for disk I/O metrics.
type Option interface {
	apply(*config)
}

// DefaultMinimumReadInterval is the default minimum interval between calls to
// disk.IOCounters().
const DefaultMinimumReadInterval time.Duration = 15 * time.Second

// WithMinimumReadInterval sets a minimum interval between calls to
// disk.IOCounters(), which involves system calls. This setting is ignored
// when d is negative.
func WithMinimumReadInterval(d time.Duration) Option {
	return minimumReadIntervalOption(d)
}

type minimumReadIntervalOption time.Duration

func (o minimumReadIntervalOption) apply(c *config) {
	if o >= 0 {
		c.MinimumReadInterval = time.Duration(o)
	}
}

// WithMeterProvider sets the metric.MeterProvider to use for reporting.
// If this option is not used, the global metric.MeterProvider will be used.
func WithMeterProvider(provider metric.MeterProvider) Option {
	return metricProviderOption{provider}
}

type metricProviderOption struct{ metric.MeterProvider }

func (o metricProviderOption) apply(c *config) {
	if o.MeterProvider != nil {
		c.MeterProvider = o.MeterProvider
	}
}

func newConfig(opts ...Option) config {
	c := config{
		MeterProvider: otel.GetMeterProvider(),
	}
	for _, opt := range opts {
		opt.apply(&c)
	}
	if c.MinimumReadInterval <= 0 {
		c.MinimumReadInterval = DefaultMinimumReadInterval
	}
	return c
}
