package cometbft

import (
	"cosmossdk.io/server/v2/api/grpc"
	"cosmossdk.io/server/v2/cometbft/types"
)

// Config is the configuration for the CometBFT application
type Config struct {
	// app.toml config options
	Name            string
	Version         string
	InitialHeight   uint64
	MinRetainBlocks uint64
	IndexEvents     map[string]struct{}
	HaltHeight      uint64
	HaltTime        uint64
	// end of app.toml config options

	AddrPeerFilter types.PeerFilter // filter peers by address and port
	IdPeerFilter   types.PeerFilter // filter peers by node ID

	Transport  string
	Addr       string
	Standalone bool
	Trace      bool

	GrpcConfig grpc.Config
	MempoolConfig
	// CmtConfig *cmtcfg.Config
}
