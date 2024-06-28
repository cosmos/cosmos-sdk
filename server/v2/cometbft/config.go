package cometbft

import (
	cmtcfg "github.com/cometbft/cometbft/config"
	"github.com/spf13/viper"

	serverv2 "cosmossdk.io/server/v2"
	"cosmossdk.io/server/v2/api/grpc"
	"cosmossdk.io/server/v2/cometbft/types"
)

// TODO REDO/VERIFY THIS

func GetConfigFromViper(v *viper.Viper) *cmtcfg.Config {
	conf := cmtcfg.DefaultConfig()
	err := v.Unmarshal(conf)
	rootDir := v.GetString(serverv2.FlagHome)
	if err != nil {
		return cmtcfg.DefaultConfig().SetRoot(rootDir)
	}

	return conf.SetRoot(rootDir)
}

// Config is the configuration for the CometBFT application
type Config struct {
	// app.toml config options
	Name            string              `mapstructure:"name" toml:"name"`
	Version         string              `mapstructure:"version" toml:"version"`
	InitialHeight   uint64              `mapstructure:"initial_height" toml:"initial_height"`
	MinRetainBlocks uint64              `mapstructure:"min_retain_blocks" toml:"min_retain_blocks"`
	IndexEvents     map[string]struct{} `mapstructure:"index_events" toml:"index_events"`
	HaltHeight      uint64              `mapstructure:"halt_height" toml:"halt_height"`
	HaltTime        uint64              `mapstructure:"halt_time" toml:"halt_time"`
	// end of app.toml config options

	AddrPeerFilter types.PeerFilter // filter peers by address and port
	IdPeerFilter   types.PeerFilter // filter peers by node ID

	Transport  string `mapstructure:"transport" toml:"transport"`
	Addr       string `mapstructure:"addr" toml:"addr"`
	Standalone bool   `mapstructure:"standalone" toml:"standalone"`
	Trace      bool   `mapstructure:"trace" toml:"trace"`

	GrpcConfig grpc.Config

	// MempoolConfig
	CmtConfig *cmtcfg.Config

	// Must be set by the application to grant authority to the consensus engine to send messages to the consensus module
	ConsensusAuthority string
}
