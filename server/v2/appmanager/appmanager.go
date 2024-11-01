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

// GenesisManager is a coordinator for all things related to an application
// It is responsible for interacting with stf and store.
// Runtime/v2 is an extension of this interface.
type GenesisManager[T transaction.Tx] interface {
	// InitGenesis initializes the genesis state of the application.
	InitGenesis(
		ctx context.Context,
		stf StateTransitionFunction[T],
		blockRequest *server.BlockRequest[T],
		initGenesisJSON []byte,
		txDecoder transaction.Codec[T],
	) (*server.BlockResponse, corestore.WriterMap, error)

	// ExportGenesis exports the genesis state of the application.
	ExportGenesis(ctx context.Context, version uint64) ([]byte, error)
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

// genesisManager is a coordinator for all things related to an application
type genesisManager[T transaction.Tx] struct {
	// InitGenesis is a function that initializes the application state from a genesis file.
	// It takes a context, a source reader for the genesis file, and a transaction handler function.
	initGenesis InitGenesis
	// ExportGenesis is a function that exports the application state to a genesis file.
	// It takes a context and a version number for the genesis file.
	exportGenesis ExportGenesis
}

func New[T transaction.Tx](
	initGenesisImpl InitGenesis,
	exportGenesisImpl ExportGenesis,
) GenesisManager[T] {
	return &genesisManager[T]{
		initGenesis:   initGenesisImpl,
		exportGenesis: exportGenesisImpl,
	}
}

// InitGenesis initializes the genesis state of the application.
func (a genesisManager[T]) InitGenesis(
	ctx context.Context,
	stf StateTransitionFunction[T],
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

	blockResponse, blockZeroState, err := stf.DeliverBlock(ctx, blockRequest, genesisState)
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
func (a genesisManager[T]) ExportGenesis(ctx context.Context, version uint64) ([]byte, error) {
	if a.exportGenesis == nil {
		return nil, errors.New("export genesis function not set")
	}

	return a.exportGenesis(ctx, version)
}
