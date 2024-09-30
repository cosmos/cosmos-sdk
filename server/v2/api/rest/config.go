package rest

func DefaultConfig() *Config {
	return &Config{
		Enable:  true,
		Address: "localhost:8080",
	}
}

type CfgOption func(*Config)

// Config defines configuration for the HTTP server.
type Config struct {
	// Enable defines if the HTTP server should be enabled.
	Enable bool `mapstructure:"enable" toml:"enable" comment:"Enable defines if the HTTP server should be enabled."`
	// Address defines the API server to listen on
	Address string `mapstructure:"address" toml:"address" comment:"Address defines the HTTP server address to bind to."`
}
