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

// AppManager is a coordinator for all things related to an application
// It is responsible for interacting with stf and store.
// Runtime/v2 is an extension of this interface.
type AppManager[T transaction.Tx] interface {
	// InitGenesis initializes the genesis state of the application.
	InitGenesis(
		ctx context.Context,
		blockRequest *server.BlockRequest[T],
		initGenesisJSON []byte,
		txDecoder transaction.Codec[T],
	) (*server.BlockResponse, corestore.WriterMap, error)

	// ExportGenesis exports the genesis state of the application.
	ExportGenesis(ctx context.Context, version uint64) ([]byte, error)

	// DeliverBlock executes a block of transactions.
	DeliverBlock(
		ctx context.Context,
		block *server.BlockRequest[T],
	) (*server.BlockResponse, corestore.WriterMap, error)

	// ValidateTx will validate the tx against the latest storage state. This means that
	// only the stateful validation will be run, not the execution portion of the tx.
	// If full execution is needed, Simulate must be used.
	ValidateTx(ctx context.Context, tx T) (server.TxResult, error)

	// Simulate runs validation and execution flow of a Tx.
	Simulate(ctx context.Context, tx T) (server.TxResult, corestore.WriterMap, error)

	// SimulateWithState runs validation and execution flow of a Tx,
	// using the provided state instead of loading the latest state from the underlying database.
	SimulateWithState(ctx context.Context, state corestore.ReaderMap, tx T) (server.TxResult, corestore.WriterMap, error)

	// Query queries the application at the provided version.
	// CONTRACT: Version must always be provided, if 0, get latest
	Query(ctx context.Context, version uint64, request transaction.Msg) (transaction.Msg, error)

	// QueryWithState executes a query with the provided state. This allows to process a query
	// independently of the db state. For example, it can be used to process a query with temporary
	// and uncommitted state
	QueryWithState(ctx context.Context, state corestore.ReaderMap, request transaction.Msg) (transaction.Msg, error)
}

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

// appManager is a coordinator for all things related to an application
type appManager[T transaction.Tx] struct {
	// Gas limits for validating, querying, and simulating transactions.
	config Config
	// InitGenesis is a function that initializes the application state from a genesis file.
	// It takes a context, a source reader for the genesis file, and a transaction handler function.
	initGenesis InitGenesis
	// ExportGenesis is a function that exports the application state to a genesis file.
	// It takes a context and a version number for the genesis file.
	exportGenesis ExportGenesis
	// The database for storing application data.
	db Store
	// The state transition function for processing transactions.
	stf StateTransitionFunction[T]
}

func New[T transaction.Tx](
	config Config,
	db Store,
	stf StateTransitionFunction[T],
	initGenesisImpl InitGenesis,
	exportGenesisImpl ExportGenesis,
) AppManager[T] {
	return &appManager[T]{
		config:        config,
		db:            db,
		stf:           stf,
		initGenesis:   initGenesisImpl,
		exportGenesis: exportGenesisImpl,
	}
}

// InitGenesis initializes the genesis state of the application.
func (a appManager[T]) InitGenesis(
	ctx context.Context,
	blockRequest *server.BlockRequest[T],
	initGenesisJSON []byte,
	txDecoder transaction.Codec[T],
) (*server.BlockResponse, corestore.WriterMap, error) {
	var genTxs []T
	genesisState, err := a.initGenesis(
		ctx,
		bytes.NewBuffer(initGenesisJSON),
		func(jsonTx json.RawMessage) error {
			genTx, err := txDecoder.DecodeJSON(jsonTx)
			if err != nil {
				return fmt.Errorf("failed to decode genesis transaction: %w", err)
			}
			genTxs = append(genTxs, genTx)
			return nil
		},
	)
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
func (a appManager[T]) ExportGenesis(ctx context.Context, version uint64) ([]byte, error) {
	if a.exportGenesis == nil {
		return nil, errors.New("export genesis function not set")
	}

	return a.exportGenesis(ctx, version)
}

// DeliverBlock executes a block of transactions.
func (a appManager[T]) DeliverBlock(
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
func (a appManager[T]) ValidateTx(ctx context.Context, tx T) (server.TxResult, error) {
	_, latestState, err := a.db.StateLatest()
	if err != nil {
		return server.TxResult{}, err
	}
	res := a.stf.ValidateTx(ctx, latestState, a.config.ValidateTxGasLimit, tx)
	return res, res.Error
}

// Simulate runs validation and execution flow of a Tx.
func (a appManager[T]) Simulate(ctx context.Context, tx T) (server.TxResult, corestore.WriterMap, error) {
	_, state, err := a.db.StateLatest()
	if err != nil {
		return server.TxResult{}, nil, err
	}
	result, cs := a.stf.Simulate(ctx, state, a.config.SimulationGasLimit, tx) // TODO: check if this is done in the antehandler
	return result, cs, nil
}

// SimulateWithState runs validation and execution flow of a Tx,
// using the provided state instead of loading the latest state from the underlying database.
func (a appManager[T]) SimulateWithState(ctx context.Context, state corestore.ReaderMap, tx T) (server.TxResult, corestore.WriterMap, error) {
	result, cs := a.stf.Simulate(ctx, state, a.config.SimulationGasLimit, tx) // TODO: check if this is done in the antehandler
	return result, cs, nil
}

// Query queries the application at the provided version.
// CONTRACT: Version must always be provided, if 0, get latest
func (a appManager[T]) Query(ctx context.Context, version uint64, request transaction.Msg) (transaction.Msg, error) {
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
func (a appManager[T]) QueryWithState(ctx context.Context, state corestore.ReaderMap, request transaction.Msg) (transaction.Msg, error) {
	return a.stf.Query(ctx, state, a.config.QueryGasLimit, request)
}
