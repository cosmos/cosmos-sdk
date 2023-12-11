package appmanager

import (
	"context"

	"cosmossdk.io/log"
	"github.com/cosmos/gogoproto/proto"

	"github.com/cosmos/cosmos-sdk/serverv2/core/appmanager"
	"github.com/cosmos/cosmos-sdk/serverv2/core/event"
	"github.com/cosmos/cosmos-sdk/serverv2/core/transaction"
)

// ExecTx executes a transaction. It returns the result of the transaction execution and an error if the state transition fails.
// This is meant to be called in three different contexts:
// 1. DeliverTx: the transaction is executed and the state transition is persisted
// 2. SimulateTx: the transaction is executed and the state transition is not persisted
// 3. Prepare/Process: the transaction is executed and the state transition is not persisted, but cached
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
// This is meant to be called in two different contexts:
// 1. ExternalMsg: the message being executed is from a transaction
// 2. InternalMsg: the message being executed is from another module
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
