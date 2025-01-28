package swagger

// Config defines the configuration for the Swagger UI server
type Config struct {
	// Enable enables/disables the Swagger UI server
	Enable bool `mapstructure:"enable" toml:"enable" comment:"Enable enables/disables the Swagger UI server"`
	// Address defines the server address to bind to
	Address string `mapstructure:"address" toml:"address" comment:"Address defines the server address to bind to"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Enable:  true,
		Address: "localhost:8090",
	}
}

// CfgOption defines a function for configuring the settings
type CfgOption func(*Config)

// OverwriteDefaultConfig overwrites the default config with the new config.
func OverwriteDefaultConfig(newCfg *Config) CfgOption {
	return func(cfg *Config) {
		*cfg = *newCfg
	}
}

// Disable the grpc server by default (default enabled).
func Disable() CfgOption {
	return func(cfg *Config) {
		cfg.Enable = false
	}
}
