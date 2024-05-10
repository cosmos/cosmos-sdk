package appmanager

import (
	"context"
	"testing"

	appmanager "cosmossdk.io/core/app"
	appmodulev2 "cosmossdk.io/core/appmodule/v2"
	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/core/transaction"
	ammstore "cosmossdk.io/server/v2/appmanager/store"
	"cosmossdk.io/server/v2/stf"
	"cosmossdk.io/server/v2/stf/branch"
	"cosmossdk.io/server/v2/stf/mock"
	"cosmossdk.io/store/v2/storage/pebbledb"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestAppManager(t *testing.T) {
	mockTx := mock.Tx{
		Sender:   []byte("sender"),
		Msg:      wrapperspb.Bool(true), // msg does not matter at all because our handler does nothing.
		GasLimit: 100_000,
	}

	s := stf.NewSTF(
		func(ctx context.Context, msg transaction.Msg) (msgResp transaction.Msg, err error) {
			err = kvSet(t, ctx, "exec")
			return nil, err
		},
		func(ctx context.Context, msg transaction.Msg) (msgResp transaction.Msg, err error) {
			err = kvSet(t, ctx, "query")
			return nil, err
		},
		func(ctx context.Context, txs []mock.Tx) error { return nil },
		func(ctx context.Context) error {
			return kvSet(t, ctx, "begin-block")
		},
		func(ctx context.Context) error {
			return kvSet(t, ctx, "end-block")
		},
		func(ctx context.Context, tx mock.Tx) error {
			return kvSet(t, ctx, "validate")
		},
		func(ctx context.Context) ([]appmodulev2.ValidatorUpdate, error) { return nil, nil },
		func(ctx context.Context, tx mock.Tx, success bool) error {
			return kvSet(t, ctx, "post-tx-exec")
		},
		branch.DefaultNewWriterMap,
	)

	storageDB, err := pebbledb.New(t.TempDir())
	require.NoError(t, err)

	store, err := ammstore.New(storageDB)
	require.NoError(t, err)

	b := Builder[mock.Tx]{
		STF:                s,
		DB:                 store,
		ValidateTxGasLimit: 100_000,
		QueryGasLimit:      100_000,
		SimulationGasLimit: 100_000,
	}

	am, err := b.Build()
	require.NoError(t, err)

	t.Run("Invalid block height", func(t *testing.T) {
		_, _, err := am.DeliverBlock(context.Background(), &appmanager.BlockRequest[mock.Tx]{Height: 0})
		require.Error(t, err)
	})

	t.Run("begin and end block", func(t *testing.T) {
		_, newState, err := am.DeliverBlock(context.Background(), &appmanager.BlockRequest[mock.Tx]{Height: 1})
		require.NoError(t, err)
		stateHas(t, newState, "begin-block")
		stateHas(t, newState, "end-block")
	})

	t.Run("basic tx", func(t *testing.T) {
		result, newState, err := am.DeliverBlock(context.Background(), &appmanager.BlockRequest[mock.Tx]{
			Height: 1,
			Txs:    []mock.Tx{mockTx},
		})
		require.NoError(t, err)
		stateHas(t, newState, "validate")
		stateHas(t, newState, "exec")
		stateHas(t, newState, "post-tx-exec")

		require.Len(t, result.TxResults, 1)
		txResult := result.TxResults[0]
		require.NotZero(t, txResult.GasUsed)
		require.Equal(t, mockTx.GasLimit, txResult.GasWanted)
	})

	t.Run("validate tx", func(t *testing.T) {
		_, err := am.ValidateTx(context.Background(), mockTx)
		require.NoError(t, err)
	})

	t.Run("validate tx out of gas", func(t *testing.T) {
		am.config.ValidateTxGasLimit = 0
		result, _ := am.ValidateTx(context.Background(), mockTx)
		require.Error(t, result.Error, "out of gas")
	})

	t.Run("simulate tx", func(t *testing.T) {
		_, newState, err := am.Simulate(context.Background(), mockTx)
		require.NoError(t, err)
		stateHas(t, newState, "validate")
		stateHas(t, newState, "exec")
	})

	// SimulationGasLimit is unusing in STF
	t.Run("simulate tx out of gas", func(t *testing.T) {
		am.config.SimulationGasLimit = 0
		_, newState, err := am.Simulate(context.Background(), mockTx)
		require.NoError(t, err)
		stateHas(t, newState, "validate") // should not has
		stateHas(t, newState, "exec")     // should not has
	})

	t.Run("query basic", func(t *testing.T) {
		_, err = am.Query(context.Background(), 1, nil)
		require.NoError(t, err)
	})

	t.Run("query basic, out of gas", func(t *testing.T) {
		am.config.QueryGasLimit = 0
		_, err := am.Query(context.Background(), 1, nil)
		require.Error(t, err)
	})

}

var actorName = []byte("cookies")

func kvSet(t *testing.T, ctx context.Context, v string) error {
	t.Helper()
	executionCtx := stf.GetExecutionContext(ctx) //TODO: what to do here?
	require.NotNil(t, executionCtx)
	state, err := executionCtx.State.GetWriter(actorName)
	require.NoError(t, err)
	return state.Set([]byte(v), []byte(v))
}

func stateHas(t *testing.T, accountState corestore.ReaderMap, key string) {
	t.Helper()
	state, err := accountState.GetReader(actorName)
	require.NoError(t, err)
	has, err := state.Has([]byte(key))
	require.NoError(t, err)
	require.Truef(t, has, "state did not have key: %s", key)
}

func stateNotHas(t *testing.T, accountState corestore.ReaderMap, key string) {
	t.Helper()
	state, err := accountState.GetReader(actorName)
	require.NoError(t, err)
	has, err := state.Has([]byte(key))
	require.NoError(t, err)
	require.Falsef(t, has, "state was not supposed to have key: %s", key)
}
