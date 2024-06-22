package grpcgateway

type Config struct {
	// Enable defines if the gRPC-gateway should be enabled.
	Enable bool `mapstructure:"enable" toml:"enable" comment:"Enable defines if the gRPC-gateway should be enabled."`
}

func DefaultConfig() Config {
	return Config{
		Enable: true,
	}
}
