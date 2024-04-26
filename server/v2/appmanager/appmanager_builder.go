package appmanager

import (
	"context"
	"encoding/json"
	"io"

	"cosmossdk.io/core/transaction"
	"cosmossdk.io/server/v2/appmanager/store"
	"cosmossdk.io/server/v2/stf"
)

type Builder[T transaction.Tx] struct {
	STF *stf.STF[T]
	DB  store.Store
	ValidateTxGasLimit,
	QueryGasLimit,
	SimulationGasLimit uint64

	InitGenesis func(ctx context.Context, src io.Reader, txHandler func(json.RawMessage) error) error
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
		initGenesis: b.InitGenesis,
		stf:         b.STF,
	}, nil
}
