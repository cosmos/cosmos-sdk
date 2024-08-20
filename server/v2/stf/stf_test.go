package stf

import (
	"context"
	"crypto/sha256"
	"errors"
	"testing"
	"time"

	gogotypes "github.com/cosmos/gogoproto/types"
	"github.com/stretchr/testify/require"

	appmanager "cosmossdk.io/core/app"
	appmodulev2 "cosmossdk.io/core/appmodule/v2"
	coregas "cosmossdk.io/core/gas"
	"cosmossdk.io/core/store"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/server/v2/stf/branch"
	"cosmossdk.io/server/v2/stf/gas"
	"cosmossdk.io/server/v2/stf/mock"
)

func addMsgHandlerToSTF[T any, PT interface {
	*T
	transaction.Msg
},
	U any, UT interface {
		*U
		transaction.Msg
	}](
	t *testing.T,
	stf *STF[mock.Tx],
	handler func(ctx context.Context, msg PT) (UT, error),
) {
	t.Helper()
	msgRouterBuilder := NewMsgRouterBuilder()
	err := msgRouterBuilder.RegisterHandler(
		msgTypeURL(PT(new(T))),
		func(ctx context.Context, msg transaction.Msg) (msgResp transaction.Msg, err error) {
			typedReq := msg.(PT)
			typedResp, err := handler(ctx, typedReq)
			if err != nil {
				return nil, err
			}

			return typedResp, nil
		},
	)
	require.NoError(t, err)

	msgRouter, err := msgRouterBuilder.Build()
	require.NoError(t, err)
	stf.msgRouter = msgRouter
}

