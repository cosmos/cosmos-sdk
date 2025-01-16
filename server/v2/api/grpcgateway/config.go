package grpcgateway

func DefaultConfig() *Config {
	return &Config{
		Enable:  true,
		Address: "localhost:1317",
	}
}

type Config struct {
	// Enable defines if the gRPC-Gateway should be enabled.
	Enable bool `mapstructure:"enable" toml:"enable" comment:"Enable defines if the gRPC-Gateway should be enabled."`

	// Address defines the address the gRPC-Gateway server binds to.
	Address string `mapstructure:"address" toml:"address" comment:"Address defines the address the gRPC-Gateway server binds to."`
}

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
