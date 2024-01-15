package appmanager

import (
	"cosmossdk.io/server/v2/core/appmanager"
	"cosmossdk.io/server/v2/core/store"
	"cosmossdk.io/server/v2/core/transaction"
)

type Builder[T transaction.Tx] struct {
	STF STF[T]
	DB  store.Store
	ValidateTxGasLimit,
	QueryGasLimit,
	SimulationGasLimit uint64
	PrepareBlockHandler appmanager.PrepareHandler[T]
	VerifyBlockHandler  appmanager.ProcessHandler[T]
}

func (b Builder[T]) Build() (*AppManager[T], error) {
	return &AppManager[T]{
		ValidateTxGasLimit: b.ValidateTxGasLimit,
		queryGasLimit:      b.QueryGasLimit,
		simulationGasLimit: b.SimulationGasLimit,
		db:                 b.DB,
		exportState:        nil,
		importState:        nil,
		prepareHandler:     b.PrepareBlockHandler,
		processHandler:     b.VerifyBlockHandler,
		stf:                b.STF,
	}, nil
}
