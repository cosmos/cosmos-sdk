package services

import (
	"strings"
	"time"

	"github.com/hashicorp/go-metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"cosmossdk.io/core/telemetry"
)

var (
	_ telemetry.Service = &GlobalTelemetryService{}
	_ telemetry.Service = &PrometheusTelemetryService{}
)

type GlobalTelemetryService struct{}

// RegisterHistogram implements telemetry.Service.
func (g *GlobalTelemetryService) RegisterHistogram(key []string, buckets []float64, labels ...string) {
	panic("unimplemented")
}

// RegisterSummary implements telemetry.Service.
func (g *GlobalTelemetryService) RegisterSummary(key []string, labels ...string) {
	panic("unimplemented")
}

func NewGlobalTelemetryService() *GlobalTelemetryService {
	return &GlobalTelemetryService{}
}

// MeasureSince implements telemetry.Service.
func (g *GlobalTelemetryService) MeasureSince(start time.Time, key []string, labels ...telemetry.Label) {
	ls := make([]metrics.Label, len(labels))
	for i, label := range labels {
		ls[i] = metrics.Label{Name: label.Name, Value: label.Value}
	}
	metrics.MeasureSinceWithLabels(key, start.UTC(), ls)
}

func (g *GlobalTelemetryService) IncrCounter(key []string, val float32, labels ...telemetry.Label) {
	ls := make([]metrics.Label, len(labels))
	for i, label := range labels {
		ls[i] = metrics.Label{Name: label.Name, Value: label.Value}
	}
	metrics.IncrCounterWithLabels(key, val, ls)
}

func (g *GlobalTelemetryService) RegisterMeasure([]string, ...string) {}

func (g *GlobalTelemetryService) RegisterCounter([]string, ...string) {}

type PrometheusTelemetryService struct {
	metrics map[string]prometheus.Collector
}

var prometheusInst = &PrometheusTelemetryService{
	metrics: make(map[string]prometheus.Collector),
}

func NewPrometheusTelemetryService() *PrometheusTelemetryService {
	return prometheusInst
}

type labeledOberserver interface {
	With(labels prometheus.Labels) prometheus.Observer
}

func (p *PrometheusTelemetryService) MeasureSince(start time.Time, key []string, labels ...telemetry.Label) {
	dur := time.Since(start)
	name := strings.Join(key, "_")
	m, ok := p.metrics[name]
	if !ok {
		return
	}

	h, ok := m.(labeledOberserver)
	if !ok {
		return
	}
	ls := make(map[string]string, len(labels))
	for _, label := range labels {
		ls[label.Name] = label.Value
	}
	// fmt.Printf("MeasureSince: %s, %v, %v\n", name, ls, dur.Seconds())
	h.With(ls).Observe(dur.Seconds())
}

func (p *PrometheusTelemetryService) IncrCounter(key []string, val float32, labels ...telemetry.Label) {
	name := strings.Join(key, "_")
	m, ok := p.metrics[name]
	if !ok {
		return
	}
	c, ok := m.(*prometheus.CounterVec)
	if !ok {
		return
	}
	ls := make(map[string]string, len(labels))
	for _, label := range labels {
		ls[label.Name] = label.Value
	}
	c.With(ls).Add(float64(val))
}

func (p *PrometheusTelemetryService) RegisterCounter(key []string, labels ...string) {
	name := strings.Join(key, "_")
	p.metrics[name] = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: name,
	}, labels)
}

func (p *PrometheusTelemetryService) RegisterHistogram(key []string, buckets []float64, labels ...string) {
	name := strings.Join(key, "_")
	p.metrics[name] = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    name,
		Buckets: buckets,
	}, labels)
}

func (p *PrometheusTelemetryService) RegisterSummary(key []string, labels ...string) {
	name := strings.Join(key, "_")
	p.metrics[name] = promauto.NewSummaryVec(prometheus.SummaryOpts{
		Name:       name,
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
	}, labels)
}
