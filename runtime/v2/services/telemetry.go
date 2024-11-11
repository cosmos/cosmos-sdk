package services

import (
	"time"

	"github.com/hashicorp/go-metrics"

	"cosmossdk.io/core/telemetry"
)

var _ telemetry.Service = &GlobalTelemetryService{}

type GlobalTelemetryService struct {
	globalLabels []metrics.Label
}

func NewGlobalTelemetryService(globalLabels []telemetry.Label) *GlobalTelemetryService {
	labels := make([]metrics.Label, len(globalLabels))
	for i, label := range globalLabels {
		labels[i] = metrics.Label{Name: label.Name, Value: label.Value}
	}
	return &GlobalTelemetryService{
		globalLabels: labels,
	}
}

// MeasureSince implements telemetry.Service.
func (g *GlobalTelemetryService) MeasureSince(start time.Time, key []string, labels ...telemetry.Label) {
	l := make([]metrics.Label, len(labels))
	for i, label := range labels {
		l[i] = metrics.Label{Name: label.Name, Value: label.Value}
	}
	metrics.MeasureSinceWithLabels(key, start.UTC(), append(g.globalLabels, l...))
}
