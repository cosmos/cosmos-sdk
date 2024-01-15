package cometbft

import (
	cmtcfg "github.com/cometbft/cometbft/config"

	"cosmossdk.io/server/v2/cometbft/types"
	"cosmossdk.io/store/v2/snapshots"
)

type Config struct {
	Name            string // TODO: we might want to put some of these in the app manager
	Version         string
	InitialHeight   uint64
	MinRetainBlocks uint64
	IndexEvents     map[string]struct{}
	HaltHeight      uint64
	HaltTime        uint64

	SnapshotManager *snapshots.Manager

	AddrPeerFilter types.PeerFilter // filter peers by address and port
	IdPeerFilter   types.PeerFilter // filter peers by node ID

	Transport  string
	Addr       string
	Standalone bool
	Trace      bool

	CmtConfig *cmtcfg.Config
}
