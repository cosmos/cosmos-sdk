package types

import (
	"context"

	abci "github.com/cometbft/cometbft/abci/types"
)

// ABCI is an interface that enables any finite, deterministic state machine
// to be driven by a blockchain-based replication engine via the ABCI.
type ABCI interface {
	// Info/Query Connection

	// Info returns application info
	Info(*abci.RequestInfo) (*abci.ResponseInfo, error)
	// Query  returns application state
	Query(context.Context, *abci.RequestQuery) (*abci.ResponseQuery, error)

	// Mempool Connection

	// CheckTx validate a tx for the mempool
	CheckTx(*abci.RequestCheckTx) (*abci.ResponseCheckTx, error)

	// Consensus Connection

	// InitChain Initialize blockchain w validators/other info from CometBFT
	InitChain(*abci.RequestInitChain) (*abci.ResponseInitChain, error)
	PrepareProposal(*abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error)
	ProcessProposal(*abci.RequestProcessProposal) (*abci.ResponseProcessProposal, error)
	// FinalizeBlock deliver the decided block with its txs to the Application
	FinalizeBlock(*abci.RequestFinalizeBlock) (*abci.ResponseFinalizeBlock, error)
	// ExtendVote create application specific vote extension
	ExtendVote(context.Context, *abci.RequestExtendVote) (*abci.ResponseExtendVote, error)
	// VerifyVoteExtension verify application's vote extension data
	VerifyVoteExtension(*abci.RequestVerifyVoteExtension) (*abci.ResponseVerifyVoteExtension, error)
	// Commit the state and return the application Merkle root hash
	Commit() (*abci.ResponseCommit, error)

	// State Sync Connection

	// ListSnapshots list available snapshots
	ListSnapshots(*abci.RequestListSnapshots) (*abci.ResponseListSnapshots, error)
	// OfferSnapshot offer a snapshot to the application
	OfferSnapshot(*abci.RequestOfferSnapshot) (*abci.ResponseOfferSnapshot, error)
	// LoadSnapshotChunk load a snapshot chunk
	LoadSnapshotChunk(*abci.RequestLoadSnapshotChunk) (*abci.ResponseLoadSnapshotChunk, error)
	// ApplySnapshotChunk apply a snapshot chunk
	ApplySnapshotChunk(*abci.RequestApplySnapshotChunk) (*abci.ResponseApplySnapshotChunk, error)
}
