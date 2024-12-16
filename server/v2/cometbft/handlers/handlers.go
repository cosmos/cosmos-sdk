package handlers

import (
	"context"

	abci "github.com/cometbft/cometbft/api/cometbft/abci/v1"

	"cosmossdk.io/core/server"
	"cosmossdk.io/core/store"
	"cosmossdk.io/core/transaction"
)

type (
	// PrepareHandler passes in the list of Txs that are being proposed. The app can then do stateful operations
	// over the list of proposed transactions. It can return a modified list of txs to include in the proposal.
	PrepareHandler[T transaction.Tx] func(ctx context.Context, app AppManager[T], cdc transaction.Codec[T], req *abci.PrepareProposalRequest, chainID string) ([]T, error)

	// ProcessHandler is a function that takes a list of transactions and returns a boolean and an error.
	// If the verification of a transaction fails, the boolean is false and the error is non-nil.
	ProcessHandler[T transaction.Tx] func(ctx context.Context, app AppManager[T], cdc transaction.Codec[T], req *abci.ProcessProposalRequest, chainID string) error

	// VerifyVoteExtensionHandler is a function type that handles the verification of a vote extension request.
	// It takes a context, a store reader map, and a request to verify a vote extension.
	// It returns a response to verify the vote extension and an error if any.
	VerifyVoteExtensionHandler func(context.Context, store.ReaderMap, *abci.VerifyVoteExtensionRequest) (*abci.VerifyVoteExtensionResponse, error)

	// ExtendVoteHandler is a function type that handles the extension of a vote.
	// It takes a context, a store reader map, and a request to extend a vote.
	// It returns a response to extend the vote and an error if any.
	ExtendVoteHandler func(context.Context, store.ReaderMap, *abci.ExtendVoteRequest) (*abci.ExtendVoteResponse, error)

	// CheckTxHandler is a function type that handles the execution of a transaction.
	CheckTxHandler[T transaction.Tx] func(func(ctx context.Context, tx T) (server.TxResult, error)) (*abci.CheckTxResponse, error)
)
