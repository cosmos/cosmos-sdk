package common

import (
	"context"
	common2 "github.com/cosmos/cosmos-sdk/simsx/common"
	"github.com/cosmos/cosmos-sdk/types"
)

// internal data type
type tuple struct {
	signer []common2.SimAccount
	msg    types.Msg
}

// SafeRunFactoryMethod runs the factory method on a separate goroutine to abort early when the context is canceled via reporter skip
func SafeRunFactoryMethod(
	ctx context.Context,
	data *common2.ChainDataSource,
	reporter common2.SimulationReporter,
	f common2.FactoryMethod,
) (signer []common2.SimAccount, msg types.Msg) {
	r := make(chan tuple)
	go func() {
		defer recoverPanicForSkipped(reporter, r)
		signer, msg := f(ctx, data, reporter)
		r <- tuple{signer: signer, msg: msg}
	}()
	select {
	case t, ok := <-r:
		if !ok {
			return nil, nil
		}
		return t.signer, t.msg
	case <-ctx.Done():
		reporter.Skip("context closed")
		return nil, nil
	}
}

func recoverPanicForSkipped(reporter common2.SimulationReporter, resultChan chan tuple) {
	if r := recover(); r != nil {
		if !reporter.IsAborted() {
			panic(r)
		}
		close(resultChan)
	}
}
