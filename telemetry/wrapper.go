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
	metrics.MeasureSinceWithLabels(keys, time.Now().UTC(), []metrics.Label{NewLabel(MetricLabelNameModule, module)})
}
