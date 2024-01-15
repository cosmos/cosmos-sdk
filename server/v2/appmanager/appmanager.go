package appmanager

import (
	"context"
	"fmt"
	"io"

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
		state store.ReadonlyState,
	) (*appmanager.BlockResponse, store.WritableState, error)
	// Simulate simulates the execution of a transaction over the provided state, with the provided gas limit.
	// TODO: Might be useful to return the state changes caused by the TX.
	Simulate(ctx context.Context, state store.ReadonlyState, gasLimit uint64, tx T) appmanager.TxResult
	// Query runs the provided query over the provided readonly state.
	Query(ctx context.Context, state store.ReadonlyState, gasLimit uint64, queryRequest Type) (queryResponse Type, err error)
	// ValidateTx validates the TX.
	ValidateTx(ctx context.Context, state store.ReadonlyState, gasLimit uint64, tx T) appmanager.TxResult
}

// AppManager is a coordinator for all things related to an application
type AppManager[T transaction.Tx] struct {
	// configs
	ValidateTxGasLimit uint64
	queryGasLimit      uint64
	simulationGasLimit uint64
	// configs - end

	db store.Store

	exportState func(ctx context.Context, dst map[string]io.Writer) error
	importState func(ctx context.Context, src map[string]io.Reader) error

	prepareHandler appmanager.PrepareHandler[T]
	processHandler appmanager.ProcessHandler[T]
	stf            STF[T] // consider if instead of having an interface (which is boxed?), we could have another type Parameter defining STF.
}

// BuildBlock builds a block when requested by consensus. It will take in the total size txs to be included and return a list of transactions
func (a AppManager[T]) BuildBlock(ctx context.Context, height uint64, txs []T) ([]T, error) {
	latestVersion, currentState, err := a.db.StateLatest()
	if err != nil {
		return nil, fmt.Errorf("unable to create new state for height %d: %w", height, err)
	}

	if latestVersion+1 != height {
		return nil, fmt.Errorf("invalid BuildBlock height wanted %d, got %d", latestVersion+1, height)
	}

	txs, err = a.prepareHandler(ctx, txs, currentState)
	if err != nil {
		return nil, err
	}

	return txs, nil
}

func (a AppManager[T]) VerifyBlock(ctx context.Context, height uint64, txs []T) error {
	latestVersion, currentState, err := a.db.StateLatest()
	if err != nil {
		return fmt.Errorf("unable to create new state for height %d: %w", height, err)
	}

	if latestVersion+1 != height {
		return fmt.Errorf("invalid VerifyBlock height wanted %d, got %d", latestVersion+1, height)
	}

	err = a.processHandler(ctx, txs, currentState)
	if err != nil {
		return err
	}

	return nil
}

func (a AppManager[T]) DeliverBlock(ctx context.Context, block *appmanager.BlockRequest[T]) (*appmanager.BlockResponse, []store.ChangeSet, error) {
	latestVersion, currentState, err := a.db.StateLatest()
	if err != nil {
		return nil, nil, fmt.Errorf("unable to create new state for height %d: %w", block.Height, err)
	}

	if latestVersion+1 != block.Height {
		return nil, nil, fmt.Errorf("invalid DeliverBlock height wanted %d, got %d", latestVersion+1, block.Height)
	}

	blockResponse, newState, err := a.stf.DeliverBlock(ctx, block, currentState)
	if err != nil {
		return nil, nil, fmt.Errorf("block delivery failed: %w", err)
	}

	newStateChanges, err := newState.ChangeSets()
	if err != nil {
		return nil, nil, fmt.Errorf("change set: %w", err)
	}

	return blockResponse, newStateChanges, nil
}

// ValidateTx will validate the tx against the latest storage state. This means that
// only the stateful validation will be run, not the execution portion of the tx.
// If full execution is needed, Simulate must be used.
func (a AppManager[T]) ValidateTx(ctx context.Context, tx T) (appmanager.TxResult, error) {
	_, latestState, err := a.db.StateLatest()
	if err != nil {
		return appmanager.TxResult{}, err
	}
	return a.stf.ValidateTx(ctx, latestState, a.ValidateTxGasLimit, tx), nil
}

// Simulate runs validation and execution flow of a Tx.
func (a AppManager[T]) Simulate(ctx context.Context, tx T) (appmanager.TxResult, error) {
	_, state, err := a.db.StateLatest()
	if err != nil {
		return appmanager.TxResult{}, err
	}
	result := a.stf.Simulate(ctx, state, a.simulationGasLimit, tx)
	return result, nil
}

// Query queries the application at the provided version.
func (a AppManager[T]) Query(ctx context.Context, version uint64, request Type) (response Type, err error) {
	// if version is provided attempt to do a heightened query.
	if version != 0 {
		queryState, err := a.db.StateAt(version)
		if err != nil {
			return nil, err
		}
		return a.stf.Query(ctx, queryState, a.queryGasLimit, request)
	}
	// otherwise rely on latest available state.
	_, queryState, err := a.db.StateLatest()
	if err != nil {
		return nil, err
	}
	return a.stf.Query(ctx, queryState, a.queryGasLimit, request)
}
