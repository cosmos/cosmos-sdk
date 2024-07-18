package cometbft

import (
	cmtcfg "github.com/cometbft/cometbft/config"
	"github.com/spf13/viper"

	"cosmossdk.io/core/transaction"
	serverv2 "cosmossdk.io/server/v2"
	"cosmossdk.io/server/v2/cometbft/types"
)

// Config is the configuration for the CometBFT application
type Config struct {
	AddrPeerFilter     types.PeerFilter // filter peers by address and port
	IdPeerFilter       types.PeerFilter // filter peers by node ID
	ConsensusAuthority string           // Set by the application to grant authority to the consensus engine to send messages to the consensus module

	AppTomlConfig    *AppTomlConfig
	ConfigTomlConfig *cmtcfg.Config
}

func DefaultConfig() *AppTomlConfig {
	return &AppTomlConfig{
		MinRetainBlocks: 1,
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
	MinRetainBlocks uint64   `mapstructure:"min_retain_blocks" toml:"min_retain_blocks"`
	IndexEvents     []string `mapstructure:"index_events" toml:"index_events"`
	HaltHeight      uint64   `mapstructure:"halt_height" toml:"halt_height"`
	HaltTime        uint64   `mapstructure:"halt_time" toml:"halt_time"`
	Address         string   `mapstructure:"address" toml:"address"`
	Transport       string   `mapstructure:"transport" toml:"transport"`
	Trace           bool     `mapstructure:"trace" toml:"trace"`
	Standalone      bool     `mapstructure:"standalone" toml:"standalone"`
}

// CfgOption is a function that allows to overwrite the default server configuration.
type CfgOption func(*Config)

// OverwriteDefaultConfigTomlConfig overwrites the default comet config with the new config.
func OverwriteDefaultConfigTomlConfig(newCfg *cmtcfg.Config) CfgOption {
	return func(cfg *Config) { // nolint:staticcheck // We want to overwrite everything
		cfg.ConfigTomlConfig = newCfg // nolint:ineffassign,staticcheck // We want to overwrite everything
	}
}

// OverwriteDefaultAppTomlConfig overwrites the default comet config with the new config.
func OverwriteDefaultAppTomlConfig(newCfg *AppTomlConfig) CfgOption {
	return func(cfg *Config) { // nolint:staticcheck // We want to overwrite everything
		cfg.AppTomlConfig = newCfg // nolint:ineffassign,staticcheck // We want to overwrite everything
	}
}

func GetConfigTomlFromViper(v *viper.Viper) *cmtcfg.Config {
	conf := cmtcfg.DefaultConfig()
	err := v.Unmarshal(conf)
	rootDir := v.GetString(serverv2.FlagHome)
	if err != nil {
		return cmtcfg.DefaultConfig().SetRoot(rootDir)
	}

	return conf.SetRoot(rootDir)
}

func GetAppTomlFromViper(v *viper.Viper) *AppTomlConfig {
	cfg := DefaultConfig()
	if err := v.Sub((&CometBFTServer[transaction.Tx]{}).Name()).Unmarshal(&cfg); err != nil {
		return cfg
	}

	return cfg
}
