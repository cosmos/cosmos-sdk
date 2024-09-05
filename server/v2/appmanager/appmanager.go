package appmanager

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"cosmossdk.io/core/server"
	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/core/transaction"
)

// Store defines the underlying storage behavior needed by AppManager.
type Store interface {
	// StateLatest returns a readonly view over the latest
	// committed state of the store. Alongside the version
	// associated with it.
	StateLatest() (uint64, corestore.ReaderMap, error)

	// StateAt returns a readonly view over the provided
	// state. Must error when the version does not exist.
	StateAt(version uint64) (corestore.ReaderMap, error)
}

// AppManager is a coordinator for all things related to an application
type AppManager[T transaction.Tx] struct {
	config Config

	db Store

	initGenesis   InitGenesis
	exportGenesis ExportGenesis

	stf StateTransitionFunction[T]
}

// InitGenesis initializes the genesis state of the application.
func (a AppManager[T]) InitGenesis(
	ctx context.Context,
	blockRequest *server.BlockRequest[T],
	initGenesisJSON []byte,
	txDecoder transaction.Codec[T],
) (*server.BlockResponse, corestore.WriterMap, error) {
	v, zeroState, err := a.db.StateLatest()
	if err != nil {
		return nil, nil, fmt.Errorf("unable to get latest state: %w", err)
	}
	if v != 0 { // TODO: genesis state may be > 0, we need to set version on store
		return nil, nil, errors.New("cannot init genesis on non-zero state")
	}

	var genTxs []T
	genesisState, err := a.stf.RunWithCtx(ctx, zeroState, func(ctx context.Context) error {
		return a.initGenesis(ctx, bytes.NewBuffer(initGenesisJSON), func(jsonTx json.RawMessage) error {
			genTx, err := txDecoder.DecodeJSON(jsonTx)
			if err != nil {
				return fmt.Errorf("failed to decode genesis transaction: %w", err)
			}
			genTxs = append(genTxs, genTx)
			return nil
		})
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to import genesis state: %w", err)
	}
	// run block
	blockRequest.Txs = genTxs

	blockResponse, blockZeroState, err := a.stf.DeliverBlock(ctx, blockRequest, genesisState)
	if err != nil {
		return blockResponse, nil, fmt.Errorf("failed to deliver block %d: %w", blockRequest.Height, err)
	}

	// after executing block 0, we extract the changes and apply them to the genesis state.
	stateChanges, err := blockZeroState.GetStateChanges()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get block zero state changes: %w", err)
	}

	err = genesisState.ApplyStateChanges(stateChanges)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to apply block zero state changes to genesis state: %w", err)
	}

	return blockResponse, genesisState, err
}

// ExportGenesis exports the genesis state of the application.
func (a AppManager[T]) ExportGenesis(ctx context.Context, version uint64) ([]byte, error) {
	zeroState, err := a.db.StateAt(version)
	if err != nil {
		return nil, fmt.Errorf("unable to get latest state: %w", err)
	}

	bz := make([]byte, 0)
	_, err = a.stf.RunWithCtx(ctx, zeroState, func(ctx context.Context) error {
		if a.exportGenesis == nil {
			return errors.New("export genesis function not set")
		}

		bz, err = a.exportGenesis(ctx, version)
		if err != nil {
			return fmt.Errorf("failed to export genesis state: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to export genesis state: %w", err)
	}

	return bz, nil
}

func (a AppManager[T]) DeliverBlock(
	ctx context.Context,
	block *server.BlockRequest[T],
) (*server.BlockResponse, corestore.WriterMap, error) {
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
func (a AppManager[T]) ValidateTx(ctx context.Context, tx T) (server.TxResult, error) {
	_, latestState, err := a.db.StateLatest()
	if err != nil {
		return server.TxResult{}, err
	}
	res := a.stf.ValidateTx(ctx, latestState, a.config.ValidateTxGasLimit, tx)
	return res, res.Error
}

// Simulate runs validation and execution flow of a Tx.
func (a AppManager[T]) Simulate(ctx context.Context, tx T) (server.TxResult, corestore.WriterMap, error) {
	_, state, err := a.db.StateLatest()
	if err != nil {
		return server.TxResult{}, nil, err
	}
	result, cs := a.stf.Simulate(ctx, state, a.config.SimulationGasLimit, tx) // TODO: check if this is done in the antehandler
	return result, cs, nil
}

// Query queries the application at the provided version.
// CONTRACT: Version must always be provided, if 0, get latest
func (a AppManager[T]) Query(ctx context.Context, version uint64, request transaction.Msg) (transaction.Msg, error) {
	// if version is provided attempt to do a height query.
	if version != 0 {
		queryState, err := a.db.StateAt(version)
		if err != nil {
			return nil, err
		}
		return a.stf.Query(ctx, queryState, a.config.QueryGasLimit, request)
	}

	// otherwise rely on latest available state.
	_, queryState, err := a.db.StateLatest()
	if err != nil {
		return nil, err
	}
	return a.stf.Query(ctx, queryState, a.config.QueryGasLimit, request)
}

// QueryWithState executes a query with the provided state. This allows to process a query
// independently of the db state. For example, it can be used to process a query with temporary
// and uncommitted state
func (a AppManager[T]) QueryWithState(
	ctx context.Context,
	state corestore.ReaderMap,
	request transaction.Msg,
) (transaction.Msg, error) {
	return a.stf.Query(ctx, state, a.config.QueryGasLimit, request)
}
