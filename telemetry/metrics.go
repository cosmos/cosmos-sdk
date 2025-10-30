// Package telemetry provides observability through metrics and distributed tracing.
//
// # Metrics Collection
//
// Metrics collection uses hashicorp/go-metrics with support for multiple sink backends:
//   - mem: In-memory aggregation with SIGUSR1 signal dumping to stderr
//   - prometheus: Prometheus registry for pull-based scraping via /metrics endpoint
//   - statsd: Push-based metrics to StatsD daemon
//   - dogstatsd: Push-based metrics to Datadog StatsD daemon with tagging
//   - file: Write metrics to a file as JSON lines (useful for tests and debugging)
//
// Multiple sinks can be active simultaneously via FanoutSink (e.g., both in-memory and Prometheus).
//
// # Distributed Tracing
//
// Tracing support is provided via OtelSpan, which wraps OpenTelemetry for hierarchical span tracking.
// See otel.go for the log.Tracer implementation.
//
// # Usage
//
// Initialize metrics at application startup:
//
//	m, err := telemetry.New(telemetry.Config{
//		Enabled:                 true,
//		ServiceName:             "cosmos-app",
//		PrometheusRetentionTime: 60,
//		GlobalLabels:            [][]string{{"chain_id", "cosmoshub-1"}},
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer m.Close()
//
// Emit metrics from anywhere in the application:
//
//	telemetry.IncrCounter(1, "tx", "processed")
//	telemetry.SetGauge(1024, "mempool", "size")
//	defer telemetry.MeasureSince(telemetry.Now(), "block", "execution")
package telemetry

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/hashicorp/go-metrics"
	"github.com/hashicorp/go-metrics/datadog"
	metricsprom "github.com/hashicorp/go-metrics/prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
)

// globalTelemetryEnabled is a private variable that stores the telemetry enabled state.
// It is set on initialization and does not change for the lifetime of the program.
var globalTelemetryEnabled bool

// IsTelemetryEnabled provides controlled access to check if telemetry is enabled.
func IsTelemetryEnabled() bool {
	return globalTelemetryEnabled
}

// EnableTelemetry allows for the global telemetry enabled state to be set.
func EnableTelemetry() {
	globalTelemetryEnabled = true
}

// globalLabels defines the set of global labels that will be applied to all
// metrics emitted using the telemetry package function wrappers.
var globalLabels = []metrics.Label{}

// Metrics supported format types.
const (
	FormatDefault    = ""
	FormatPrometheus = "prometheus"
	FormatText       = "text"
	ContentTypeText  = `text/plain; version=` + expfmt.TextVersion + `; charset=utf-8`

	MetricSinkInMem      = "mem"
	MetricSinkStatsd     = "statsd"
	MetricSinkDogsStatsd = "dogstatsd"
	MetricSinkFile       = "file"
)

// DisplayableSink is an interface that defines a method for displaying metrics.
type DisplayableSink interface {
	DisplayMetrics(resp http.ResponseWriter, req *http.Request) (any, error)
}

