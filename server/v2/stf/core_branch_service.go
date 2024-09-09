package stf

import (
	"context"

	"cosmossdk.io/core/branch"
	"cosmossdk.io/core/store"
)

type branchFn func(state store.ReaderMap) store.WriterMap

var _ branch.Service = (*BranchService)(nil)

type BranchService struct{}

func (bs BranchService) Execute(ctx context.Context, f func(ctx context.Context) error) error {
	exCtx, err := getExecutionCtxFromContext(ctx)
	if err != nil {
		return err
	}

	return bs.execute(exCtx, f)
}

func (bs BranchService) ExecuteWithGasLimit(
	ctx context.Context,
	gasLimit uint64,
	f func(ctx context.Context) error,
) (gasUsed uint64, err error) {
	exCtx, err := getExecutionCtxFromContext(ctx)
	if err != nil {
		return 0, err
	}

	originalGasMeter := exCtx.meter

	exCtx.setGasLimit(gasLimit)

	// execute branched, with predefined gas limit.
	err = bs.execute(exCtx, f)
	// restore original context
	gasUsed = exCtx.meter.Limit() - exCtx.meter.Remaining()
	_ = originalGasMeter.Consume(gasUsed, "execute-with-gas-limit")
	exCtx.setGasLimit(originalGasMeter.Remaining())

	return gasUsed, err
}

func (bs BranchService) execute(ctx *executionContext, f func(ctx context.Context) error) error {
	branchedState := ctx.branchFn(ctx.unmeteredState)
	meteredBranchedState := ctx.makeGasMeteredStore(ctx.meter, branchedState)

	branchedCtx := &executionContext{
		Context:             ctx.Context,
		unmeteredState:      branchedState,
		state:               meteredBranchedState,
		meter:               ctx.meter,
		events:              nil,
		sender:              ctx.sender,
		headerInfo:          ctx.headerInfo,
		execMode:            ctx.execMode,
		branchFn:            ctx.branchFn,
		makeGasMeter:        ctx.makeGasMeter,
		makeGasMeteredStore: ctx.makeGasMeteredStore,
		msgRouter:           ctx.msgRouter,
		queryRouter:         ctx.queryRouter,
	}

	err := f(branchedCtx)
	if err != nil {
		return err
	}

	// apply state changes to original state
	if len(branchedCtx.events) != 0 {
		ctx.events = append(ctx.events, branchedCtx.events...)
	}

	err = applyStateChanges(ctx.state, branchedCtx.unmeteredState)
	if err != nil {
		return err
	}

	return nil
}
