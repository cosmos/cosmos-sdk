package swagger

import "github.com/cosmos/cosmos-sdk/server/v2/config"

// Config represents Swagger configuration options
type Config struct {
    // Enable enables the Swagger UI endpoint
    Enable bool `mapstructure:"enable"`
    // Path is the URL path where Swagger UI will be served
    Path string `mapstructure:"path"`
}

// DefaultConfig returns default configuration for Swagger
func DefaultConfig() Config {
    return Config{
        Enable: false,
        Path:   "/swagger",
    }
}

// Validate validates the configuration
func (c Config) Validate() error {
    if c.Path == "" {
        return fmt.Errorf("swagger path cannot be empty")
    }
    return nil
} 
