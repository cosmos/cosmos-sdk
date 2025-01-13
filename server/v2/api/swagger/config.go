package swagger

import (
    "fmt"
    "net/http"

    "cosmossdk.io/core/server"
)

const ServerName = "swagger"

// Config defines the configuration for the Swagger UI server
type Config struct {
    // Enable enables/disables the Swagger UI server
    Enable bool `toml:"enable,omitempty" mapstructure:"enable"`
    // Address defines the server address to bind to
    Address string `toml:"address,omitempty" mapstructure:"address"`
    // SwaggerUI defines the file system for serving Swagger UI files
    SwaggerUI http.FileSystem `toml:"-" mapstructure:"-"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
    return &Config{
        Enable:  true,
        Address: "localhost:8090",
    }
}

// Validate returns an error if the config is invalid
func (c *Config) Validate() error {
    if !c.Enable {
        return nil
    }

    if c.Address == "" {
        return fmt.Errorf("address is required when swagger UI is enabled")
    }

    return nil
}

// CfgOption defines a function for configuring the settings
type CfgOption func(*Config)
