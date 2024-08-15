package appmanager

import (
	"cosmossdk.io/core/transaction"
)

// Builder is a struct that represents the application builder for managing transactions.
// It contains various fields and methods for initializing the application and handling transactions.
type Builder[T transaction.Tx] struct {
	STF StateTransitionFunction[T] // The state transition function for processing transactions.
	DB  Store                      // The database for storing application data.

	// Gas limits for validating, querying, and simulating transactions.
	ValidateTxGasLimit uint64
	QueryGasLimit      uint64
	SimulationGasLimit uint64

	// InitGenesis is a function that initializes the application state from a genesis file.
	// It takes a context, a source reader for the genesis file, and a transaction handler function.
	InitGenesis InitGenesis
	// ExportGenesis is a function that exports the application state to a genesis file.
	// It takes a context and a version number for the genesis file.
	ExportGenesis ExportGenesis
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
		db:            b.DB,
		initGenesis:   b.InitGenesis,
		exportGenesis: b.ExportGenesis,
		stf:           b.STF,
	}, nil
}
