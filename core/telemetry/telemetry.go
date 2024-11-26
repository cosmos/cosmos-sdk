package telemetry

import (
	"time"

	"cosmossdk.io/core/server"
)

type Service interface {
	// MeasureSince emits a time measure metric with the provided keys.
	MeasureSince(start time.Time, key []string, labels ...Label)

	IncrCounter(key []string, val float32, labels ...Label)

	RegisterHistogram(key []string, buckets []float64, labels ...string)

	RegisterSummary(key []string, labels ...string)

	RegisterCounter(key []string, labels ...string)
}

type ServiceFactory func(server.ConfigMap) Service

type Label struct {
	Name  string
	Value string
}

type buckets struct {
	Default       []float64
	StoreOpsOrder []float64
}

var Buckets = buckets{
	Default:       []float64{0.5e-6, 1e-6, 1e-5, .0005, .025, .1, .5, 1, 5, 10, 30, 120},
	StoreOpsOrder: []float64{0.1e-6, 1e-6, 2.5e-6, 5e-6, 10e-6, 50e-6, 100e-6, 500e-6, 1e-3, 5e-3, 50e-3, 250e-3, 500e-3, 1},
}

var _ Service = &NopService{}

// NopService is a no-op implementation of telemetry Service.
type NopService struct{}

func (n *NopService) IncrCounter(key []string, val float32, labels ...Label) {}

func (n *NopService) MeasureSince(start time.Time, key []string, labels ...Label) {}

func (n *NopService) RegisterCounter(key []string, labels ...string) {}

func (n *NopService) RegisterHistogram(key []string, buckets []float64, labels ...string) {}

func (n *NopService) RegisterSummary(key []string, labels ...string) {}
