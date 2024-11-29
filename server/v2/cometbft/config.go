package cometbft

import (
	cmtcfg "github.com/cometbft/cometbft/config"

	"cosmossdk.io/schema/indexer"
	"cosmossdk.io/server/v2/cometbft/mempool"
)

// Config is the configuration for the CometBFT application
type Config struct {
	AppTomlConfig    *AppTomlConfig
	ConfigTomlConfig *cmtcfg.Config
}

func DefaultAppTomlConfig() *AppTomlConfig {
	return &AppTomlConfig{
		MinRetainBlocks: 0,
		HaltHeight:      0,
		HaltTime:        0,
		Address:         "tcp://127.0.0.1:26658",
		Transport:       "socket",
		Trace:           false,
		Standalone:      false,
		Mempool:         mempool.DefaultConfig(),
		Indexer: indexer.IndexingConfig{
			Target:            make(map[string]indexer.Config),
			ChannelBufferSize: 1024,
		},
		IndexABCIEvents:        make([]string, 0),
		DisableIndexABCIEvents: false,
		DisableABCIEvents:      false,
	}
}

type AppTomlConfig struct {
	MinRetainBlocks uint64 `mapstructure:"min-retain-blocks" toml:"min-retain-blocks" comment:"min-retain-blocks defines the minimum block height offset from the current block being committed, such that all blocks past this offset are pruned from CometBFT. A value of 0 indicates that no blocks should be pruned."`
	HaltHeight      uint64 `mapstructure:"halt-height" toml:"halt-height" comment:"halt-height contains a non-zero block height at which a node will gracefully halt and shutdown that can be used to assist upgrades and testing."`
	HaltTime        uint64 `mapstructure:"halt-time" toml:"halt-time" comment:"halt-time contains a non-zero minimum block time (in Unix seconds) at which a node will gracefully halt and shutdown that can be used to assist upgrades and testing."`
	Address         string `mapstructure:"address" toml:"address" comment:"address defines the CometBFT RPC server address to bind to."`
	Transport       string `mapstructure:"transport" toml:"transport" comment:"transport defines the CometBFT RPC server transport protocol: socket, grpc"`
	Trace           bool   `mapstructure:"trace" toml:"trace" comment:"trace enables the CometBFT RPC server to output trace information about its internal operations."`
	Standalone      bool   `mapstructure:"standalone" toml:"standalone" comment:"standalone starts the application without the CometBFT node. The node should be started separately."`

	// Sub configs
	Mempool                mempool.Config         `mapstructure:"mempool" toml:"mempool" comment:"mempool defines the configuration for the SDK built-in app-side mempool implementations."`
	Indexer                indexer.IndexingConfig `mapstructure:"indexer" toml:"indexer" comment:"indexer defines the configuration for the SDK built-in indexer implementation."`
	IndexABCIEvents        []string               `mapstructure:"index-abci-events" toml:"index-abci-events" comment:"index-abci-events defines the set of events in the form {eventType}.{attributeKey}, which informs CometBFT what to index. If empty, all events will be indexed."`
	DisableIndexABCIEvents bool                   `mapstructure:"disable-index-abci-events" toml:"disable-index-abci-events" comment:"disable-index-abci-events disables the ABCI event indexing done by CometBFT. Useful when relying on the SDK indexer for event indexing, but still want events to be included in FinalizeBlockResponse."`
	DisableABCIEvents      bool                   `mapstructure:"disable-abci-events" toml:"disable-abci-events" comment:"disable-abci-events disables all ABCI events. Useful when relying on the SDK indexer for event indexing."`
}

// CfgOption is a function that allows to overwrite the default server configuration.
type CfgOption func(*Config)

// OverwriteDefaultConfigTomlConfig overwrites the default comet config with the new config.
func OverwriteDefaultConfigTomlConfig(newCfg *cmtcfg.Config) CfgOption {
	return func(cfg *Config) {
		cfg.ConfigTomlConfig = newCfg
	}
}

// OverwriteDefaultAppTomlConfig overwrites the default comet config with the new config.
func OverwriteDefaultAppTomlConfig(newCfg *AppTomlConfig) CfgOption {
	return func(cfg *Config) {
		cfg.AppTomlConfig = newCfg
	}
}