// Config defines the configuration options for application telemetry.
type Config struct {
	// ServiceName is the identifier for this service, used as a prefix for all metric keys.
	// Example: "cosmos-app" → metrics like "cosmos-app.tx.count"
	ServiceName string `mapstructure:"service-name"`

	// Enabled controls whether telemetry is active. When false, all telemetry operations
	// become no-ops with zero overhead. When true, metrics collection is activated.
	Enabled bool `mapstructure:"enabled"`

	// EnableHostname prefixes gauge values with the hostname.
	// Useful in multi-node deployments to identify which node emitted a metric.
	EnableHostname bool `mapstructure:"enable-hostname"`

	// EnableHostnameLabel adds a "hostname" label to all metrics.
	// Alternative to EnableHostname that works better with label-based systems like Prometheus.
	EnableHostnameLabel bool `mapstructure:"enable-hostname-label"`

	// EnableServiceLabel adds a "service" label with the ServiceName to all metrics.
	// Useful when aggregating metrics from multiple services in one monitoring system.
	EnableServiceLabel bool `mapstructure:"enable-service-label"`

	// PrometheusRetentionTime, when positive, enables a Prometheus metrics sink.
	// Defines how long (in seconds) metrics are retained in memory for scraping.
	// The Prometheus sink is added to a FanoutSink alongside the primary sink.
	// Recommended value: 60 seconds or more.
	PrometheusRetentionTime int64 `mapstructure:"prometheus-retention-time"`

	// GlobalLabels defines a set of key-value label pairs applied to ALL metrics.
	// These labels are automatically attached to every metric emission.
	// Useful for static identifiers like chain ID, environment, region, etc.
	//
	// Example: [][]string{{"chain_id", "cosmoshub-1"}, {"env", "production"}}
	//
	// Note: The outer array contains label pairs, each inner array has exactly 2 elements [key, value].
	GlobalLabels [][]string `mapstructure:"global-labels"`

	// MetricsSink defines the metrics backend type. Supported values:
	//   - "mem" (default): In-memory sink with SIGUSR1 dump-to-stderr capability
	//   - "prometheus": Prometheus exposition format (use with PrometheusRetentionTime)
	//   - "statsd": StatsD protocol (push-based, requires StatsdAddr)
	//   - "dogstatsd": Datadog-enhanced StatsD with tags (requires StatsdAddr, DatadogHostname)
	//   - "file": JSON lines written to a file (requires MetricsFile)
	//
	// Multiple sinks can be active via FanoutSink (e.g., mem + prometheus).
	MetricsSink string `mapstructure:"metrics-sink" default:"mem"`

	// StatsdAddr is the address of the StatsD or DogStatsD server (host:port).
	// Only used when MetricsSink is "statsd" or "dogstatsd".
	// Example: "localhost:8125"
	StatsdAddr string `mapstructure:"statsd-addr"`

	// DatadogHostname is the hostname to report when using DogStatsD.
	// Only used when MetricsSink is "dogstatsd".
	// If empty, the system hostname is used.
	DatadogHostname string `mapstructure:"datadog-hostname"`

	// MetricsFile is the file path to write metrics to in JSONL format.
	// Only used when MetricsSink is "file".
	// Each metric emission creates a JSON line: {"timestamp":"...","type":"counter","key":[...],"value":1.0}
	// Example: "/tmp/metrics.jsonl" or "./metrics.jsonl"
	MetricsFile string `mapstructure:"metrics-file"`

	TraceSink string `mapstructure:"trace-sink"`
}

// Metrics provides access to the application's metrics collection system.
// It wraps the go-metrics global registry and configured sinks.
//
// When using the in-memory sink, sending SIGUSR1 to the process (kill -USR1 <pid>)
// will dump current metrics to stderr for debugging.
//
// The Metrics object maintains references to configured sinks and provides
// a Gather() method for pull-based metric retrieval (useful for testing and monitoring).
//
// When using the file sink, call Close() to flush buffered data and close the file.
//
// Note: go-metrics uses a singleton global registry. Only one Metrics instance
// should be created per process.
type Metrics struct {
	sink              metrics.MetricSink
	prometheusEnabled bool
	closer            io.Closer // non-nil when using file sink
}

// GatherResponse contains collected metrics in the requested format.
// The Metrics field holds the serialized metric data, and ContentType
// indicates how it's encoded ("application/json" or prometheus text format).
type GatherResponse struct {
	Metrics     []byte
	ContentType string
}

