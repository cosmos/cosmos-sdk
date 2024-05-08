package appmanager

import (
	"context"
	"encoding/json"
	"io"

	"cosmossdk.io/core/transaction"
	"cosmossdk.io/server/v2/appmanager/store"
	"cosmossdk.io/server/v2/stf"
)

// Builder is a struct that represents the application builder for managing transactions.
// It contains various fields and methods for initializing the application and handling transactions.
type Builder[T transaction.Tx] struct {
	STF *stf.STF[T] // The state transition function for processing transactions.
	DB  store.Store // The database for storing application data.

	// Gas limits for validating, querying, and simulating transactions.
	ValidateTxGasLimit uint64
	QueryGasLimit      uint64
	SimulationGasLimit uint64

	// InitGenesis is a function that initializes the application state from a genesis file.
	// It takes a context, a source reader for the genesis file, and a transaction handler function.
	InitGenesis func(ctx context.Context, src io.Reader, txHandler func(json.RawMessage) error) error
}

// Build creates a new instance of AppManager with the provided configuration and returns it.
// It initializes the AppManager with the given database, export state, import state, initGenesis function, and state transition function.
func (b Builder[T]) Build() (*AppManager[T], error) {
	return &AppManager[T]{
		config: Config{
			ValidateTxGasLimit: b.ValidateTxGasLimit,
			QueryGasLimit:      b.QueryGasLimit,
			SimulationGasLimit: b.SimulationGasLimit,
		},
		db:          b.DB,
		exportState: nil,
		importState: nil,
		initGenesis: b.InitGenesis,
		stf:         b.STF,
	}, nil
}
