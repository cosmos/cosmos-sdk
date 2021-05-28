package baseapp

import (
	"context"

	"github.com/tendermint/tendermint/abci/types"
)

type ABCIConsensusHandler interface {
	InitChain(context.Context, types.RequestInitChain) types.ResponseInitChain    // Initialize blockchain w validators/other info from TendermintCore
	BeginBlock(context.Context, types.RequestBeginBlock) types.ResponseBeginBlock // Signals the beginning of a block
	DeliverTx(context.Context, types.RequestDeliverTx) types.ResponseDeliverTx    // Deliver a tx for full processing
	EndBlock(context.Context, types.RequestEndBlock) types.ResponseEndBlock       // Signals the end of a block, returns changes to the validator set
	Commit(context.Context) types.ResponseCommit                                  // Commit the state and return the application Merkle root hash
}

type ABCIConsensusMiddleware interface {
	OnInitChain(context.Context, types.RequestInitChain, ABCIConsensusHandler) types.ResponseInitChain    // Initialize blockchain w validators/other info from TendermintCore
	OnBeginBlock(context.Context, types.RequestBeginBlock, ABCIConsensusHandler) types.ResponseBeginBlock // Signals the beginning of a block
	OnDeliverTx(context.Context, types.RequestDeliverTx, ABCIConsensusHandler) types.ResponseDeliverTx    // Deliver a tx for full processing
	OnEndBlock(context.Context, types.RequestEndBlock, ABCIConsensusHandler) types.ResponseEndBlock       // Signals the end of a block, returns changes to the validator set
	OnCommit(context.Context, ABCIConsensusHandler) types.ResponseCommit                                  // Commit the state and return the application Merkle root hash
}

type ABCIMempoolHandler interface {
	CheckTx(context.Context, types.RequestCheckTx) types.ResponseCheckTx // Validate a tx for the mempool
}

type ABCIMempoolMiddleware interface {
	CheckTx(context.Context, types.RequestCheckTx, ABCIMempoolHandler) types.ResponseCheckTx // Validate a tx for the mempool
}

type ABCIQueryMiddleware interface {
	OnInfo(context.Context, types.RequestInfo, types.Application) types.ResponseInfo                // Return application info
	OnSetOption(context.Context, types.RequestSetOption, types.Application) types.ResponseSetOption // Set application option
	OnQuery(context.Context, types.RequestQuery, types.Application) types.ResponseQuery             // Query for state
}

type ABCIStateSyncMiddleware interface {
	OnListSnapshots(context.Context, types.RequestListSnapshots, types.Application) types.ResponseListSnapshots                // List available snapshots
	OnOfferSnapshot(context.Context, types.RequestOfferSnapshot, types.Application) types.ResponseOfferSnapshot                // Offer a snapshot to the application
	OnLoadSnapshotChunk(context.Context, types.RequestLoadSnapshotChunk, types.Application) types.ResponseLoadSnapshotChunk    // Load a snapshot chunk
	OnApplySnapshotChunk(context.Context, types.RequestApplySnapshotChunk, types.Application) types.ResponseApplySnapshotChunk // Apply a shapshot chunk
}
