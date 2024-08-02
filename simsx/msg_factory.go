package simsx

import (
	"context"
	"sync"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type SimMsgFactoryX interface {
	MsgType() sdk.Msg
	Create() FactoryMethod
	DeliveryResultHandler() SimDeliveryResultHandler
}
type (
	// FactoryMethod method signature that is implemented by the concrete message factories
	FactoryMethod func(ctx context.Context, testData *ChainDataSource, reporter SimulationReporter) (signer []SimAccount, msg sdk.Msg)

	// FactoryMethodWithFutureOps extended message factory method for the gov module or others that have to schedule operations for a future block.
	FactoryMethodWithFutureOps[T sdk.Msg] func(ctx context.Context, testData *ChainDataSource, reporter SimulationReporter, fOpsReg FutureOpsRegistry) ([]SimAccount, T)

	// FactoryMethodWithDeliveryResultHandler extended factory method that can return a result handler, that is executed on the delivery tx error result.
	// This is used in staking for example to validate negative execution results.
	FactoryMethodWithDeliveryResultHandler[T sdk.Msg] func(ctx context.Context, testData *ChainDataSource, reporter SimulationReporter) (signer []SimAccount, msg T, handler SimDeliveryResultHandler)
)

var _ SimMsgFactoryX = &ResultHandlingSimMsgFactory[sdk.Msg]{}

// ResultHandlingSimMsgFactory message factory with a delivery error result handler configured.
type ResultHandlingSimMsgFactory[T sdk.Msg] struct {
	SimMsgFactoryFn[T]
	resultHandler SimDeliveryResultHandler
}

// NewSimMsgFactoryWithDeliveryResultHandler constructor
func NewSimMsgFactoryWithDeliveryResultHandler[T sdk.Msg](f FactoryMethodWithDeliveryResultHandler[T]) *ResultHandlingSimMsgFactory[T] {
	var (
		mx sync.RWMutex
		// the result handler is always called after the factory. so we initialize it lazy for syntactic sugar and simplicity
		// in the message factory function that is implemented by the users.
		// they are piped here as a block may cause multiple factory calls
		lazyResultHandlers []SimDeliveryResultHandler
	)
	r := &ResultHandlingSimMsgFactory[T]{
		resultHandler: func(err error) error {
			mx.Lock()
			defer mx.Unlock()
			if len(lazyResultHandlers) == 0 {
				return err // default to no error handling
			}
			h := lazyResultHandlers[0]
			lazyResultHandlers = lazyResultHandlers[1:]
			return h(err)
		},
	}
	r.SimMsgFactoryFn = func(ctx context.Context, testData *ChainDataSource, reporter SimulationReporter) (signer []SimAccount, msg T) {
		var resultHandler SimDeliveryResultHandler
		signer, msg, resultHandler = f(ctx, testData, reporter)
		if resultHandler != nil {
			mx.Lock()
			defer mx.Unlock()
			lazyResultHandlers = append(lazyResultHandlers, resultHandler)
		}
		return
	}
	return r
}

func (f ResultHandlingSimMsgFactory[T]) DeliveryResultHandler() SimDeliveryResultHandler {
	return f.resultHandler
}

var (
	_ SimMsgFactoryX       = &LazyStateSimMsgFactory[sdk.Msg]{}
	_ hasFutureOpsRegistry = &LazyStateSimMsgFactory[sdk.Msg]{}
)

// LazyStateSimMsgFactory stateful message factory with weighted proposals and future operation
// registry initialized lazy before execution.
type LazyStateSimMsgFactory[T sdk.Msg] struct {
	SimMsgFactoryFn[T]
	fsOpsReg FutureOpsRegistry
}

func NewSimMsgFactoryWithFutureOps[T sdk.Msg](f FactoryMethodWithFutureOps[T]) *LazyStateSimMsgFactory[T] {
	r := &LazyStateSimMsgFactory[T]{}
	r.SimMsgFactoryFn = func(ctx context.Context, testData *ChainDataSource, reporter SimulationReporter) (signer []SimAccount, msg T) {
		signer, msg = f(ctx, testData, reporter, r.fsOpsReg)
		return
	}
	return r
}

func (c *LazyStateSimMsgFactory[T]) SetFutureOpsRegistry(registry FutureOpsRegistry) {
	c.fsOpsReg = registry
}

// pass errors through and don't handle them
func expectNoError(err error) error {
	return err
}

var _ SimMsgFactoryX = SimMsgFactoryFn[sdk.Msg](nil)

// SimMsgFactoryFn is the default factory for most cases. It does not create future operations but ensures successful message delivery.
type SimMsgFactoryFn[T sdk.Msg] func(ctx context.Context, testData *ChainDataSource, reporter SimulationReporter) (signer []SimAccount, msg T)

// MsgType returns an empty instance of type T, which implements `sdk.Msg`.
func (f SimMsgFactoryFn[T]) MsgType() sdk.Msg {
	var x T
	return x
}

func (f SimMsgFactoryFn[T]) Create() FactoryMethod {
	// adapter to return sdk.Msg instead of typed result to match FactoryMethod signature
	return func(ctx context.Context, testData *ChainDataSource, reporter SimulationReporter) ([]SimAccount, sdk.Msg) {
		return f(ctx, testData, reporter)
	}
}

func (f SimMsgFactoryFn[T]) DeliveryResultHandler() SimDeliveryResultHandler {
	return expectNoError
}

func (f SimMsgFactoryFn[T]) Cast(msg sdk.Msg) T {
	return msg.(T)
}
