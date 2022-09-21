package telemetry

import (
	"fmt"
	"time"

	gometrics "github.com/armon/go-metrics"
	"github.com/spf13/viper"
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
	// It defines the retention duration in seconds.
	PrometheusRetentionTime time.Duration `mapstructure:"prometheus-retention-time"`

	// GlobalLabels defines a global set of name/value label tuples applied to all
	// metrics emitted using the wrapper functions defined in telemetry package.
	//
	// Example:
	// [["chain_id", "cosmoshub-1"]]
	GlobalLabels [][]string `mapstructure:"global-labels"`

	//Suggestion : use the gometrics.Lable and update the config contract. also provide proper Options for best accessibility.
	globalLabels []gometrics.Label
}

// loadDefaults loads the default configuration for the metric config.
func (c *Config) loadDefaults() {
	c.Enabled = defaultEnabled
	c.EnableHostname = defaultEnableHostname
	c.EnableHostnameLabel = defaultEnableHostnameLabel
	c.EnableServiceLabel = defaultEnableServiceLabel

	c.PrometheusRetentionTime = defaultPrometheusRetentionTime
}

func (c *Config) fromViper(v *viper.Viper) error {
	c.ServiceName = v.GetString("telemetry.service-name")
	c.Enabled = v.GetBool("telemetry.enabled")
	c.EnableHostname = v.GetBool("telemetry.enable-hostname")
	c.EnableHostnameLabel = v.GetBool("telemetry.enable-hostname-label")
	c.EnableServiceLabel = v.GetBool("telemetry.enable-service-label")
	c.PrometheusRetentionTime = v.GetDuration("telemetry.prometheus-retention-time")

	// ─── Global Labels ──────────────────────────────────────────────────────────────
	globalLabelsRaw, ok := v.Get("telemetry.global-labels").([]interface{})
	if !ok {
		return fmt.Errorf("failed to parse global-labels config")
	}

	globalLabels := make([][]string, 0, len(globalLabelsRaw))
	for idx, glr := range globalLabelsRaw {
		labelsRaw, ok := glr.([]interface{})
		if !ok {
			return fmt.Errorf("failed to parse global label number %d from config", idx)
		}
		if len(labelsRaw) == 2 {
			globalLabels = append(globalLabels, []string{labelsRaw[0].(string), labelsRaw[1].(string)})
		}
	}
	c.GlobalLabels = globalLabels
	return nil
}
