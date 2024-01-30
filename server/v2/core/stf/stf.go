package stf

import (
	"context"
	"errors"
	"math"

	"cosmossdk.io/core/transaction"
	"cosmossdk.io/server/v2/core/appmanager"
	"cosmossdk.io/server/v2/core/store"
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
		state store.ReaderMap,
	) (*appmanager.BlockResponse, store.WriterMap, error)
	// Simulate simulates the execution of a transaction over the provided state, with the provided gas limit.
	Simulate(ctx context.Context, state store.ReaderMap, gasLimit uint64, tx T) (appmanager.TxResult, store.WriterMap)
	// Query runs the provided query over the provided readonly state.
	Query(ctx context.Context, state store.ReaderMap, gasLimit uint64, queryRequest appmanager.Type) (queryResponse appmanager.Type, err error)
	// ValidateTx validates the TX.
	ValidateTx(ctx context.Context, state store.ReaderMap, gasLimit uint64, tx T) appmanager.TxResult
}

// ErrOutOfGas must be used by GasMeter implementers to signal
// that the state transition consumed all the allowed computational
// gas.
var ErrOutOfGas = errors.New("out of gas")

// Gas defines type alias of uint64 for gas consumption. Gas is used
// to measure computational overhead when executing state transitions,
// it might be related to storage access and not only.
type Gas = uint64

// NoGasLimit signals that no gas limit must be applied.
const NoGasLimit Gas = math.MaxUint64

// GasMeter defines an interface for gas consumption tracking.
type GasMeter interface {
	// GasConsumed returns the amount of gas consumed so far.
	GasConsumed() Gas
	// Limit returns the gas limit (if any).
	Limit() Gas
	// ConsumeGas adds the given amount of gas to the gas consumed and should error
	// if it overflows the gas limit (if any).
	ConsumeGas(amount Gas, descriptor string) error
	// RefundGas will deduct the given amount from the gas consumed so far. If the
	// amount is greater than the gas consumed, the function should error.
	RefundGas(amount Gas, descriptor string) error
}
