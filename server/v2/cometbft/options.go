package cometbft

import (
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/server/v2/cometbft/handlers"
	"cosmossdk.io/server/v2/cometbft/mempool"
	"cosmossdk.io/server/v2/cometbft/types"
	"cosmossdk.io/store/v2/snapshots"
)

// ServerOptions defines the options for the CometBFT server.
// When an option takes a map[string]any, it can access the app.tom's cometbft section and the config.toml config.
type ServerOptions[T transaction.Tx] struct {
	PrepareProposalHandler     handlers.PrepareHandler[T]
	ProcessProposalHandler     handlers.ProcessHandler[T]
	VerifyVoteExtensionHandler handlers.VerifyVoteExtensionhandler
	ExtendVoteHandler          handlers.ExtendVoteHandler

	Mempool         func(cfg map[string]any) mempool.Mempool[T]
	SnapshotOptions func(cfg map[string]any) snapshots.SnapshotOptions

	AddrPeerFilter types.PeerFilter // filter peers by address and port
	IdPeerFilter   types.PeerFilter // filter peers by node ID
}

// DefaultServerOptions returns the default server options.
// It defaults to a NoOpMempool and NoOp handlers.
func DefaultServerOptions[T transaction.Tx]() ServerOptions[T] {
	return ServerOptions[T]{
		PrepareProposalHandler:     handlers.NoOpPrepareProposal[T](),
		ProcessProposalHandler:     handlers.NoOpProcessProposal[T](),
		VerifyVoteExtensionHandler: handlers.NoOpVerifyVoteExtensionHandler(),
		ExtendVoteHandler:          handlers.NoOpExtendVote(),
		Mempool:                    func(cfg map[string]any) mempool.Mempool[T] { return mempool.NoOpMempool[T]{} },
		SnapshotOptions:            func(cfg map[string]any) snapshots.SnapshotOptions { return snapshots.NewSnapshotOptions(0, 0) },
		AddrPeerFilter:             nil,
		IdPeerFilter:               nil,
	}
}
