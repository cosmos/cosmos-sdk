package appmanager

import (
	"cosmossdk.io/server/v2/core/appmanager"
	"cosmossdk.io/server/v2/core/mempool"
	"cosmossdk.io/server/v2/core/store"
	"cosmossdk.io/server/v2/core/transaction"
)

type Builder[T transaction.Tx] struct {
	STF STF[T]
	DB  store.Store
	ValidateTxGasLimit,
	QueryGasLimit,
	SimulationGasLimit uint64
	Mempool             mempool.Mempool[T]
	PrepareBlockHandler appmanager.PrepareHandler[T]
	VerifyBlockHandler  appmanager.ProcessHandler[T]
}

func (b Builder[T]) Build() (AppManager[T], error) {
	return AppManager[T]{
		ValidateTxGasLimit: b.ValidateTxGasLimit,
		queryGasLimit:      b.QueryGasLimit,
		simulationGasLimit: b.SimulationGasLimit,
		db:                 b.DB,
		mempool:            b.Mempool,
		exportState:        nil,
		importState:        nil,
		prepareHandler:     b.PrepareBlockHandler,
		processHandler:     b.VerifyBlockHandler,
		stf:                b.STF,
	}, nil
}
