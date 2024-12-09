package types

import (
	"context"

	abci "github.com/cometbft/cometbft/abci/types"
)

// InitChainer initializes application state at genesis
type InitChainer func(ctx Context, req *abci.RequestInitChain) (*abci.ResponseInitChain, error)

// PrepareCheckStater runs code during commit after the block has been committed, and the `checkState`
// has been branched for the new block.
type PrepareCheckStater func(ctx Context)

// Precommiter runs code during commit immediately before the `deliverState` is written to the `rootMultiStore`.
type Precommiter func(ctx Context)

type RequestDeliverTx struct {
	Tx []byte `protobuf:"bytes,1,opt,name=tx,proto3" json:"tx,omitempty"`
}

type DeliverTxer func(ctx Context, req RequestDeliverTx, tx abci.ExecTxResult)

type BeforeCommitter func(ctx Context)

type AfterCommitter func(ctx Context)

type CreateOracleResultTxHandler func(Context, *abci.RequestCreateOracleResultTx) (*abci.ResponseCreateOracleResultTx, error)

type FetchOracleVotesHandler func(context.Context, *abci.RequestFetchOracleVotes) (*abci.ResponseFetchOracleVotes, error)

type FetchOracleResultsHandler func(context.Context, *abci.RequestFetchOracleResults) (*abci.ResponseFetchOracleResults, error)

type DoesSubAccountBelongToValHandler func(context.Context, *abci.RequestDoesSubAccountBelongToVal) (*abci.ResponseDoesSubAccountBelongToVal, error)

type ValidateOracleVotesHandler func(Context, *abci.RequestValidateOracleVotes) (*abci.ResponseValidateOracleVotes, error)

type MsgHandlerMiddleware func(ctx Context, msg Msg, handler Handler) (*Result, error)

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

// PreBlocker runs code before the `BeginBlocker` and defines a function type alias for executing logic right
// before FinalizeBlock is called (but after its context has been set up). It is
// intended to allow applications to perform computation on vote extensions and
// persist their results in state.
//
// Note: returning an error will make FinalizeBlock fail.
type PreBlocker func(Context, *abci.RequestFinalizeBlock) (*ResponsePreBlock, error)

// BeginBlocker defines a function type alias for executing application
// business logic before transactions are executed.
//
// Note: The BeginBlock ABCI method no longer exists in the ABCI specification
// as of CometBFT v0.38.0. This function type alias is provided for backwards
// compatibility with applications that still use the BeginBlock ABCI method
// and allows for existing BeginBlock functionality within applications.
type BeginBlocker func(Context) (BeginBlock, error)

// EndBlocker defines a function type alias for executing application
// business logic after transactions are executed but before committing.
//
// Note: The EndBlock ABCI method no longer exists in the ABCI specification
// as of CometBFT v0.38.0. This function type alias is provided for backwards
// compatibility with applications that still use the EndBlock ABCI method
// and allows for existing EndBlock functionality within applications.
type EndBlocker func(Context) (EndBlock, error)

// EndBlock defines a type which contains endblock events and validator set updates
type EndBlock struct {
	ValidatorUpdates []abci.ValidatorUpdate
	Events           []abci.Event
}

// BeginBlock defines a type which contains beginBlock events
type BeginBlock struct {
	Events []abci.Event
}

type ResponsePreBlock struct {
	ConsensusParamsChanged bool
	Events                 []abci.Event
}

func (r ResponsePreBlock) IsConsensusParamsChanged() bool {
	return r.ConsensusParamsChanged
}