// New creates and initializes the metrics system with the given configuration.
//
// Returns nil if telemetry is disabled (cfg.Enabled == false), which allows
// callers to safely ignore the Metrics object.
//
// The function:
//   - Initializes the go-metrics global registry
//   - Configures the specified sink(s) (mem, prometheus, statsd, dogstatsd, file)
//   - Sets up global labels to be applied to all metrics
//   - Enables SIGUSR1 signal handling for in-memory sink dumps
//   - Creates a FanoutSink if multiple sinks are needed (e.g., mem + prometheus)
//
// Example:
//
//	m, err := telemetry.New(telemetry.Config{
//		Enabled:                 true,
//		ServiceName:             "cosmos-app",
//		MetricsSink:             telemetry.MetricSinkInMem,
//		PrometheusRetentionTime: 60,
//	})
//	if err != nil {
//		return err
//	}
//	defer m.Close()
func New(cfg Config) (_ *Metrics, rerr error) {
	globalTelemetryEnabled = cfg.Enabled
	if !cfg.Enabled {
		return nil, nil
	}

	if numGlobalLabels := len(cfg.GlobalLabels); numGlobalLabels > 0 {
		parsedGlobalLabels := make([]metrics.Label, numGlobalLabels)
		for i, gl := range cfg.GlobalLabels {
			parsedGlobalLabels[i] = NewLabel(gl[0], gl[1])
		}
		globalLabels = parsedGlobalLabels
	}

	metricsConf := metrics.DefaultConfig(cfg.ServiceName)
	metricsConf.EnableHostname = cfg.EnableHostname
	metricsConf.EnableHostnameLabel = cfg.EnableHostnameLabel

	var (
		sink   metrics.MetricSink
		closer io.Closer
		err    error
	)
	switch cfg.MetricsSink {
	case MetricSinkStatsd:
		sink, err = metrics.NewStatsdSink(cfg.StatsdAddr)
	case MetricSinkDogsStatsd:
		sink, err = datadog.NewDogStatsdSink(cfg.StatsdAddr, cfg.DatadogHostname)
	case MetricSinkFile:
		if cfg.MetricsFile == "" {
			return nil, errors.New("metrics-file must be set when metrics-sink is 'file'")
		}
		file, err := os.OpenFile(cfg.MetricsFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open metrics file: %w", err)
		}
		fileSink := NewFileSink(file)
		sink = fileSink
		closer = fileSink
	default:
		memSink := metrics.NewInmemSink(10*time.Second, time.Minute)
		sink = memSink
		inMemSig := metrics.DefaultInmemSignal(memSink)
		defer func() {
			if rerr != nil {
				inMemSig.Stop()
			}
		}()
	}

	if err != nil {
		return nil, err
	}

	m := &Metrics{sink: sink, closer: closer}
	fanout := metrics.FanoutSink{sink}

	if cfg.PrometheusRetentionTime > 0 {
		m.prometheusEnabled = true
		prometheusOpts := metricsprom.PrometheusOpts{
			Expiration: time.Duration(cfg.PrometheusRetentionTime) * time.Second,
		}

		promSink, err := metricsprom.NewPrometheusSinkFrom(prometheusOpts)
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

// Gather collects all registered metrics and returns a GatherResponse where the
// metrics are encoded depending on the type. Metrics are either encoded via
// Prometheus or JSON if in-memory.
func (m *Metrics) Gather(format string) (GatherResponse, error) {
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

// gatherPrometheus collects Prometheus metrics and returns a GatherResponse.
// If Prometheus metrics are not enabled, it returns an error.
func (m *Metrics) gatherPrometheus() (GatherResponse, error) {
	if !m.prometheusEnabled {
		return GatherResponse{}, errors.New("prometheus metrics are not enabled")
	}

	metricsFamilies, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		return GatherResponse{}, fmt.Errorf("failed to gather prometheus metrics: %w", err)
	}

	buf := &bytes.Buffer{}
	defer buf.Reset()

	e := expfmt.NewEncoder(buf, expfmt.NewFormat(expfmt.TypeTextPlain))

	for _, mf := range metricsFamilies {
		if err := e.Encode(mf); err != nil {
			return GatherResponse{}, fmt.Errorf("failed to encode prometheus metrics: %w", err)
		}
	}

	return GatherResponse{ContentType: ContentTypeText, Metrics: buf.Bytes()}, nil
}

// gatherGeneric collects generic metrics and returns a GatherResponse.
func (m *Metrics) gatherGeneric() (GatherResponse, error) {
	gm, ok := m.sink.(DisplayableSink)
	if !ok {
		return GatherResponse{}, errors.New("non in-memory metrics sink does not support generic format")
	}

	summary, err := gm.DisplayMetrics(nil, nil)
	if err != nil {
		return GatherResponse{}, fmt.Errorf("failed to gather in-memory metrics: %w", err)
	}

	content, err := json.Marshal(summary)
	if err != nil {
		return GatherResponse{}, fmt.Errorf("failed to encode in-memory metrics: %w", err)
	}

	return GatherResponse{ContentType: "application/json", Metrics: content}, nil
}

// Close flushes any buffered data and closes resources associated with the metrics system.
// This is primarily needed when using the file sink to ensure all data is written to disk.
// It is safe to call Close() multiple times, and safe to call on a nil Metrics object.
//
// For other sink types (mem, statsd, prometheus), Close() is a no-op.
//
// Example:
//
//	m, err := telemetry.New(cfg)
//	if err != nil {
//		return err
//	}
//	defer m.Close()
func (m *Metrics) Close() error {
	if m == nil || m.closer == nil {
		return nil
	}
	return m.closer.Close()
}
