package telemetry

import (
	"time"

	gometrics "github.com/armon/go-metrics"
)

// Common metric key constants
const (
	MetricKeyBeginBlocker = "begin_blocker"
	MetricKeyEndBlocker   = "end_blocker"
	MetricLabelNameModule = "module"
)

// NewLabel creates a new instance of Label with name and value
func NewLabel(name, value string) gometrics.Label {
	return gometrics.Label{Name: name, Value: value}
}

// ModuleMeasureSince provides a short hand method for emitting a time measure
// metric for a module with a given set of keys. If any global labels are defined,
// they will be added to the module label.
func ModuleMeasureSince(module string, start time.Time, keys ...string) {
	gometrics.MeasureSinceWithLabels(
		keys,
		start.UTC(),
		append([]gometrics.Label{NewLabel(MetricLabelNameModule, module)}, globalLabels...),
	)
}

// ModuleSetGauge provides a short hand method for emitting a gauge metric for a
// module with a given set of keys. If any global labels are defined, they will
// be added to the module label.
func ModuleSetGauge(module string, val float32, keys ...string) {
	gometrics.SetGaugeWithLabels(
		keys,
		val,
		append([]gometrics.Label{NewLabel(MetricLabelNameModule, module)}, globalLabels...),
	)
}

// IncrCounter provides a wrapper functionality for emitting a counter metric with
// global labels (if any).
func IncrCounter(val float32, keys ...string) {
	gometrics.IncrCounterWithLabels(keys, val, globalLabels)
}

// IncrCounterWithLabels provides a wrapper functionality for emitting a counter
// metric with global labels (if any) along with the provided labels.
func IncrCounterWithLabels(keys []string, val float32, labels []gometrics.Label) {
	gometrics.IncrCounterWithLabels(keys, val, append(labels, globalLabels...))
}

// SetGauge provides a wrapper functionality for emitting a gauge metric with
// global labels (if any).
func SetGauge(val float32, keys ...string) {
	gometrics.SetGaugeWithLabels(keys, val, globalLabels)
}

// SetGaugeWithLabels provides a wrapper functionality for emitting a gauge
// metric with global labels (if any) along with the provided labels.
func SetGaugeWithLabels(keys []string, val float32, labels []gometrics.Label) {
	gometrics.SetGaugeWithLabels(keys, val, append(labels, globalLabels...))
}

// MeasureSince provides a wrapper functionality for emitting a a time measure
// metric with global labels (if any).
func MeasureSince(start time.Time, keys ...string) {
	gometrics.MeasureSinceWithLabels(keys, start.UTC(), globalLabels)
}
