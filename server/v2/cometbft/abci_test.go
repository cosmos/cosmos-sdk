package cometbft

import (
	"context"
	"testing"

	appmodulev2 "cosmossdk.io/core/appmodule/v2"
	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/core/transaction"
	am "cosmossdk.io/server/v2/appmanager"
	ammstore "cosmossdk.io/server/v2/appmanager/store"
	"cosmossdk.io/server/v2/cometbft/mempool"
	"cosmossdk.io/server/v2/stf"
	"cosmossdk.io/server/v2/stf/branch"
	"cosmossdk.io/server/v2/stf/mock"
	"cosmossdk.io/store/v2/storage/pebbledb"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestConsensus(t *testing.T) {
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
	ss, _ := ammstore.New(storageDB)

	b := am.Builder[mock.Tx]{
		STF:                s,
		DB:                 ss,
		ValidateTxGasLimit: 100_000,
		QueryGasLimit:      100_000,
		SimulationGasLimit: 100_000,
	}

	am, err := b.Build()
	require.NoError(t, err)

	mockStore := NewMockStore()

	c := NewConsensus[mock.Tx](am, mempool.NoOpMempool[mock.Tx]{}, mockStore, Config{}, mock.TxCodec{}, nil)

	t.Run("Check tx basic", func(t *testing.T) {
		res, err := c.CheckTx(context.Background(), &abci.CheckTxRequest{
			Tx:   mockTx.Bytes(),
			Type: 0,
		})
		require.NotNil(t, res.GasUsed)
		require.NoError(t, err)
	})
}

var actorName = []byte("cookies")

func kvSet(t *testing.T, ctx context.Context, v string) error {
	t.Helper()
	executionCtx := stf.GetExecutionContext(ctx)
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
