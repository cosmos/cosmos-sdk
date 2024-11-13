package services

import (
	"github.com/prometheus/client_golang/prometheus/promauto"
	"strings"
	"time"

	"github.com/hashicorp/go-metrics"
	"github.com/prometheus/client_golang/prometheus"

	"cosmossdk.io/core/telemetry"
)

var (
	_ telemetry.Service = &GlobalTelemetryService{}
)

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

func (g *GlobalTelemetryService) IncrCounter(key []string, val float32, labels ...telemetry.Label) {
	l := make([]metrics.Label, len(labels))
	for i, label := range labels {
		l[i] = metrics.Label{Name: label.Name, Value: label.Value}
	}
	metrics.IncrCounterWithLabels(key, val, append(g.globalLabels, l...))
}

func (g *GlobalTelemetryService) RegisterMeasure([]string, ...string) {}

func (g *GlobalTelemetryService) RegisterCounter([]string, ...string) {}

type PrometheusTelemetryService struct {
	globalLabels []metrics.Label
	metrics      map[string]prometheus.Collector
}

var prometheusInst = &PrometheusTelemetryService{
	metrics: make(map[string]prometheus.Collector),
}

func NewPrometheusTelemetryService(globalLabels []telemetry.Label) *PrometheusTelemetryService {
	labels := make([]metrics.Label, len(globalLabels))
	for i, label := range globalLabels {
		labels[i] = metrics.Label{Name: label.Name, Value: label.Value}
	}
	prometheusInst.globalLabels = labels
	return prometheusInst
}

func (p *PrometheusTelemetryService) MeasureSince(start time.Time, key []string, labels ...telemetry.Label) {
	dur := time.Since(start)
	name := strings.Join(key, "_")
	m, ok := p.metrics[name]
	if !ok {
		return
	}
	h, ok := m.(*prometheus.HistogramVec)
	if !ok {
		return
	}
	ls := make(map[string]string, len(labels))
	for _, label := range labels {
		ls[label.Name] = label.Value
	}
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

func (p *PrometheusTelemetryService) RegisterMeasure(key []string, labels ...string) {
	name := strings.Join(key, "_")
	p.metrics[name] = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: name,
		Help: "Histogram for " + name,
	}, labels)
}

func (p *PrometheusTelemetryService) RegisterCounter(key []string, labels ...string) {
	name := strings.Join(key, "_")
	p.metrics[name] = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: name,
		Help: "Counter for " + name,
	}, labels)
}
