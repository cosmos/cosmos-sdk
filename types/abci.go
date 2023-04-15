package types

import (
	abci "github.com/cometbft/cometbft/abci/types"
)

// InitChainer initializes application state at genesis
type InitChainer func(ctx Context, req *abci.RequestInitChain) (*abci.ResponseInitChain, error)

// PeerFilter responds to p2p filtering queries from Tendermint
type PeerFilter func(info string) *abci.ResponseQuery

// ProcessProposalHandler defines a function type alias for processing a proposer
type ProcessProposalHandler func(Context, *abci.RequestProcessProposal) (*abci.ResponseProcessProposal, error)

// PrepareProposalHandler defines a function type alias for preparing a proposal
type PrepareProposalHandler func(Context, *abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error)

// ExtendVoteHandler defines a function type alias for extending a pre-commit vote.
type ExtendVoteHandler func(Context, *abci.RequestExtendVote) (*abci.ResponseExtendVote, error)

// VerifyVoteExtensionHandler defines a function type alias for verifying a
// pre-commit vote extension.
type VerifyVoteExtensionHandler func(Context, *abci.RequestVerifyVoteExtension) (*abci.ResponseVerifyVoteExtension, error)

// LegacyBeginBlocker defines a function type alias for executing application
// business logic before transactions are executed.
//
// Note: The BeginBlock ABCI method no longer exists in the ABCI specification
// as of CometBFT v0.38.0. This function type alias is provided for backwards
// compatibility with applications that still use the BeginBlock ABCI method
// and allows for existing BeginBlock functionality within applications.
type LegacyBeginBlocker func(Context, LegacyRequestBeginBlock) (LegacyResponseBeginBlock, error)

// LegacyEndBlocker defines a function type alias for executing application
// business logic after transactions are executed but before committing.
//
// Note: The EndBlock ABCI method no longer exists in the ABCI specification
// as of CometBFT v0.38.0. This function type alias is provided for backwards
// compatibility with applications that still use the EndBlock ABCI method
// and allows for existing EndBlock functionality within applications.
type EndBlocker func(Context) (EndBlock, error)

// EndBlocker defines a type which contains endblock events and validator set updates
type EndBlock struct {
	ValidatorUpdates []abci.ValidatorUpdate
	Events           []abci.Event
}
