package log

import (
	"sync/atomic"
)

var otelConfigured atomic.Bool

// IsOpenTelemetryConfigured returns true if OpenTelemetry has been configured.
// This is used to determine whether to route log messages to OpenTelemetry.
func IsOpenTelemetryConfigured() bool {
	return otelConfigured.Load()
}

// SetOpenTelemetryConfigured sets whether OpenTelemetry has been configured.
// Setting this true, that indicates that global loggers should route log messages to OpenTelemetry.
// It also will indicate to automatic OpenTelemetry configuration code that OpenTelemetry has already been set up,
// so that it does not attempt to set it up again.
func SetOpenTelemetryConfigured(configured bool) {
	otelConfigured.Store(configured)
}
