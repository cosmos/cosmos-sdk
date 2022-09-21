package telemetry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	gometrics "github.com/armon/go-metrics"
	metricsprom "github.com/armon/go-metrics/prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
)

// ─── Types ──────────────────────────────────────────────────────────────────────

// metrics defines a wrapper around application telemetry functionality. It allows
// metrics to be gathered at any point in time. When creating a Metrics object,
// internally, a global metrics is registered with a set of sinks as configured
// by the operator. In addition to the sinks, when a process gets a SIGUSR1, a
// dump of formatted recent metrics will be sent to STDERR.
type metrics struct {
	memSink *gometrics.InmemSink
	gMetric *gometrics.Metrics

	prometheusEnabled bool
	//cnf can be use in runtime process
	cnf *Config
}

// GatherResponse is the response type of registered metrics
type GatherResponse struct {
	Metrics     []byte
	ContentType string
}

// ─── Factory ────────────────────────────────────────────────────────────────────

// New creates a new instance of metrics
func New(opts ...Option) (Metrics, error) {
	var err error

	c := new(Config)
	c.loadDefaults()
	for _, fn := range opts {
		err = fn(c)
		if err != nil {
			return nil, err
		}
	}

	if !c.Enabled {
		return nil, nil
	}

	if numGlobalLables := len(c.GlobalLabels); numGlobalLables > 0 {
		parsedGlobalLabels := make([]gometrics.Label, numGlobalLables)
		for i, gl := range c.GlobalLabels {
			parsedGlobalLabels[i] = NewLabel(gl[0], gl[1])
		}

		c.globalLabels = parsedGlobalLabels
	}

	metricsConf := gometrics.DefaultConfig(c.ServiceName)
	metricsConf.EnableHostname = c.EnableHostname
	metricsConf.EnableHostnameLabel = c.EnableHostnameLabel

	memSink := gometrics.NewInmemSink(10*time.Second, time.Minute)
	gometrics.DefaultInmemSignal(memSink)

	m := &metrics{memSink: memSink, cnf: c}
	fanout := gometrics.FanoutSink{memSink}
	var promSink *metricsprom.PrometheusSink

	//Initialize the external library only once.
	initOnce.Do(func() {
		if c.PrometheusRetentionTime > 0 {
			m.prometheusEnabled = true
			prometheusOpts := metricsprom.PrometheusOpts{
				Expiration: c.PrometheusRetentionTime,
			}

			promSink, err = metricsprom.NewPrometheusSinkFrom(prometheusOpts)

		}
	})
	if err != nil {
		return nil, err
	}

	fanout = append(fanout, promSink)

	if m.cnf.useGlobalMetricRegistration {
		m.gMetric, err = gometrics.NewGlobal(metricsConf, fanout)
	} else {
		m.gMetric, err = gometrics.New(metricsConf, fanout)
	}

	if err != nil {
		return nil, err
	}
	return m, nil
}

// Init create a new instance and set it as the default Metric
func Init(opts ...Option) error {
	m, err := New(opts...)
	if err != nil {
		return err
	}
	Default = m
	return nil
}

// ─── Functions ──────────────────────────────────────────────────────────────────

// Gather collects all registered metrics and returns a GatherResponse where the
// metrics are encoded depending on the type. metrics are either encoded via
// Prometheus or JSON if in-memory.
func (m *metrics) Gather(format string) (GatherResponse, error) {
	switch format {
	case FormatPrometheus:
		return m.gatherPrometheus()

	case FormatText:
		return m.gatherGeneric()

	case FormatDefault:
		return m.gatherGeneric()

	default:
		return GatherResponse{}, fmt.Errorf("unsupported metrics format: %s", format)
	}
}

// ModuleMeasureSince provides a short hand method for emitting a time measure
// metric for a module with a given set of keys. If any global labels are defined,
// they will be added to the module label.
func (m *metrics) ModuleMeasureSince(module string, start time.Time, keys ...string) {
	gometrics.MeasureSinceWithLabels(
		keys,
		start.UTC(),
		append([]gometrics.Label{NewLabel(MetricLabelNameModule, module)}, m.cnf.globalLabels...),
	)
}

// ModuleSetGauge provides a short hand method for emitting a gauge metric for a
// module with a given set of keys. If any global labels are defined, they will
// be added to the module label.
func (m *metrics) ModuleSetGauge(module string, val float32, keys ...string) {
	gometrics.SetGaugeWithLabels(
		keys,
		val,
		append([]gometrics.Label{NewLabel(MetricLabelNameModule, module)}, m.cnf.globalLabels...),
	)
}

// IncrCounter provides a wrapper functionality for emitting a counter metric with
// global labels (if any).
func (m *metrics) IncrCounter(val float32, keys ...string) {
	gometrics.IncrCounterWithLabels(keys, val, m.cnf.globalLabels)
}

// IncrCounterWithLabels provides a wrapper functionality for emitting a counter
// metric with global labels (if any) along with the provided labels.
func (m *metrics) IncrCounterWithLabels(keys []string, val float32, labels []gometrics.Label) {
	gometrics.IncrCounterWithLabels(keys, val, append(labels, m.cnf.globalLabels...))
}

// SetGauge provides a wrapper functionality for emitting a gauge metric with
// global labels (if any).
func (m *metrics) SetGauge(val float32, keys ...string) {
	gometrics.SetGaugeWithLabels(keys, val, m.cnf.globalLabels)
}

// SetGaugeWithLabels provides a wrapper functionality for emitting a gauge
// metric with global labels (if any) along with the provided labels.
func (m *metrics) SetGaugeWithLabels(keys []string, val float32, labels []gometrics.Label) {
	gometrics.SetGaugeWithLabels(keys, val, append(labels, m.cnf.globalLabels...))
}

// MeasureSince provides a wrapper functionality for emitting a a time measure
// metric with global labels (if any).
func (m *metrics) MeasureSince(start time.Time, keys ...string) {
	gometrics.MeasureSinceWithLabels(keys, start.UTC(), m.cnf.globalLabels)
}

// ─── Utils ──────────────────────────────────────────────────────────────────────

func (m *metrics) gatherPrometheus() (GatherResponse, error) {
	if !m.prometheusEnabled {
		return GatherResponse{}, fmt.Errorf("prometheus metrics are not enabled")
	}

	metricsFamilies, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		return GatherResponse{}, fmt.Errorf("failed to gather prometheus metrics: %w", err)
	}

	buf := &bytes.Buffer{}
	defer buf.Reset()

	e := expfmt.NewEncoder(buf, expfmt.FmtText)
	for _, mf := range metricsFamilies {
		if err := e.Encode(mf); err != nil {
			return GatherResponse{}, fmt.Errorf("failed to encode prometheus metrics: %w", err)
		}
	}

	return GatherResponse{ContentType: string(expfmt.FmtText), Metrics: buf.Bytes()}, nil
}

func (m *metrics) gatherGeneric() (GatherResponse, error) {
	summary, err := m.memSink.DisplayMetrics(nil, nil)
	if err != nil {
		return GatherResponse{}, fmt.Errorf("failed to gather in-memory metrics: %w", err)
	}

	content, err := json.Marshal(summary)
	if err != nil {
		return GatherResponse{}, fmt.Errorf("failed to encode in-memory metrics: %w", err)
	}

	return GatherResponse{ContentType: "application/json", Metrics: content}, nil
}
