package stf

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/wrapperspb"

	appmanager "cosmossdk.io/core/app"
	appmodulev2 "cosmossdk.io/core/appmodule/v2"
	coregas "cosmossdk.io/core/gas"
	"cosmossdk.io/core/store"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/server/v2/stf/branch"
	"cosmossdk.io/server/v2/stf/gas"
	"cosmossdk.io/server/v2/stf/mock"
)

func TestSTF(t *testing.T) {
	state := mock.DB()
	mockTx := mock.Tx{
		Sender:   []byte("sender"),
		Msg:      wrapperspb.Bool(true), // msg does not matter at all because our handler does nothing.
		GasLimit: 100_000,
	}

	s := &STF[mock.Tx]{
		handleMsg: func(ctx context.Context, msg transaction.Type) (msgResp transaction.Type, err error) {
			kvSet(t, ctx, "exec")
			return nil, nil
		},
		handleQuery: nil,
		doPreBlock:  func(ctx context.Context, txs []mock.Tx) error { return nil },
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
		branch:           branch.DefaultNewWriterMap,
		getGasMeter:      gas.DefaultGasMeter,
		wrapWithGasMeter: gas.DefaultWrapWithGasMeter,
	}

	t.Run("begin and end block", func(t *testing.T) {
		_, newState, err := s.DeliverBlock(context.Background(), &appmanager.BlockRequest[mock.Tx]{}, state)
		require.NoError(t, err)
		stateHas(t, newState, "begin-block")
		stateHas(t, newState, "end-block")
	})

	t.Run("basic tx", func(t *testing.T) {
		result, newState, err := s.DeliverBlock(context.Background(), &appmanager.BlockRequest[mock.Tx]{
			Txs: []mock.Tx{mockTx},
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
			Msg:      wrapperspb.Bool(true), // msg does not matter at all because our handler does nothing.
			GasLimit: 0,                     // NO GAS!
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
			Txs: []mock.Tx{mockTx},
		}, state)
		require.NoError(t, err)
		stateNotHas(t, newState, "gas_failure") // assert during out of gas no state changes leaked.
		require.ErrorIs(t, result.TxResults[0].Error, coregas.ErrOutOfGas, result.TxResults[0].Error)
	})

	t.Run("fail exec tx", func(t *testing.T) {
		// update the stf to fail on the handler
		s := s.clone()
		s.handleMsg = func(ctx context.Context, msg transaction.Type) (msgResp transaction.Type, err error) {
			return nil, fmt.Errorf("failure")
		}

		blockResult, newState, err := s.DeliverBlock(context.Background(), &appmanager.BlockRequest[mock.Tx]{
			Txs: []mock.Tx{mockTx},
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
			return fmt.Errorf("post tx failure")
		}
		blockResult, newState, err := s.DeliverBlock(context.Background(), &appmanager.BlockRequest[mock.Tx]{
			Txs: []mock.Tx{mockTx},
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
		s.handleMsg = func(ctx context.Context, msg transaction.Type) (msgResp transaction.Type, err error) {
			return nil, fmt.Errorf("exec failure")
		}
		s.postTxExec = func(ctx context.Context, tx mock.Tx, success bool) error { return fmt.Errorf("post tx failure") }
		blockResult, newState, err := s.DeliverBlock(context.Background(), &appmanager.BlockRequest[mock.Tx]{
			Txs: []mock.Tx{mockTx},
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
		s.doTxValidation = func(ctx context.Context, tx mock.Tx) error { return fmt.Errorf("failure") }
		blockResult, newState, err := s.DeliverBlock(context.Background(), &appmanager.BlockRequest[mock.Tx]{
			Txs: []mock.Tx{mockTx},
		}, state)
		require.NoError(t, err)
		require.ErrorContains(t, blockResult.TxResults[0].Error, "failure")
		stateHas(t, newState, "begin-block")
		stateHas(t, newState, "end-block")
		stateNotHas(t, newState, "validate")
		stateNotHas(t, newState, "exec")
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
