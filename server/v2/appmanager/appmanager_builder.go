package appmanager

import (
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/server/v2/core/appmanager"
	"cosmossdk.io/server/v2/core/stf"
	"cosmossdk.io/server/v2/core/store"
)

type Builder[T transaction.Tx] struct {
	STF stf.STF[T]
	DB  store.Store
	ValidateTxGasLimit,
	QueryGasLimit,
	SimulationGasLimit uint64
	PrepareBlockHandler appmanager.PrepareHandler[T]
	VerifyBlockHandler  appmanager.ProcessHandler[T]
}

func (b Builder[T]) Build() (*AppManager[T], error) {
	return &AppManager[T]{
<<<<<<< HEAD
		config: Config{
			ValidateTxGasLimit: b.ValidateTxGasLimit,
			queryGasLimit:      b.QueryGasLimit,
			simulationGasLimit: b.SimulationGasLimit,
		},
		db:             b.DB,
		exportState:    nil,
		importState:    nil,
		prepareHandler: b.PrepareBlockHandler,
		processHandler: b.VerifyBlockHandler,
		stf:            b.STF,
||||||| be6720d7be
		ValidateTxGasLimit: b.ValidateTxGasLimit,
		queryGasLimit:      b.QueryGasLimit,
		simulationGasLimit: b.SimulationGasLimit,
		db:                 b.DB,
		exportState:        nil,
		importState:        nil,
		prepareHandler:     b.PrepareBlockHandler,
		processHandler:     b.VerifyBlockHandler,
		stf:                b.STF,
=======
		validateTxGasLimit: b.ValidateTxGasLimit,
		queryGasLimit:      b.QueryGasLimit,
		simulationGasLimit: b.SimulationGasLimit,
		db:                 b.DB,
		exportState:        nil,
		importState:        nil,
		prepareHandler:     b.PrepareBlockHandler,
		processHandler:     b.VerifyBlockHandler,
		stf:                b.STF,
>>>>>>> server_modular
	}, nil
}
