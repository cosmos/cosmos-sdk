package types

import (
	"context"

	abci "github.com/cometbft/cometbft/v2/abci/types"
)

// ABCI is an interface that enables any finite, deterministic state machine
// to be driven by a blockchain-based replication engine via the ABCI.
type ABCI interface {
	// Info/Query Connection
	Info(*abci.InfoRequest) (*abci.InfoResponse, error)                     // Return application info
	Query(context.Context, *abci.QueryRequest) (*abci.QueryResponse, error) // Query for state

	// Mempool Connection
	CheckTx(*abci.CheckTxRequest) (*abci.CheckTxResponse, error) // Validate a tx for the mempool

	// Consensus Connection
	InitChain(*abci.InitChainRequest) (*abci.InitChainResponse, error) // Initialize blockchain w validators/other info from CometBFT
	PrepareProposal(*abci.PrepareProposalRequest) (*abci.PrepareProposalResponse, error)
	ProcessProposal(*abci.ProcessProposalRequest) (*abci.ProcessProposalResponse, error)
	// Deliver the decided block with its txs to the Application
	FinalizeBlock(*abci.FinalizeBlockRequest) (*abci.FinalizeBlockResponse, error)
	// Create application specific vote extension
	ExtendVote(context.Context, *abci.ExtendVoteRequest) (*abci.ExtendVoteResponse, error)
	// Verify application's vote extension data
	VerifyVoteExtension(*abci.VerifyVoteExtensionRequest) (*abci.VerifyVoteExtensionResponse, error)
	// Commit the state and return the application Merkle root hash
	Commit() (*abci.CommitResponse, error)

	// State Sync Connection
	ListSnapshots(*abci.ListSnapshotsRequest) (*abci.ListSnapshotsResponse, error)                // List available snapshots
	OfferSnapshot(*abci.OfferSnapshotRequest) (*abci.OfferSnapshotResponse, error)                // Offer a snapshot to the application
	LoadSnapshotChunk(*abci.LoadSnapshotChunkRequest) (*abci.LoadSnapshotChunkResponse, error)    // Load a snapshot chunk
	ApplySnapshotChunk(*abci.ApplySnapshotChunkRequest) (*abci.ApplySnapshotChunkResponse, error) // Apply a shapshot chunk
}
