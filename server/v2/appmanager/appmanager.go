package appmanager

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	appmanager "cosmossdk.io/core/app"
	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/server/v2/appmanager/store"
	"cosmossdk.io/server/v2/stf"
)

// AppManager is a coordinator for all things related to an application
// TODO: add exportGenesis function
type AppManager[T transaction.Tx] struct {
	config Config

	db store.Store

	exportState func(ctx context.Context, dst map[string]io.Writer) error
	importState func(ctx context.Context, src map[string]io.Reader) error //TODO: possibly remove?

	initGenesis func(ctx context.Context, state io.Reader, txHandler func(tx json.RawMessage) error) error

	stf *stf.STF[T] // TODO: define interface
}

func (a AppManager[T]) InitGenesis(
	ctx context.Context,
	blockRequest *appmanager.BlockRequest[T],
	initGenesisJSON []byte,
	txDecoder transaction.Codec[T],
) (*appmanager.BlockResponse, corestore.WriterMap, error) {
	v, zeroState, err := a.db.StateLatest()
	if err != nil {
		return nil, nil, fmt.Errorf("unable to get latest state: %w", err)
	}
	if v != 0 { // TODO: genesis state may be > 0, we need to set version on store
		return nil, nil, fmt.Errorf("cannot init genesis on non-zero state")
	}

	var genTxs []T
	zeroState, err = a.stf.RunWithCtx(ctx, zeroState, func(ctx context.Context) error {
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
	// TODO: in an ideal world, genesis state is simply an initial state being applied
	// unaware of what that state means in relation to every other, so here we can
	// chain genesis
	blockRequest.Txs = genTxs

	blockresponse, genesisState, err := a.stf.DeliverBlock(ctx, blockRequest, zeroState)
	if err != nil {
		return blockresponse, nil, fmt.Errorf("failed to deliver block %s: %w", blockRequest.Height, err)
	}

	return blockresponse, genesisState, err
	// consensus server will need to set the version of the store
}

func (a AppManager[T]) DeliverBlock(
	ctx context.Context,
	block *appmanager.BlockRequest[T],
) (*appmanager.BlockResponse, corestore.WriterMap, error) {
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
	return a.stf.ValidateTx(ctx, latestState, a.config.ValidateTxGasLimit, tx), nil
}

// Simulate runs validation and execution flow of a Tx.
func (a AppManager[T]) Simulate(ctx context.Context, tx T) (appmanager.TxResult, corestore.WriterMap, error) {
	_, state, err := a.db.StateLatest()
	if err != nil {
		return appmanager.TxResult{}, nil, err
	}
	result, cs := a.stf.Simulate(ctx, state, a.config.SimulationGasLimit, tx) // TODO: check if this is done in the antehandler
	return result, cs, nil
}

// Query queries the application at the provided version.
// CONTRACT: Version must always be provided, if 0, get latest
func (a AppManager[T]) Query(ctx context.Context, version uint64, request transaction.Type) (transaction.Type, error) {
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
	request transaction.Type,
) (transaction.Type, error) {
	return a.stf.Query(ctx, state, a.config.QueryGasLimit, request)
}
