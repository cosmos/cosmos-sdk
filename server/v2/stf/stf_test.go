package stf

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"cosmossdk.io/server/v2/core/appmanager"
	"cosmossdk.io/server/v2/core/store"
	"cosmossdk.io/server/v2/core/transaction"
	"cosmossdk.io/server/v2/stf/branch"
	"cosmossdk.io/server/v2/stf/mock"
)

func TestSTF(t *testing.T) {
	state := mock.DB()
	mockTx := mock.Tx{
		Sender: []byte("sender"),
		Msg:    wrapperspb.Bool(true), // msg does not matter at all because our handler does nothing.
	}

	stf := &STF[mock.Tx]{
		handleMsg: func(ctx context.Context, msg Type) (msgResp Type, err error) {
			kvSet(t, ctx, "exec")
			return nil, nil
		},
		doUpgradeBlock: func(ctx context.Context) (bool, error) { return false, nil },
		doBeginBlock: func(ctx context.Context) error {
			kvSet(t, ctx, "begin-block")
			return nil
		},
		doEndBlock: func(ctx context.Context) error {
			kvSet(t, ctx, "end-block")
			return nil
		},
		doTxValidation: func(ctx context.Context, tx mock.Tx) error {
			kvSet(t, ctx, "validate")
			return nil
		},
		postTxExec: func(ctx context.Context, tx mock.Tx, success bool) error {
			kvSet(t, ctx, "post-tx-exec")
			return nil
		},
		branch:            func(store store.ReadonlyState) store.WritableState { return branch.NewStore(store) },
		doValidatorUpdate: func(ctx context.Context) ([]appmanager.ValidatorUpdate, error) { return nil, nil },
	}

	t.Run("begin and end block", func(t *testing.T) {
		_, newState, err := stf.DeliverBlock(context.Background(), &appmanager.BlockRequest[mock.Tx]{}, state)
		require.NoError(t, err)
		stateHas(t, newState, "begin-block")
		stateHas(t, newState, "end-block")
	})

	t.Run("basic tx", func(t *testing.T) {
		_, newState, err := stf.DeliverBlock(context.Background(), &appmanager.BlockRequest[mock.Tx]{
			Txs: []mock.Tx{mockTx},
		}, state)
		require.NoError(t, err)
		stateHas(t, newState, "validate")
		stateHas(t, newState, "exec")
		stateHas(t, newState, "post-tx-exec")
	})

	t.Run("fail exec tx", func(t *testing.T) {
		// update the stf to fail on the handler
		stf := cloneSTF(stf)
		stf.handleMsg = func(ctx context.Context, msg Type) (msgResp Type, err error) { return nil, fmt.Errorf("failure") }

		blockResult, newState, err := stf.DeliverBlock(context.Background(), &appmanager.BlockRequest[mock.Tx]{
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
		stf := cloneSTF(stf)
		stf.postTxExec = func(ctx context.Context, tx mock.Tx, success bool) error {
			return fmt.Errorf("post tx failure")
		}
		blockResult, newState, err := stf.DeliverBlock(context.Background(), &appmanager.BlockRequest[mock.Tx]{
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
		stf := cloneSTF(stf)
		stf.handleMsg = func(ctx context.Context, msg Type) (msgResp Type, err error) { return nil, fmt.Errorf("exec failure") }
		stf.postTxExec = func(ctx context.Context, tx mock.Tx, success bool) error { return fmt.Errorf("post tx failure") }
		blockResult, newState, err := stf.DeliverBlock(context.Background(), &appmanager.BlockRequest[mock.Tx]{
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
		stf := cloneSTF(stf)
		stf.doTxValidation = func(ctx context.Context, tx mock.Tx) error { return fmt.Errorf("failure") }
		blockResult, newState, err := stf.DeliverBlock(context.Background(), &appmanager.BlockRequest[mock.Tx]{
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

func kvSet(t *testing.T, ctx context.Context, v string) {
	t.Helper()
	require.NoError(t, ctx.(*executionContext).store.Set([]byte(v), []byte(v)))
}

func stateHas(t *testing.T, state store.ReadonlyState, key string) {
	t.Helper()
	has, err := state.Has([]byte(key))
	require.NoError(t, err)
	require.Truef(t, has, "state did not have key: %s", key)
}

func stateNotHas(t *testing.T, state store.ReadonlyState, key string) {
	t.Helper()
	has, err := state.Has([]byte(key))
	require.NoError(t, err)
	require.Falsef(t, has, "state was not supposed to have key: %s", key)
}

func cloneSTF[T transaction.Tx](stf *STF[T]) *STF[T] {
	return &STF[T]{
		handleMsg:         stf.handleMsg,
		handleQuery:       stf.handleQuery,
		doUpgradeBlock:    stf.doUpgradeBlock,
		doBeginBlock:      stf.doBeginBlock,
		doEndBlock:        stf.doEndBlock,
		doValidatorUpdate: stf.doValidatorUpdate,
		doTxValidation:    stf.doTxValidation,
		postTxExec:        stf.postTxExec,
		branch:            stf.branch,
	}
}
