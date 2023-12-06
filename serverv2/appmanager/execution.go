package appmanager

import (
	"context"

	"cosmossdk.io/core/appmanager"
	"cosmossdk.io/core/event"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/log"
)

func ExecTx[T transaction.Tx](ctx context.Context, logger log.Logger, tx T, simulate bool) (appmanager.TxResult, error) {
	// gInfo sdk.GasInfo, result *sdk.Result, anteEvents []event.Event,
	return appmanager.TxResult{}, nil
}

// BeginBlocker is a function type alias for the begin blocker used in the
func BeginBlock() ([]event.Event, error) {
	panic("implement me")
}

func EndBlock() ([]event.Event, error) {
	panic("implement me")
}
