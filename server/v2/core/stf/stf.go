package stf

import (
	"context"

	"cosmossdk.io/server/v2/core/appmanager"
	"cosmossdk.io/server/v2/core/store"
	"cosmossdk.io/server/v2/core/transaction"
)

// STF defines the state transition handler used by AppManager to execute
// state transitions over some state. STF never writes to state, instead
// returns the state changes caused by the state transitions.
type STF[T transaction.Tx] interface {
	// DeliverBlock is used to process an entire block, given a state to apply the state transition to.
	// Returns the state changes of the transition.
	DeliverBlock(
		ctx context.Context,
		block *appmanager.BlockRequest[T],
		state store.GetReader,
	) (*appmanager.BlockResponse, store.GetWriter, error)
	// Simulate simulates the execution of a transaction over the provided state, with the provided gas limit.
	Simulate(ctx context.Context, state store.GetReader, gasLimit uint64, tx T) (appmanager.TxResult, store.GetWriter)
	// Query runs the provided query over the provided readonly state.
	Query(ctx context.Context, state store.GetReader, gasLimit uint64, queryRequest appmanager.Type) (queryResponse appmanager.Type, err error)
	// ValidateTx validates the TX.
	ValidateTx(ctx context.Context, state store.GetReader, gasLimit uint64, tx T) appmanager.TxResult
}
