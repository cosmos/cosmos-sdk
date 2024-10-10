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
