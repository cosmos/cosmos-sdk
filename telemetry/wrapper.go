package telemetry

import (
	"time"

	"github.com/armon/go-metrics"
)

// Common metric key constants
const (
	MetricKeyBeginBlocker = "begin_blocker"
	MetricKeyEndBlocker   = "end_blocker"
	MetricLabelNameModule = "module"
)

func NewLabel(name, value string) metrics.Label {
	return metrics.Label{Name: name, Value: value}
}

func ModuleMeasureSince(module string, keys ...string) {
	metrics.MeasureSinceWithLabels(
		keys,
		time.Now().UTC(),
		append([]metrics.Label{NewLabel(MetricLabelNameModule, module)}, globalLabels...),
	)
}

func ModuleSetGauge(module string, val float32, keys ...string) {
	metrics.SetGaugeWithLabels(
		keys,
		val,
		append([]metrics.Label{NewLabel(MetricLabelNameModule, module)}, globalLabels...),
	)
}

func SetGauge(val float32, keys ...string) {
	metrics.SetGaugeWithLabels(keys, val, globalLabels)
}

func MeasureSince(keys ...string) {
	metrics.MeasureSinceWithLabels(keys, time.Now().UTC(), globalLabels)
}
