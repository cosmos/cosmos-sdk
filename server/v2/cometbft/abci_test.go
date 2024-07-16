package cometbft

import (
	"context"
	"fmt"
	"testing"
	"time"

	appmodulev2 "cosmossdk.io/core/appmodule/v2"
	"cosmossdk.io/core/log"
	corestore "cosmossdk.io/core/store"
	am "cosmossdk.io/server/v2/appmanager"
	"cosmossdk.io/server/v2/cometbft/mempool"
	cometmock "cosmossdk.io/server/v2/cometbft/mock"
	"cosmossdk.io/server/v2/stf"
	"cosmossdk.io/server/v2/stf/branch"
	"cosmossdk.io/server/v2/stf/mock"
	abciproto "github.com/cometbft/cometbft/api/cometbft/abci/v1"

	// gogoproto "github.com/cosmos/gogoproto/proto"
	// gogotypes "github.com/cosmos/gogoproto/types"
	"crypto/sha256"

	"github.com/stretchr/testify/require"
)

func TestConsensus(t *testing.T) {
	// mockTx := mock.Tx{
	// 	Sender:   []byte("sender"),
	// 	Msg:      &gogotypes.BoolValue{Value: true},
	// 	GasLimit: 100_000,
	// }

	sum := sha256.Sum256([]byte("test-hash"))

	s, err := stf.NewSTF(
		log.NewNopLogger().With("module", "stf"),
		stf.NewMsgRouterBuilder(),
		stf.NewMsgRouterBuilder(),
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

	ss := cometmock.NewMockStorage(log.NewNopLogger())
	sc := cometmock.NewMockCommiter(log.NewNopLogger())
	mockStore := cometmock.NewMockStore(ss, sc)

	b := am.Builder[mock.Tx]{
		STF:                s,
		DB:                 mockStore,
		ValidateTxGasLimit: 100_000,
		QueryGasLimit:      100_000,
		SimulationGasLimit: 100_000,
	}

	am, err := b.Build()
	require.NoError(t, err)

	c := NewConsensus[mock.Tx](am, mempool.NoOpMempool[mock.Tx]{}, mockStore, Config{}, mock.TxCodec{}, nil)

	// t.Run("Check tx basic", func(t *testing.T) {
	// 	res, err := c.CheckTx(context.Background(), &abciproto.CheckTxRequest{
	// 		Tx:   mockTx.Bytes(),
	// 		Type: 0,
	// 	})
	// 	require.NotNil(t, res.GasUsed)
	// 	require.NoError(t, err)
	// })

	t.Run("Finalize block", func(t *testing.T) {
		res, err := c.FinalizeBlock(context.Background(), &abciproto.FinalizeBlockRequest{
			Txs:    nil,
			Height: 1,
			Time:   time.Now(),
			Hash:   sum[:],
		})
		fmt.Println(len(sum[:]))
		fmt.Println("Finalize ", res, err)
		require.NoError(t, err)
	})
}

var actorName = []byte("cookies")

func kvSet(t *testing.T, ctx context.Context, v string) error {
	t.Helper()
	executionCtx := stf.GetExecutionContext(ctx)
	require.NotNil(t, executionCtx)
	state, err := stf.GetStateFromContext(executionCtx).GetWriter(actorName)
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
