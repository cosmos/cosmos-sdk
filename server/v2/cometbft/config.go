package cometbft

import (
	cmtcfg "github.com/cometbft/cometbft/config"
	"github.com/spf13/viper"

	serverv2 "cosmossdk.io/server/v2"
)

// Config is the configuration for the CometBFT application
type Config struct {
	AppTomlConfig    *AppTomlConfig
	ConfigTomlConfig *cmtcfg.Config
}

func DefaultAppTomlConfig() *AppTomlConfig {
	return &AppTomlConfig{
		MinRetainBlocks: 0,
		IndexEvents:     make([]string, 0),
		HaltHeight:      0,
		HaltTime:        0,
		Address:         "tcp://127.0.0.1:26658",
		Transport:       "socket",
		Trace:           false,
		Standalone:      false,
	}
}

type AppTomlConfig struct {
	MinRetainBlocks uint64   `mapstructure:"min-retain-blocks" toml:"min-retain-blocks" comment:"min-retain-blocks defines the minimum block height offset from the current block being committed, such that all blocks past this offset are pruned from CometBFT. A value of 0 indicates that no blocks should be pruned."`
	IndexEvents     []string `mapstructure:"index-events" toml:"index-events" comment:"index-events defines the set of events in the form {eventType}.{attributeKey}, which informs CometBFT what to index. If empty, all events will be indexed."`
	HaltHeight      uint64   `mapstructure:"halt-height" toml:"halt-height" comment:"halt-height contains a non-zero block height at which a node will gracefully halt and shutdown that can be used to assist upgrades and testing."`
	HaltTime        uint64   `mapstructure:"halt-time" toml:"halt-time" comment:"halt-time contains a non-zero minimum block time (in Unix seconds) at which a node will gracefully halt and shutdown that can be used to assist upgrades and testing."`
	Address         string   `mapstructure:"address" toml:"address" comment:"address defines the CometBFT RPC server address to bind to."`
	Transport       string   `mapstructure:"transport" toml:"transport" comment:"transport defines the CometBFT RPC server transport protocol: socket, grpc"`
	Trace           bool     `mapstructure:"trace" toml:"trace" comment:"trace enables the CometBFT RPC server to output trace information about its internal operations."`
	Standalone      bool     `mapstructure:"standalone" toml:"standalone" comment:"standalone starts the application without the CometBFT node. The node should be started separately."`
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

func getConfigTomlFromViper(v *viper.Viper) *cmtcfg.Config {
	rootDir := v.GetString(serverv2.FlagHome)

	conf := cmtcfg.DefaultConfig()
	if err := v.Unmarshal(conf); err != nil {
		return cmtcfg.DefaultConfig().SetRoot(rootDir)
	}

	return conf.SetRoot(rootDir)
}
