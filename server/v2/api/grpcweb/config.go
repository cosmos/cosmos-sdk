package grpcweb

// GRPCWebConfig defines configuration for the gRPC-web server.
type Config struct {
	// GRPCWebEnable defines if the gRPC-web should be enabled.
	// NOTE: gRPC must also be enabled, otherwise, this configuration is a no-op.
	// NOTE: gRPC-Web uses the same address as the API server.
	Enable bool `mapstructure:"enable" toml:"enable" comment:"Enable defines if the gRPC-web should be enabled. NOTE: gRPC must also be enabled, otherwise, this configuration is a no-op. NOTE: gRPC-Web uses the same address as the API server."`
}

func DefaultConfig() Config {
	return Config{
		Enable: true,
	}
}
