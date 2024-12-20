package module

import (
	"context"
	"github.com/cosmos/cosmos-sdk/simsx/common"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ common.SimMsgFactoryX = &ResultHandlingSimMsgFactory[sdk.Msg]{}

// ResultHandlingSimMsgFactory message factory with a delivery error result handler configured.
type ResultHandlingSimMsgFactory[T sdk.Msg] struct {
	SimMsgFactoryFn[T]
	resultHandler common.SimDeliveryResultHandler
}

// NewSimMsgFactoryWithDeliveryResultHandler constructor
func NewSimMsgFactoryWithDeliveryResultHandler[T sdk.Msg](f common.FactoryMethodWithDeliveryResultHandler[T]) *ResultHandlingSimMsgFactory[T] {
	r := &ResultHandlingSimMsgFactory[T]{
		resultHandler: expectNoError,
	}
	r.SimMsgFactoryFn = func(ctx context.Context, testData *common.ChainDataSource, reporter common.SimulationReporter) (signer []common.SimAccount, msg T) {
		signer, msg, r.resultHandler = f(ctx, testData, reporter)
		if r.resultHandler == nil {
			r.resultHandler = expectNoError
		}
		return
	}
	return r
}

// DeliveryResultHandler result handler of the last msg factory invocation
func (f ResultHandlingSimMsgFactory[T]) DeliveryResultHandler() common.SimDeliveryResultHandler {
	return f.resultHandler
}

var (
	_ common.SimMsgFactoryX       = &LazyStateSimMsgFactory[sdk.Msg]{}
	_ common.HasFutureOpsRegistry = &LazyStateSimMsgFactory[sdk.Msg]{}
)

// LazyStateSimMsgFactory stateful message factory with weighted proposals and future operation
// registry initialized lazy before execution.
type LazyStateSimMsgFactory[T sdk.Msg] struct {
	SimMsgFactoryFn[T]
	fsOpsReg common.FutureOpsRegistry
}

func NewSimMsgFactoryWithFutureOps[T sdk.Msg](f common.FactoryMethodWithFutureOps[T]) *LazyStateSimMsgFactory[T] {
	r := &LazyStateSimMsgFactory[T]{}
	r.SimMsgFactoryFn = func(ctx context.Context, testData *common.ChainDataSource, reporter common.SimulationReporter) (signer []common.SimAccount, msg T) {
		signer, msg = f(ctx, testData, reporter, r.fsOpsReg)
		return
	}
	return r
}

func (c *LazyStateSimMsgFactory[T]) SetFutureOpsRegistry(registry common.FutureOpsRegistry) {
	c.fsOpsReg = registry
}

// pass errors through and don't handle them
func expectNoError(err error) error {
	return err
}

var _ common.SimMsgFactoryX = SimMsgFactoryFn[sdk.Msg](nil)

// SimMsgFactoryFn is the default factory for most cases. It does not create future operations but ensures successful message delivery.
type SimMsgFactoryFn[T sdk.Msg] func(ctx context.Context, testData *common.ChainDataSource, reporter common.SimulationReporter) (signer []common.SimAccount, msg T)

// MsgType returns an empty instance of type T, which implements `sdk.Msg`.
func (f SimMsgFactoryFn[T]) MsgType() sdk.Msg {
	var x T
	return x
}

func (f SimMsgFactoryFn[T]) Create() common.FactoryMethod {
	// adapter to return sdk.Msg instead of typed result to match FactoryMethod signature
	return func(ctx context.Context, testData *common.ChainDataSource, reporter common.SimulationReporter) ([]common.SimAccount, sdk.Msg) {
		return f(ctx, testData, reporter)
	}
}

func (f SimMsgFactoryFn[T]) DeliveryResultHandler() common.SimDeliveryResultHandler {
	return expectNoError
}

func (f SimMsgFactoryFn[T]) Cast(msg sdk.Msg) T {
	return msg.(T)
}
