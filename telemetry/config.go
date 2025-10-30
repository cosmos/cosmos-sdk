package telemetry

import (
	"net/http"

	"github.com/hashicorp/go-metrics"
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

	TraceSinkOtel = "otel"
	TraceSinkNoop = "noop"
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

	// TraceSink is the sink for trace data. Supported values:
	// - "otel": OpenTelemetry trace sink
	// - "metrics": all spans will be redirected to emit invocation counter and timing histogram metrics
	// - "noop": No-op trace sink (default)
	TraceSink string `mapstructure:"trace-sink"`

	// OtelTraceExporters is a list of OTLP exporters to use for trace data.
	// This is only used when trace sink is set to "otel".
	OtelTraceExporters []OtelTraceExportConfig `mapstructure:"otel-trace-exporters"`
}

type OtelTraceExportConfig struct {
	// Type is the exporter type.
	// Must be one of:
	//   - "stdout"
	//   - "otlp"
	//
	// OTLP exporters must set the endpoint URL and can optionally set the transport protocol.
	Type string `mapstructure:"type"`

	// OTLPTransport is the transport protocol to use for OTLP.
	// Must be one of:
	// 	- "http" (default)
	//  - "grpc"
	OTLPTransport string `mapstructure:"otlp-transport"`

	// Endpoint is the OTLP exporter endpoint URL (grpc or http).
	Endpoint string `mapstructure:"endpoint"`

	// Insecure disables TLS certificate verification for OTLP exporters.
	Insecure bool `mapstructure:"insecure"`

	// File is the file path to write trace data to when using the "stdout" exporter.
	// If it is empty, the trace data is written to stdout.
	File string `mapstructure:"file"`

	// PrettyPrint enables pretty-printing of JSON output when using the "stdout" exporter.
	PrettyPrint bool `mapstructure:"pretty-print"`
}
