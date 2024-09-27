package telemetry

func DefaultConfig() *Config {
	return &Config{
		Enable:                  true,
		Address:                 "localhost:1318",
		ServiceName:             "",
		EnableHostname:          false,
		EnableHostnameLabel:     false,
		EnableServiceLabel:      false,
		PrometheusRetentionTime: 0,
		GlobalLabels:            nil,
		MetricsSink:             "",
		StatsdAddr:              "",
		DatadogHostname:         "",
	}
}

type Config struct {
	// Enable enables the application telemetry functionality. When enabled,
	// an in-memory sink is also enabled by default. Operators may also enabled
	// other sinks such as Prometheus.
	Enable bool `mapstructure:"enable" toml:"enable" comment:"Enable enables the application telemetry functionality. When enabled, an in-memory sink is also enabled by default. Operators may also enabled other sinks such as Prometheus."`

	// Address defines the API server to listen on
	Address string `mapstructure:"address" toml:"address" comment:"Address defines the metrics server address to bind to."`

	// Prefixed with keys to separate services
	ServiceName string `mapstructure:"service-name" toml:"service-name" comment:"Prefixed with keys to separate services."`

	// Enable prefixing gauge values with hostname
	EnableHostname bool `mapstructure:"enable-hostname" toml:"enable-hostname" comment:"Enable prefixing gauge values with hostname."`

	// Enable adding hostname to labels
	EnableHostnameLabel bool `mapstructure:"enable-hostname-label" toml:"enable-hostname-label" comment:"Enable adding hostname to labels."`

	// Enable adding service to labels
	EnableServiceLabel bool `mapstructure:"enable-service-label" toml:"enable-service-label" comment:"Enable adding service to labels."`

	// PrometheusRetentionTime, when positive, enables a Prometheus metrics sink.
	// It defines the retention duration in seconds.
	PrometheusRetentionTime int64 `mapstructure:"prometheus-retention-time" toml:"prometheus-retention-time" comment:"PrometheusRetentionTime, when positive, enables a Prometheus metrics sink. It defines the retention duration in seconds."`

	// GlobalLabels defines a global set of name/value label tuples applied to all
	// metrics emitted using the wrapper functions defined in telemetry package.
	//
	// Example:
	// [["chain_id", "cosmoshub-1"]]
	GlobalLabels [][]string `mapstructure:"global-labels" toml:"global-labels" comment:"GlobalLabels defines a global set of name/value label tuples applied to all metrics emitted using the wrapper functions defined in telemetry package.\n Example:\n  [[\"chain_id\", \"cosmoshub-1\"]]"`

	// MetricsSink defines the type of metrics backend to use.
	MetricsSink string `mapstructure:"type" toml:"metrics-sink" comment:"MetricsSink defines the type of metrics backend to use. Default is in memory"`

	// StatsdAddr defines the address of a statsd server to send metrics to.
	// Only utilized if MetricsSink is set to "statsd" or "dogstatsd".
	StatsdAddr string `mapstructure:"statsd-addr" toml:"stats-addr" comment:"StatsdAddr defines the address of a statsd server to send metrics to. Only utilized if MetricsSink is set to \"statsd\" or \"dogstatsd\"."`

	// DatadogHostname defines the hostname to use when emitting metrics to
	// Datadog. Only utilized if MetricsSink is set to "dogstatsd".
	DatadogHostname string `mapstructure:"datadog-hostname" toml:"data-dog-hostname" comment:"DatadogHostname defines the hostname to use when emitting metrics to Datadog. Only utilized if MetricsSink is set to \"dogstatsd\"."`
}
