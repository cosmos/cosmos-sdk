package streaming

// State Streaming configuration

// StreamingConfig defines application configuration for external streaming services
type StreamingConfig struct {
	ListenerConfig ListenerConfig `mapstructure:"listener-config" toml:"listener-config" comment:"ListenerConfig defines application configuration for ABCIListener streaming service"`
}

// ListenerConfig defines application configuration for ABCIListener streaming service
type ListenerConfig struct {
	// List of kv store keys to stream out via gRPC.
	// The store key names MUST match the module's StoreKey name.
	//
	// Example:
	// ["acc", "bank", "gov", "staking", "mint"[,...]]
	// ["*"] to expose all keys.
	Keys []string `mapstructure:"keys" toml:"keys" comment:"List of kv store keys to stream out via gRPC. The store key names MUST match the module's StoreKey name. Example: [\"acc\", \"bank\", \"gov\", \"staking\", \"mint\"[,...]] [\"*\"] to expose all keys."`
	// The plugin name used for streaming via gRPC.
	// Streaming is only enabled if this is set.
	// Supported plugins: abci
	Plugin string `mapstructure:"plugin" toml:"plugin" comment:"The plugin name used for streaming via gRPC. Streaming is only enabled if this is set. Supported plugins: abci"`
	// stop-node-on-err specifies whether to stop the node on message delivery error.
	StopNodeOnErr bool `mapstructure:"stop-node-on-err" toml:"stop-node-on-err" comment:"stop-node-on-err specifies whether to stop the node on message delivery error."`
}
