// Package telemetry provides instrumentation tools for other modules and packages.
package telemetry

import (
	"time"

	"github.com/spf13/viper"
)

// Option modifies the default configuration of metrics.
type Option func(*Config) error

// OptionWithConfig overwrites the initialization config with the new value.
func OptionWithConfig(C Config) Option {
	return func(c *Config) error {
		*c = C
		return nil
	}
}

// OptionFromViper uses a viper struct to read the configuration.
// if the configuration changes in runtime the object will not update.(RUNTIME UPDATE is not supported yet.)
func OptionFromViper(v *viper.Viper) Option {
	return func(c *Config) error {
		return c.fromViper(v)
	}
}

// OptionWithEnable enables/disables the metrics based on the value.
// true: enable
// false: disable
// default : disable
func OptionWithEnable(b bool) Option {
	return func(c *Config) error {
		c.Enabled = b
		return nil
	}
}

// OptionWithServiceName sets the service name.
func OptionWithServiceName(s string) Option {
	return func(c *Config) error {
		c.ServiceName = s
		return nil
	}
}

// OptionWithEnableHostname enables/disables the hostname based on the value.
// true: enable
// false: disable
// default : disable
func OptionWithEnableHostname(b bool) Option {
	return func(c *Config) error {
		c.EnableHostname = b
		return nil
	}
}

// OptionWithEnableHostnameLabel enables/disables the hostname labels based on the value.
// true: enable
// false: disable
// default : disable
func OptionWithEnableHostnameLabel(b bool) Option {
	return func(c *Config) error {
		c.EnableHostnameLabel = b
		return nil
	}
}

// OptionWithEnableServiceLabel enables/disables the service labels based on the value.
func OptionWithEnableServiceLabel(b bool) Option {
	return func(c *Config) error {
		c.EnableServiceLabel = b
		return nil
	}
}

// OptionWithPrometheusRetentionTime sets the retention time of the prometheus.
// default: 1 second
func OptionWithPrometheusRetentionTime(d time.Duration) Option {
	return func(c *Config) error {
		c.PrometheusRetentionTime = d
		return nil
	}
}

// OptionWithGlobalLabels sets the global labels for the metrics.
func OptionWithGlobalLabels(ss [][]string) Option {
	return func(c *Config) error {
		c.GlobalLabels = ss
		return nil
	}
}

// OptionUseGlobalMetricRegistration use the global registration for external libraries like go-metrics.
// default: true
func OptionUseGlobalMetricRegistration(b bool) Option {
	return func(c *Config) error {
		c.useGlobalMetricRegistration = b
		return nil
	}
}
