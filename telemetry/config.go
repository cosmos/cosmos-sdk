package telemetry

import (
	"net/http"

	"github.com/hashicorp/go-metrics"
	"github.com/prometheus/common/expfmt"
	otelconf "go.opentelemetry.io/contrib/otelconf/v0.3.0"
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
	ServiceName string `mapstructure:"service-name" json:"service-name"`

	// Enabled controls whether telemetry is active. When false, all telemetry operations
	// become no-ops with zero overhead. When true, metrics collection is activated.
	Enabled bool `mapstructure:"enabled" json:"enabled"`

	// EnableHostname prefixes gauge values with the hostname.
	// Useful in multi-node deployments to identify which node emitted a metric.
	EnableHostname bool `mapstructure:"enable-hostname" json:"enable-hostname"`

	// EnableHostnameLabel adds a "hostname" label to all metrics.
	// Alternative to EnableHostname that works better with label-based systems like Prometheus.
	EnableHostnameLabel bool `mapstructure:"enable-hostname-label" json:"enable-hostname-label"`

	// EnableServiceLabel adds a "service" label with the ServiceName to all metrics.
	// Useful when aggregating metrics from multiple services in one monitoring system.
	EnableServiceLabel bool `mapstructure:"enable-service-label" json:"enable-service-label"`

	// PrometheusRetentionTime, when positive, enables a Prometheus metrics sink.
	// Defines how long (in seconds) metrics are retained in memory for scraping.
	// The Prometheus sink is added to a FanoutSink alongside the primary sink.
	// Recommended value: 60 seconds or more.
	PrometheusRetentionTime int64 `mapstructure:"prometheus-retention-time" json:"prometheus-retention-time"`

	// GlobalLabels defines a set of key-value label pairs applied to ALL metrics.
	// These labels are automatically attached to every metric emission.
	// Useful for static identifiers like chain ID, environment, region, etc.
	//
	// Example: [][]string{{"chain_id", "cosmoshub-1"}, {"env", "production"}}
	//
	// Note: The outer array contains label pairs, each inner array has exactly 2 elements [key, value].
	GlobalLabels [][]string `mapstructure:"global-labels" json:"global-labels"`

	// MetricsSink defines the metrics backend type. Supported values:
	//   - "mem" (default): In-memory sink with SIGUSR1 dump-to-stderr capability
	//   - "prometheus": Prometheus exposition format (use with PrometheusRetentionTime)
	//   - "statsd": StatsD protocol (push-based, requires StatsdAddr)
	//   - "dogstatsd": Datadog-enhanced StatsD with tags (requires StatsdAddr, DatadogHostname)
	//   - "file": JSON lines written to a file (requires MetricsFile)
	//
	// Multiple sinks can be active via FanoutSink (e.g., mem + prometheus).
	MetricsSink string `mapstructure:"metrics-sink" json:"metrics-sink" default:"mem"`

	// StatsdAddr is the address of the StatsD or DogStatsD server (host:port).
	// Only used when MetricsSink is "statsd" or "dogstatsd".
	// Example: "localhost:8125"
	StatsdAddr string `mapstructure:"statsd-addr" json:"statsd-addr"`

	// DatadogHostname is the hostname to report when using DogStatsD.
	// Only used when MetricsSink is "dogstatsd".
	// If empty, the system hostname is used.
	DatadogHostname string `mapstructure:"datadog-hostname" json:"datadog-hostname"`

	// MetricsFile is the file path to write metrics to in JSONL format.
	// Only used when MetricsSink is "file".
	// Each metric emission creates a JSON line: {"timestamp":"...","type":"counter","key":[...],"value":1.0}
	// Example: "/tmp/metrics.jsonl" or "./metrics.jsonl"
	MetricsFile string `mapstructure:"metrics-file" json:"metrics-file"`

	// TraceSink is the sink for trace data. Supported values:
	// - "otel": OpenTelemetry trace sink
	// - "metrics": all spans will be redirected to emit invocation counter and timing histogram metrics
	// - "noop": No-op trace sink (default)
	TraceSink string `mapstructure:"trace-sink" json:"trace-sink"`

	// OtelTraceExporters is a list of exporters to use for trace data.
	// This is only used when trace sink is set to "otel".
	OtelTraceExporters []OtelExportConfig `mapstructure:"otel-trace-exporters" json:"otel-trace-exporters"`

	// OtelMetricsExporters is a list of exporters to use for metrics data via OpenTelemetry.
	// When configured, wrapper functions (IncrCounter, SetGauge, etc.) will use OTel instruments
	// instead of go-metrics, and metrics will be exported through these exporters.
	OtelMetricsExporters []OtelExportConfig `mapstructure:"otel-metrics-exporters" json:"otel-metrics-exporters"`
}

// OtelExportConfig defines configuration for an OpenTelemetry exporter.
// This is used for both traces and metrics exporters.
type OtelExportConfig struct {
	// Type is the exporter type.
	// For traces: "stdout", "otlp"
	// For metrics: "stdout", "otlp", "prometheus"
	Type string `mapstructure:"type" json:"type"`

	// OTLPTransport is the transport protocol to use for OTLP.
	// Must be one of:
	// 	- "http" (default)
	//  - "grpc"
	// Only used when Type is "otlp".
	OTLPTransport string `mapstructure:"otlp-transport" json:"otlp-transport"`

	// Endpoint is the OTLP exporter endpoint URL (grpc or http).
	// Only used when Type is "otlp".
	// Example: "localhost:4318" for HTTP, "localhost:4317" for gRPC
	Endpoint string `mapstructure:"endpoint" json:"endpoint"`

	// Insecure disables TLS certificate verification for OTLP exporters.
	// Only used when Type is "otlp".
	Insecure bool `mapstructure:"insecure" json:"insecure"`

	// Headers is a map of HTTP headers to send with each request to the OTLP exporter.
	// Useful for authentication, authorization, or custom metadata.
	// Only used when Type is "otlp".
	// Example: {"Authorization": "Bearer token123", "X-Custom-Header": "value"}
	Headers map[string]string `mapstructure:"headers" json:"headers"`

	// File is the file path to write data to when using the "stdout" exporter.
	// If empty, data is written to stdout.
	// Only used when Type is "stdout".
	File string `mapstructure:"file" json:"file"`

	// PrettyPrint enables pretty-printing of JSON output when using the "stdout" exporter.
	// Only used when Type is "stdout" for traces.
	PrettyPrint bool `mapstructure:"pretty-print" json:"pretty-print"`

	// ListenAddress is the address to listen on for Prometheus scraping.
	// Only used when Type is "prometheus" for metrics.
	// Example: ":8080" or "localhost:9090"
	ListenAddress string `mapstructure:"listen-address" json:"listen-address"`
}
