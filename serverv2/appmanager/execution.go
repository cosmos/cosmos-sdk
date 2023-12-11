package appmanager

import (
	"context"

	"cosmossdk.io/log"
	"github.com/cosmos/gogoproto/proto"

	"github.com/cosmos/cosmos-sdk/serverv2/core/appmanager"
	"github.com/cosmos/cosmos-sdk/serverv2/core/event"
	"github.com/cosmos/cosmos-sdk/serverv2/core/transaction"
)

func ExecTx[T transaction.Tx](ctx context.Context, logger log.Logger, tx T) (appmanager.TxResult, error) {
	// gInfo sdk.GasInfo, result *sdk.Result, anteEvents []event.Event,

	// check execution mode

	// preExecution Hook

	// execute the transaction

	// postExecution Hook

	panic("implement me")
}

// ExecMsg executes a single message. It returns the result of the message execution and an error if the state transition fails,
// this is used for the internal router for modules to communicate with each other
func ExecMsg(ctx context.Context, logger log.Logger, msg proto.Message) (appmanager.TxResult, error) {
	_, err := PreExec(ctx, msg)
	if err != nil {
		return appmanager.TxResult{}, err
	}

	// handler for the message

	_, err = PostExec(ctx, msg)
	if err != nil {
		return appmanager.TxResult{}, err
	}

	panic("implement me")
}

func PreExec(ctx context.Context, msg proto.Message) (appmanager.TxResult, error) {
	panic("implement me")
}

func PostExec(ctx context.Context, msg proto.Message) (appmanager.TxResult, error) {
	panic("implement me")
}

// BeginBlocker is a function type alias for the begin blocker used in the
func BeginBlock() ([]event.Event, error) {
	panic("implement me")
}

func EndBlock() ([]event.Event, error) {
	panic("implement me")
}
