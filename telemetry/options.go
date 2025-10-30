package telemetry

import "cosmossdk.io/log"

// Option is a functional option for configuring Metrics.
type Option func(*Metrics)

// WithLogger sets the default logger to use for tracers.
// If not provided, a nop logger will be used.
func WithLogger(logger log.Logger) Option {
	return func(m *Metrics) {
		m.logger = logger
	}
}
