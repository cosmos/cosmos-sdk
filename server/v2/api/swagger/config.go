package swagger

import (
    "fmt"
    "net/http"

    "cosmossdk.io/core/server"
)

const ServerName = "swagger"

// Config defines the configuration for the Swagger UI server
type Config struct {
    Enable    bool            `toml:"enable" mapstructure:"enable"`
    Address   string         `toml:"address" mapstructure:"address"`
    Path      string         `toml:"path" mapstructure:"path"`
    SwaggerUI http.FileSystem `toml:"-" mapstructure:"-"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
    return &Config{
        Enable:  true,
        Address: "localhost:8080",
        Path:    "/swagger/",
    }
}

// Validate checks the configuration
func (c Config) Validate() error {
    if c.Path == "" {
        return fmt.Errorf("swagger path cannot be empty")
    }
    if c.Enable && c.SwaggerUI == nil {
        return fmt.Errorf("swagger UI file system must be provided when enabled")
    }
    return nil
}

// CfgOption defines a function for configuring the settings
type CfgOption func(*Config)
