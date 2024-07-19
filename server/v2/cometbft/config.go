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
	return func(cfg *Config) {
		cfg.ConfigTomlConfig = newCfg // nolint:ineffassign,staticcheck // We want to overwrite everything
	}
}

// OverwriteDefaultAppTomlConfig overwrites the default comet config with the new config.
func OverwriteDefaultAppTomlConfig(newCfg *AppTomlConfig) CfgOption {
	return func(cfg *Config) {
		cfg.AppTomlConfig = newCfg // nolint:ineffassign,staticcheck // We want to overwrite everything
	}
}

func getConfigTomlFromViper(v *viper.Viper) *cmtcfg.Config {
	conf := cmtcfg.DefaultConfig()
	err := v.Unmarshal(conf)
	rootDir := v.GetString(serverv2.FlagHome)
	if err != nil {
		return cmtcfg.DefaultConfig().SetRoot(rootDir)
	}

	return conf.SetRoot(rootDir)
}
