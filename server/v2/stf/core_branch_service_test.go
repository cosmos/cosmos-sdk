package stf

import (
	"context"
	"errors"
	"testing"

	gogotypes "github.com/cosmos/gogoproto/types"

	appmodulev2 "cosmossdk.io/core/appmodule/v2"
	"cosmossdk.io/server/v2/stf/branch"
	"cosmossdk.io/server/v2/stf/gas"
	"cosmossdk.io/server/v2/stf/mock"
)

func TestBranchService(t *testing.T) {
	s := &STF[mock.Tx]{
		doPreBlock: func(ctx context.Context, txs []mock.Tx) error { return nil },
		doBeginBlock: func(ctx context.Context) error {
			kvSet(t, ctx, "begin-block")
			return nil
		},
		doEndBlock: func(ctx context.Context) error {
			kvSet(t, ctx, "end-block")
			return nil
		},
		doValidatorUpdate: func(ctx context.Context) ([]appmodulev2.ValidatorUpdate, error) { return nil, nil },
		doTxValidation: func(ctx context.Context, tx mock.Tx) error {
			kvSet(t, ctx, "validate")
			return nil
		},
		postTxExec: func(ctx context.Context, tx mock.Tx, success bool) error {
			kvSet(t, ctx, "post-tx-exec")
			return nil
		},
		branchFn:            branch.DefaultNewWriterMap,
		makeGasMeter:        gas.DefaultGasMeter,
		makeGasMeteredState: gas.DefaultWrapWithGasMeter,
	}
	addMsgHandlerToSTF(t, s, func(ctx context.Context, msg *gogotypes.BoolValue) (*gogotypes.BoolValue, error) {
		kvSet(t, ctx, "exec")
		return nil, nil
	})

	makeContext := func() *executionContext {
		state := mock.DB()
		writableState := s.branchFn(state)
		ctx := s.makeContext(context.Background(), []byte("cookies"), writableState, 0)
		ctx.setGasLimit(1000000)
		return ctx
	}

	branchService := BranchService{}

	// TODO: add events check + gas limit precision test

	t.Run("ok", func(t *testing.T) {
		stfCtx := makeContext()
		gasUsed, err := branchService.ExecuteWithGasLimit(stfCtx, 10000, func(ctx context.Context) error {
			kvSet(t, ctx, "cookies")
			return nil
		})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if gasUsed == 0 {
			t.Error("expected non-zero gasUsed")
		}
		stateHas(t, stfCtx.state, "cookies")
	})

	t.Run("fail - reverts state", func(t *testing.T) {
		stfCtx := makeContext()
		originalGas := stfCtx.meter.Remaining()
		gasUsed, err := branchService.ExecuteWithGasLimit(stfCtx, 10000, func(ctx context.Context) error {
			kvSet(t, ctx, "cookies")
			return errors.New("fail")
		})
		if err == nil {
			t.Error("expected error")
		}
		if gasUsed == 0 {
			t.Error("expected non-zero gasUsed")
		}
		if stfCtx.meter.Remaining() != originalGas-gasUsed {
			t.Error("expected gas to be reverted")
		}

		stateNotHas(t, stfCtx.state, "cookies")
	})

	t.Run("fail - out of gas", func(t *testing.T) {
		stfCtx := makeContext()

		gasUsed, err := branchService.ExecuteWithGasLimit(stfCtx, 4000, func(ctx context.Context) error {
			state, _ := ctx.(*executionContext).state.GetWriter(actorName)
			_ = state.Set([]byte("not out of gas"), []byte{})
			return state.Set([]byte("out of gas"), []byte{})
		})
		if err == nil {
			t.Error("expected error")
		}
		if gasUsed == 0 {
			t.Error("expected non-zero gasUsed")
		}
		stateNotHas(t, stfCtx.state, "cookies")
		if stfCtx.meter.Limit()-stfCtx.meter.Remaining() != 1000 {
			t.Error("expected gas limit precision to be 1000")
		}
	})
}
