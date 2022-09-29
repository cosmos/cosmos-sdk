// Package telemetry provides instrumentation tools for other modules and packages.
package telemetry

import (
	"time"

	gometrics "github.com/armon/go-metrics"
)

// Metrics defines a wrapper around application telemetry functionality. It allows
// metrics to be gathered at any point in time. When creating a Metrics object,
// internally, a global metrics is registered with a set of sinks as configured
// by the operator. In addition to the sinks, when a process gets a SIGUSR1, a
// dump of formatted recent metrics will be sent to STDERR.
type Metrics interface {

	// Gather collects all registered metrics and returns a GatherResponse where the
	// metrics are encoded depending on the type. metrics are either encoded via
	// Prometheus or JSON if in-memory.
	Gather(format string) (GatherResponse, error)

	// ModuleMeasureSince provides a short hand method for emitting a time measure
	// metric for a module with a given set of keys. If any global labels are defined,
	// they will be added to the module label.
	ModuleMeasureSince(module string, start time.Time, keys ...string)
	// ModuleSetGauge provides a short hand method for emitting a gauge metric for a
	// module with a given set of keys. If any global labels are defined, they will
	// be added to the module label.
	ModuleSetGauge(module string, val float32, keys ...string)

	// IncrCounter provides a wrapper functionality for emitting a counter metric with
	// global labels (if any).
	IncrCounter(val float32, keys ...string)

	// IncrCounterWithLabels provides a wrapper functionality for emitting a counter
	// metric with global labels (if any) along with the provided labels.
	IncrCounterWithLabels(keys []string, val float32, labels []gometrics.Label)

	// SetGauge provides a wrapper functionality for emitting a gauge metric with
	// global labels (if any).
	SetGauge(val float32, keys ...string)

	// SetGaugeWithLabels provides a wrapper functionality for emitting a gauge
	// metric with global labels (if any) along with the provided labels.
	SetGaugeWithLabels(keys []string, val float32, labels []gometrics.Label)

	// MeasureSince provides a wrapper functionality for emitting a a time measure
	// metric with global labels (if any).
	MeasureSince(start time.Time, keys ...string)
}

//go:generate mockgen -destination=./../testutil/mock/telemetry.go -package=mock -source=interface.go

// validation
var _ Metrics = &metrics{}

// Default is the default singleton object for the metrics.
var Default Metrics = &metrics{cnf: &Config{}}
