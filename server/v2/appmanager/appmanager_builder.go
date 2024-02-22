package appmanager

import (
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/server/v2/appmanager/store"
	"cosmossdk.io/server/v2/core/stf"
)

type Builder[T transaction.Tx] struct {
	STF stf.STF[T]
	DB  store.Store
	ValidateTxGasLimit,
	QueryGasLimit,
	SimulationGasLimit uint64
}

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
		stf:         b.STF,
	}, nil
}
