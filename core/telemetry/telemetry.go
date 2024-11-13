package telemetry

import (
	"time"

	"cosmossdk.io/core/server"
)

type Service interface {
	// MeasureSince emits a time measure metric with the provided keys.
	MeasureSince(start time.Time, key []string, labels ...Label)

	IncrCounter(key []string, val float32, labels ...Label)

	RegisterMeasure(key []string, labels ...string)

	RegisterCounter(key []string, labels ...string)
}

type ServiceFactory func(server.ConfigMap) Service

type Label struct {
	Name  string
	Value string
}
