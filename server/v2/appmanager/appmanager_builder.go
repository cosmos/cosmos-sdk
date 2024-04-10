package appmanager

import (
	"context"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/server/v2/appmanager/store"
	"cosmossdk.io/server/v2/stf"
	"io"
)

type Builder[T transaction.Tx] struct {
	STF *stf.STF[T]
	DB  store.Store

	ValidateTxGasLimit,
	QueryGasLimit,
	SimulationGasLimit uint64

	ImportState func(ctx context.Context, src io.Reader) error
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
		importState: b.ImportState,
		stf:         b.STF,
	}, nil
}
