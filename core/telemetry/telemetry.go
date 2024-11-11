package telemetry

import (
	"time"

	"cosmossdk.io/core/server"
)

type Service interface {
	// MeasureSince emits a time measure metric with the provided keys.
	MeasureSince(start time.Time, key []string, labels ...Label)
}

type ServiceFactory func(server.ConfigMap) Service

type Label struct {
	Name  string
	Value string
}
