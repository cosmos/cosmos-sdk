// Package telemetry provides instrumentation tools for other modules and packages.
package telemetry

import (
	"time"

	gometrics "github.com/armon/go-metrics"
)

// NewLabel creates a new instance of Label with name and value
func NewLabel(name, value string) gometrics.Label {
	return gometrics.Label{Name: name, Value: value}
}

// ModuleMeasureSince provides a short hand method for emitting a time measure
// metric for a module with a given set of keys. If any global labels are defined,
// they will be added to the module label.
func ModuleMeasureSince(module string, start time.Time, keys ...string) {
	Default.ModuleMeasureSince(module, start, keys...)
}

// ModuleSetGauge provides a short hand method for emitting a gauge metric for a
// module with a given set of keys. If any global labels are defined, they will
// be added to the module label.
func ModuleSetGauge(module string, val float32, keys ...string) {
	Default.ModuleSetGauge(module, val, keys...)
}

// IncrCounter provides a wrapper functionality for emitting a counter metric with
// global labels (if any).
func IncrCounter(val float32, keys ...string) {
	Default.IncrCounter(val, keys...)
}

// IncrCounterWithLabels provides a wrapper functionality for emitting a counter
// metric with global labels (if any) along with the provided labels.
func IncrCounterWithLabels(keys []string, val float32, labels []gometrics.Label) {
	Default.IncrCounterWithLabels(keys, val, labels)
}

// SetGauge provides a wrapper functionality for emitting a gauge metric with
// global labels (if any).
func SetGauge(val float32, keys ...string) {
	Default.SetGauge(val, keys...)
}

// SetGaugeWithLabels provides a wrapper functionality for emitting a gauge
// metric with global labels (if any) along with the provided labels.
func SetGaugeWithLabels(keys []string, val float32, labels []gometrics.Label) {
	Default.SetGaugeWithLabels(keys, val, labels)
}

// MeasureSince provides a wrapper functionality for emitting a a time measure
// metric with global labels (if any).
func MeasureSince(start time.Time, keys ...string) {
	Default.MeasureSince(start, keys...)
}
