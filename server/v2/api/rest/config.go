package rest

func DefaultConfig() *Config {
	return &Config{
		Enable:  true,
		Address: "localhost:8080",
	}
}

type CfgOption func(*Config)

// Config defines configuration for the REST server.
type Config struct {
	// Enable defines if the REST server should be enabled.
	Enable bool `mapstructure:"enable" toml:"enable" comment:"Enable defines if the REST server should be enabled."`
	// Address defines the API server to listen on
	Address string `mapstructure:"address" toml:"address" comment:"Address defines the REST server address to bind to."`
}

// OverwriteDefaultConfig overwrites the default config with the new config.
func OverwriteDefaultConfig(newCfg *Config) CfgOption {
	return func(cfg *Config) {
		*cfg = *newCfg
	}
}

// Disable the rest server by default (default enabled).
func Disable() CfgOption {
	return func(cfg *Config) {
		cfg.Enable = false
	}
}
