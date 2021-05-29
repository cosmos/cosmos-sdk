package abci

import (
	"context"

	types "github.com/tendermint/tendermint/abci/types"
)

type Handler interface {
	// Info/Query Connection
	Info(ctx context.Context, req types.RequestInfo) types.ResponseInfo                // Return application info
	SetOption(ctx context.Context, req types.RequestSetOption) types.ResponseSetOption // Set application option
	Query(ctx context.Context, req types.RequestQuery) types.ResponseQuery             // Query for state

	// Mempool Connection
	CheckTx(ctx context.Context, req types.RequestCheckTx) types.ResponseCheckTx // Validate a tx for the mempool

	// Consensus Connection
	InitChain(ctx context.Context, req types.RequestInitChain) types.ResponseInitChain    // Initialize blockchain w validators/other info from TendermintCore
	BeginBlock(ctx context.Context, req types.RequestBeginBlock) types.ResponseBeginBlock // Signals the beginning of a block
	DeliverTx(ctx context.Context, req types.RequestDeliverTx) types.ResponseDeliverTx    // Deliver a tx for full processing
	EndBlock(ctx context.Context, req types.RequestEndBlock) types.ResponseEndBlock       // Signals the end of a block, returns changes to the validator set
	Commit(context.Context) types.ResponseCommit                                          // Commit the state and return the application Merkle root hash

	// State Sync Connection
	ListSnapshots(ctx context.Context, req types.RequestListSnapshots) types.ResponseListSnapshots                // List available snapshots
	OfferSnapshot(ctx context.Context, req types.RequestOfferSnapshot) types.ResponseOfferSnapshot                // Offer a snapshot to the application
	LoadSnapshotChunk(ctx context.Context, req types.RequestLoadSnapshotChunk) types.ResponseLoadSnapshotChunk    // Load a snapshot chunk
	ApplySnapshotChunk(ctx context.Context, req types.RequestApplySnapshotChunk) types.ResponseApplySnapshotChunk // Apply a shapshot chunk
}
