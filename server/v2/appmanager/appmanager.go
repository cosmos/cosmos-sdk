package appmanager

import (
	"context"
	"fmt"
	"io"

	"cosmossdk.io/core/transaction"
	"cosmossdk.io/server/v2/core/appmanager"
	"cosmossdk.io/server/v2/core/stf"
	"cosmossdk.io/server/v2/core/store"
)

// AppManager is a coordinator for all things related to an application
type AppManager[T transaction.Tx] struct {
	// configs - begin
	validateTxGasLimit uint64
	queryGasLimit      uint64
	simulationGasLimit uint64
	// configs - end

	db store.Store

	exportState func(ctx context.Context, dst map[string]io.Writer) error
	importState func(ctx context.Context, src map[string]io.Reader) error

	prepareHandler appmanager.PrepareHandler[T]
	processHandler appmanager.ProcessHandler[T]
	stf            stf.STF[T] // consider if instead of having an interface (which is boxed?), we could have another type Parameter defining STF.
}

// TODO add initgenesis and figure out how to use it
// TODO make sure to provide defaults for handlers and configs
// TODO: handle multimessage txs

// BuildBlock builds a block when requested by consensus. It will take in the total size txs to be included and return a list of transactions
func (a AppManager[T]) BuildBlock(ctx context.Context, height, maxBlockBytes uint64) ([]T, error) {
	latestVersion, currentState, err := a.db.StateLatest()
	if err != nil {
		return nil, fmt.Errorf("unable to create new state for height %d: %w", height, err)
	}

	if latestVersion+1 != height {
		return nil, fmt.Errorf("invalid BuildBlock height wanted %d, got %d", latestVersion+1, height)
	}

	txs, err := a.prepareHandler(ctx, currentState)
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

func (a AppManager[T]) DeliverBlock(ctx context.Context, block *appmanager.BlockRequest[T]) (*appmanager.BlockResponse, store.WriterMap, error) {
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

	return blockResponse, newState, nil
}

// ValidateTx will validate the tx against the latest storage state. This means that
// only the stateful validation will be run, not the execution portion of the tx.
// If full execution is needed, Simulate must be used.
func (a AppManager[T]) ValidateTx(ctx context.Context, tx T) (appmanager.TxResult, error) {
	_, latestState, err := a.db.StateLatest()
	if err != nil {
		return appmanager.TxResult{}, err
	}
	return a.stf.ValidateTx(ctx, latestState, a.validateTxGasLimit, tx), nil
}

// Simulate runs validation and execution flow of a Tx.
func (a AppManager[T]) Simulate(ctx context.Context, tx T) (appmanager.TxResult, store.WriterMap, error) {
	_, state, err := a.db.StateLatest()
	if err != nil {
		return appmanager.TxResult{}, nil, err
	}
	result, cs := a.stf.Simulate(ctx, state, a.simulationGasLimit, tx)
	return result, cs, nil
}

// Query queries the application at the provided version.
// CONTRACT: Version must always be provided, if 0, get latest
func (a AppManager[T]) Query(ctx context.Context, version uint64, request appmanager.Type) (response appmanager.Type, err error) {
	// if version is provided attempt to do a height query.
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

// QueryWithState executes a query with the provided state. This allows to process a query
// independently of the db state. For example, it can be used to process a query with temporary
// and uncommitted state
func (a AppManager[T]) QueryWithState(ctx context.Context, state store.ReaderMap, request appmanager.Type) (appmanager.Type, error) {
	return a.stf.Query(ctx, state, a.queryGasLimit, request)
}
