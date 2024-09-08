package logmonitor

import "fmt"

const (
	ConfigKey = "log-monitor"
)

// Config holds the configuration for LogMonitor
type Config struct {
	Enabled         bool     `mapstructure:"enabled"`
	ShutdownStrings []string `mapstructure:"shutdown-strings"`
}

// DefaultConfig returns a default configuration for LogMonitor
func DefaultConfig() Config {
	return Config{
		Enabled:         false,
		ShutdownStrings: []string{"CONSENSUS FAILURE!"},
	}
}

// Validate checks if the config is valid
func (c Config) Validate() error {
	if len(c.ShutdownStrings) == 0 {
		return fmt.Errorf("at least one shutdown string must be provided")
	}
	return nil
}
