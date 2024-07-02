package appmanager

import (
	"context"

	appmanager "cosmossdk.io/core/app"
	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/core/transaction"
)

// AppManager is an interface that defines the methods for managing an application.
type AppManager[T transaction.Tx] interface {
	// InitGenesis initializes the genesis state of the application.
	// It takes a context, a block request, the genesis JSON, and a transaction decoder.
	// It returns a block response, a writer map, and an error if any.
	InitGenesis(
		ctx context.Context,
		blockRequest *appmanager.BlockRequest[T],
		initGenesisJSON []byte,
		txDecoder transaction.Codec[T],
	) (*appmanager.BlockResponse, corestore.WriterMap, error)
	// ExportGenesis exports the genesis state of the application.
	// It takes a context and a version number.
	// It returns the genesis state as a byte slice and an error if any.
	ExportGenesis(ctx context.Context, version uint64) ([]byte, error)
	// DeliverBlock processes a block of transactions.
	// It takes a context and a block request.
	// It returns a block response, a writer map, and an error if any.
	DeliverBlock(
		ctx context.Context,
		block *appmanager.BlockRequest[T],
	) (*appmanager.BlockResponse, corestore.WriterMap, error)
	// ValidateTx validates a transaction against the latest storage state.
	// It takes a context and a transaction.
	// It returns a transaction result and an error if any.
	ValidateTx(ctx context.Context, tx T) (appmanager.TxResult, error)
	// Simulate runs validation and execution flow of a transaction.
	// It takes a context and a transaction.
	// It returns a transaction result, a writer map, and an error if any.
	Simulate(ctx context.Context, tx T) (appmanager.TxResult, corestore.WriterMap, error)
	// Query queries the application at the provided version.
	// It takes a context, a version number, and a request message.
	// It returns a response message and an error if any.
	Query(ctx context.Context, version uint64, request transaction.Msg) (transaction.Msg, error)
	// QueryWithState executes a query with the provided state.
	// It takes a context, a state, and a request message.
	// It returns a response message and an error if any.
	QueryWithState(
		ctx context.Context,
		state corestore.ReaderMap,
		request transaction.Msg,
	) (transaction.Msg, error)
}
