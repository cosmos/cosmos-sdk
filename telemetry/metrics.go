package telemetry

import (
	"time"

	"github.com/armon/go-metrics"
	"github.com/armon/go-metrics/prometheus"
)

// Config defines the configuration options for application telemetry.
type Config struct {
	// Prefixed with keys to separate services
	ServiceName string `mapstructure:"service-name"`

	// Enabled enables the application telemetry functionality. When enabled,
	// an in-memory sink is also enabled by default. Operators may also enabled
	// other sinks such as Prometheus.
	Enabled bool `mapstructure:"enabled"`

	// Enable prefixing gauge values with hostname
	EnableHostname bool `mapstructure:"enable-hostname"`

	// Enable adding hostname to labels
	EnableHostnameLabel bool `mapstructure:"enable-hostname-label"`

	// Enable adding service to labels
	EnableServiceLabel bool `mapstructure:"enable-service-label"`

	// PrometheusRetentionTime, when positive, enables a Prometheus metrics sink.
	PrometheusRetentionTime int64 `mapstructure:"prometheus-retention-time"`
}

// Metrics defines a wrapper around application telemetry functionality. It allows
// metrics to be gathered at any point in time. When creating a Metrics object,
// internally, a global metrics is registered with a set of sinks as configured
// by the operator.
type Metrics struct {
	memSink           *metrics.InmemSink
	prometheusEnabled bool
}

func New(cfg Config) (*Metrics, error) {
	if !cfg.Enabled {
		return nil, nil
	}

	metricsConf := metrics.DefaultConfig(cfg.ServiceName)
	metricsConf.EnableHostname = cfg.EnableHostname
	metricsConf.EnableHostnameLabel = cfg.EnableHostnameLabel

	memSink := metrics.NewInmemSink(10*time.Second, time.Minute)
	metrics.DefaultInmemSignal(memSink)

	m := &Metrics{memSink: memSink}
	fanout := metrics.FanoutSink{memSink}

	if cfg.PrometheusRetentionTime > 0 {
		m.prometheusEnabled = true
		prometheusOpts := prometheus.PrometheusOpts{
			Expiration: time.Duration(cfg.PrometheusRetentionTime),
		}

		promSink, err := prometheus.NewPrometheusSinkFrom(prometheusOpts)
		if err != nil {
			return nil, err
		}

		fanout = append(fanout, promSink)
	}

	if _, err := metrics.NewGlobal(metricsConf, fanout); err != nil {
		return nil, err
	}

	return m, nil
}