func TestSTF(t *testing.T) {
	state := mock.DB()
	mockTx := mock.Tx{
		Sender:   []byte("sender"),
		Msg:      &gogotypes.BoolValue{Value: true},
		GasLimit: 100_000,
	}

	sum := sha256.Sum256([]byte("test-hash"))

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

	t.Run("begin and end block", func(t *testing.T) {
		_, newState, err := s.DeliverBlock(context.Background(), &appmanager.BlockRequest[mock.Tx]{
			Height:  uint64(1),
			Time:    time.Date(2024, 2, 3, 18, 23, 0, 0, time.UTC),
			AppHash: sum[:],
			Hash:    sum[:],
		}, state)
		require.NoError(t, err)
		stateHas(t, newState, "begin-block")
		stateHas(t, newState, "end-block")
	})

	t.Run("basic tx", func(t *testing.T) {
		result, newState, err := s.DeliverBlock(context.Background(), &appmanager.BlockRequest[mock.Tx]{
			Height:  uint64(1),
			Time:    time.Date(2024, 2, 3, 18, 23, 0, 0, time.UTC),
			AppHash: sum[:],
			Hash:    sum[:],
			Txs:     []mock.Tx{mockTx},
		}, state)
		require.NoError(t, err)
		stateHas(t, newState, "validate")
		stateHas(t, newState, "exec")
		stateHas(t, newState, "post-tx-exec")

		require.Len(t, result.TxResults, 1)
		txResult := result.TxResults[0]
		require.NotZero(t, txResult.GasUsed)
		require.Equal(t, mockTx.GasLimit, txResult.GasWanted)
	})

	t.Run("exec tx out of gas", func(t *testing.T) {
		s := s.clone()

		mockTx := mock.Tx{
			Sender:   []byte("sender"),
			Msg:      &gogotypes.BoolValue{Value: true}, // msg does not matter at all because our handler does nothing.
			GasLimit: 0,                                 // NO GAS!
		}

		// this handler will propagate the storage error back, we expect
		// out of gas immediately at tx validation level.
		s.doTxValidation = func(ctx context.Context, tx mock.Tx) error {
			w, err := ctx.(*executionContext).state.GetWriter(actorName)
			require.NoError(t, err)
			err = w.Set([]byte("gas_failure"), []byte{})
			require.Error(t, err)
			return err
		}

		result, newState, err := s.DeliverBlock(context.Background(), &appmanager.BlockRequest[mock.Tx]{
			Height:  uint64(1),
			Time:    time.Date(2024, 2, 3, 18, 23, 0, 0, time.UTC),
			AppHash: sum[:],
			Hash:    sum[:],
			Txs:     []mock.Tx{mockTx},
		}, state)
		require.NoError(t, err)
		stateNotHas(t, newState, "gas_failure") // assert during out of gas no state changes leaked.
		require.ErrorIs(t, result.TxResults[0].Error, coregas.ErrOutOfGas, result.TxResults[0].Error)
	})

	t.Run("fail exec tx", func(t *testing.T) {
		// update the stf to fail on the handler
		s := s.clone()
		addMsgHandlerToSTF(t, &s, func(ctx context.Context, msg *gogotypes.BoolValue) (*gogotypes.BoolValue, error) {
			return nil, errors.New("failure")
		})

		blockResult, newState, err := s.DeliverBlock(context.Background(), &appmanager.BlockRequest[mock.Tx]{
			Height:  uint64(1),
			Time:    time.Date(2024, 2, 3, 18, 23, 0, 0, time.UTC),
			AppHash: sum[:],
			Hash:    sum[:],
			Txs:     []mock.Tx{mockTx},
		}, state)
		require.NoError(t, err)
		require.ErrorContains(t, blockResult.TxResults[0].Error, "failure")
		stateHas(t, newState, "begin-block")
		stateHas(t, newState, "end-block")
		stateHas(t, newState, "validate")
		stateNotHas(t, newState, "exec")
		stateHas(t, newState, "post-tx-exec")
	})

	t.Run("tx is success but post tx failed", func(t *testing.T) {
		s := s.clone()
		s.postTxExec = func(ctx context.Context, tx mock.Tx, success bool) error {
			return errors.New("post tx failure")
		}
		blockResult, newState, err := s.DeliverBlock(context.Background(), &appmanager.BlockRequest[mock.Tx]{
			Height:  uint64(1),
			Time:    time.Date(2024, 2, 3, 18, 23, 0, 0, time.UTC),
			AppHash: sum[:],
			Hash:    sum[:],
			Txs:     []mock.Tx{mockTx},
		}, state)
		require.NoError(t, err)
		require.ErrorContains(t, blockResult.TxResults[0].Error, "post tx failure")
		stateHas(t, newState, "begin-block")
		stateHas(t, newState, "end-block")
		stateHas(t, newState, "validate")
		stateNotHas(t, newState, "exec")
		stateNotHas(t, newState, "post-tx-exec")
	})

	t.Run("tx failed and post tx failed", func(t *testing.T) {
		s := s.clone()
		addMsgHandlerToSTF(t, &s, func(ctx context.Context, msg *gogotypes.BoolValue) (*gogotypes.BoolValue, error) {
			return nil, errors.New("exec failure")
		})
		s.postTxExec = func(ctx context.Context, tx mock.Tx, success bool) error { return errors.New("post tx failure") }
		blockResult, newState, err := s.DeliverBlock(context.Background(), &appmanager.BlockRequest[mock.Tx]{
			Height:  uint64(1),
			Time:    time.Date(2024, 2, 3, 18, 23, 0, 0, time.UTC),
			AppHash: sum[:],
			Hash:    sum[:],
			Txs:     []mock.Tx{mockTx},
		}, state)
		require.NoError(t, err)
		require.ErrorContains(t, blockResult.TxResults[0].Error, "exec failure\npost tx failure")
		stateHas(t, newState, "begin-block")
		stateHas(t, newState, "end-block")
		stateHas(t, newState, "validate")
		stateNotHas(t, newState, "exec")
		stateNotHas(t, newState, "post-tx-exec")
	})

	t.Run("fail validate tx", func(t *testing.T) {
		// update stf to fail on the validation step
		s := s.clone()
		s.doTxValidation = func(ctx context.Context, tx mock.Tx) error { return errors.New("failure") }
		blockResult, newState, err := s.DeliverBlock(context.Background(), &appmanager.BlockRequest[mock.Tx]{
			Height:  uint64(1),
			Time:    time.Date(2024, 2, 3, 18, 23, 0, 0, time.UTC),
			AppHash: sum[:],
			Hash:    sum[:],
			Txs:     []mock.Tx{mockTx},
		}, state)
		require.NoError(t, err)
		require.ErrorContains(t, blockResult.TxResults[0].Error, "failure")
		stateHas(t, newState, "begin-block")
		stateHas(t, newState, "end-block")
		stateNotHas(t, newState, "validate")
		stateNotHas(t, newState, "exec")
	})

	t.Run("test validate tx with exec mode", func(t *testing.T) {
		// update stf to fail on the validation step
		s := s.clone()
		s.doTxValidation = func(ctx context.Context, tx mock.Tx) error {
			if ctx.(*executionContext).execMode == transaction.ExecModeCheck {
				return errors.New("failure")
			}
			return nil
		}
		// test ValidateTx as it validates with check execMode
		res := s.ValidateTx(context.Background(), state, mockTx.GasLimit, mockTx)
		require.Error(t, res.Error)

		// test validate tx with exec mode as finalize
		_, _, err := s.validateTx(context.Background(), s.branchFn(state), mockTx.GasLimit,
			mockTx, transaction.ExecModeFinalize)
		require.NoError(t, err)
	})
}

var actorName = []byte("cookies")

func kvSet(t *testing.T, ctx context.Context, v string) {
	t.Helper()
	state, err := ctx.(*executionContext).state.GetWriter(actorName)
	require.NoError(t, err)
	require.NoError(t, state.Set([]byte(v), []byte(v)))
}

func stateHas(t *testing.T, accountState store.ReaderMap, key string) {
	t.Helper()
	state, err := accountState.GetReader(actorName)
	require.NoError(t, err)
	has, err := state.Has([]byte(key))
	require.NoError(t, err)
	require.Truef(t, has, "state did not have key: %s", key)
}

func stateNotHas(t *testing.T, accountState store.ReaderMap, key string) {
	t.Helper()
	state, err := accountState.GetReader(actorName)
	require.NoError(t, err)
	has, err := state.Has([]byte(key))
	require.NoError(t, err)
	require.Falsef(t, has, "state was not supposed to have key: %s", key)
}
