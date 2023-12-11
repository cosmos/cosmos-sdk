package appmanager

import (
	"context"

	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/serverv2/core/appmanager"
	"github.com/cosmos/cosmos-sdk/serverv2/core/event"
	"github.com/cosmos/cosmos-sdk/serverv2/core/transaction"
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
