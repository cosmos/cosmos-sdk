package telemetry

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
	// It defines the retention duration in seconds.
	PrometheusRetentionTime int64 `mapstructure:"prometheus-retention-time"`

	// GlobalLabels defines a global set of name/value label tuples applied to all
	// metrics emitted using the wrapper functions defined in telemetry package.
	//
	// Example:
	// [["chain_id", "cosmoshub-1"]]
	GlobalLabels [][]string `mapstructure:"global-labels"`

	// MetricsSink defines the type of metrics backend to use.
	MetricsSink string `mapstructure:"type" default:"mem"`

	// StatsdAddr defines the address of a statsd server to send metrics to.
	// Only utilized if MetricsSink is set to "statsd" or "dogstatsd".
	StatsdAddr string `mapstructure:"statsd-addr"`

	// DatadogHostname defines the hostname to use when emitting metrics to
	// Datadog. Only utilized if MetricsSink is set to "dogstatsd".
	DatadogHostname string `mapstructure:"datadog-hostname"`
}
