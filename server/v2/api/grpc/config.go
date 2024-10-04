package grpc

import "math"

func DefaultConfig() *Config {
	return &Config{
		Enable:         true,
		Address:        "localhost:9090",
		MaxRecvMsgSize: 1024 * 1024 * 10,
		MaxSendMsgSize: math.MaxInt32,
	}
}

// Config defines configuration for the gRPC server.
type Config struct {
	// Enable defines if the gRPC server should be enabled.
	Enable bool `mapstructure:"enable" toml:"enable" comment:"Enable defines if the gRPC server should be enabled."`

	// Address defines the API server to listen on
	Address string `mapstructure:"address" toml:"address" comment:"Address defines the gRPC server address to bind to."`

	// MaxRecvMsgSize defines the max message size in bytes the server can receive.
	// The default value is 10MB.
	MaxRecvMsgSize int `mapstructure:"max-recv-msg-size" toml:"max-recv-msg-size" comment:"MaxRecvMsgSize defines the max message size in bytes the server can receive.\nThe default value is 10MB."`

	// MaxSendMsgSize defines the max message size in bytes the server can send.
	// The default value is math.MaxInt32.
	MaxSendMsgSize int `mapstructure:"max-send-msg-size" toml:"max-send-msg-size" comment:"MaxSendMsgSize defines the max message size in bytes the server can send.\nThe default value is math.MaxInt32."`
}

// CfgOption is a function that allows to overwrite the default server configuration.
type CfgOption func(*Config)

// OverwriteDefaultConfig overwrites the default config with the new config.
func OverwriteDefaultConfig(newCfg *Config) CfgOption {
	return func(cfg *Config) {
		*cfg = *newCfg
	}
}

// Disable the grpc-gateway server by default (default enabled).
func Disable() CfgOption {
	return func(cfg *Config) {
		cfg.Enable = false
	}
}
