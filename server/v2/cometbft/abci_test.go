package cometbft

import (
	"context"
	"io"
	"testing"
	"time"

	appmodulev2 "cosmossdk.io/core/appmodule/v2"
	"cosmossdk.io/core/log"
	corestore "cosmossdk.io/core/store"
	am "cosmossdk.io/server/v2/appmanager"
	"cosmossdk.io/server/v2/cometbft/mempool"
	cometmock "cosmossdk.io/server/v2/cometbft/mock"
	"cosmossdk.io/server/v2/cometbft/types"
	"cosmossdk.io/server/v2/stf"
	"cosmossdk.io/server/v2/stf/branch"
	"cosmossdk.io/server/v2/stf/mock"
	abciproto "github.com/cometbft/cometbft/api/cometbft/abci/v1"
	v1 "github.com/cometbft/cometbft/api/cometbft/types/v1"

	"github.com/cosmos/gogoproto/proto"

	"encoding/json"

	consensustypes "cosmossdk.io/x/consensus/types"
	"github.com/stretchr/testify/require"
)

func addMsgHandlerToSTF[T any, PT interface {
	*T
	proto.Message
},
	U any, UT interface {
		*U
		proto.Message
	}](
	t *testing.T,
	s *stf.STF[mock.Tx],
	handler func(ctx context.Context, msg PT) (UT, error),
) {
	t.Helper()
	msgRouterBuilder := stf.NewMsgRouterBuilder()
	err := msgRouterBuilder.RegisterHandler(
		proto.MessageName(PT(new(T))),
		func(ctx context.Context, msg appmodulev2.Message) (msgResp appmodulev2.Message, err error) {
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
	stf.SetMsgRouter(s, msgRouter)
}

func TestConsensus(t *testing.T) {
	// mockTx := mock.Tx{
	// 	Sender:   []byte("sender"),
	// 	Msg:      &gogotypes.BoolValue{Value: true},
	// 	GasLimit: 100_000,
	// }

	// sum := sha256.Sum256([]byte("test-hash"))

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
	sc := cometmock.NewMockCommiter(log.NewNopLogger(), string(actorName), "stf")
	mockStore := cometmock.NewMockStore(ss, sc)

	b := am.Builder[mock.Tx]{
		STF:                s,
		DB:                 mockStore,
		ValidateTxGasLimit: 100_000,
		QueryGasLimit:      100_000,
		SimulationGasLimit: 100_000,
		InitGenesis: func(ctx context.Context, src io.Reader, txHandler func(json.RawMessage) error) error {
			return kvSet(t, ctx, "init-chain")
		},
	}

	am, err := b.Build()
	require.NoError(t, err)

	c := NewConsensus[mock.Tx](am, mempool.NoOpMempool[mock.Tx]{}, mockStore, Config{}, mock.TxCodec{}, log.NewNopLogger())

	t.Run("InitChain without update consensus params", func(t *testing.T) {
		_, err := c.InitChain(context.Background(), &abciproto.InitChainRequest{
			Time:          time.Now(),
			ChainId:       "test",
			InitialHeight: 1,
		})
		require.NoError(t, err)
		stateStorageHas(t, mockStore, "init-chain", 1)
		stateStorageHas(t, mockStore, "end-block", 1)
	})

	t.Run("InitChain with update consensus params", func(t *testing.T) {
		addMsgHandlerToSTF(t, s, func(ctx context.Context, msg *consensustypes.MsgUpdateParams) (*consensustypes.MsgUpdateParams, error) {
			kvSet(t, ctx, "exec")
			return nil, nil
		})

		_, err := c.InitChain(context.Background(), &abciproto.InitChainRequest{
			Time:    time.Now(),
			ChainId: "test",
			ConsensusParams: &v1.ConsensusParams{
				Block: &v1.BlockParams{
					MaxGas: 5000000,
				},
			},
			InitialHeight: 2,
		})
		require.NoError(t, err)
		stateStorageHas(t, mockStore, "init-chain", 2)
		stateStorageHas(t, mockStore, "exec", 2)
		stateStorageHas(t, mockStore, "end-block", 2)

		stateCommitmentNoHas(t, mockStore, "init-chain", 2)
		stateStorageHas(t, mockStore, "exec", 2)
		stateCommitmentNoHas(t, mockStore, "end-block", 2)
	})

	t.Run("FinalizeBlock genesis block", func(t *testing.T) {
		_, err := c.FinalizeBlock(context.Background(), &abciproto.FinalizeBlockRequest{
			Time:   time.Now(),
			Height: 2,
		})
		require.NoError(t, err)
		stateStorageNoHas(t, mockStore, "begin-block", 2)
		stateStorageHas(t, mockStore, "end-block", 2)

		// commit genesis state
		stateCommitmentHas(t, mockStore, "init-chain", 2)
		stateCommitmentHas(t, mockStore, "exec", 2)
		stateCommitmentHas(t, mockStore, "end-block", 2)

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

func stateStorageHas(t *testing.T, store types.Store, key string, version uint64) {
	t.Helper()
	has, err := store.GetStateStorage().Has(actorName, version, []byte(key))
	require.NoError(t, err)
	require.Truef(t, has, "state storage did not have key: %s", key)
}

func stateStorageNoHas(t *testing.T, store types.Store, key string, version uint64) {
	t.Helper()
	has, err := store.GetStateStorage().Has(actorName, version, []byte(key))
	require.NoError(t, err)
	require.Falsef(t, has, "state storage had key: %s", key)
}

func stateCommitmentHas(t *testing.T, store types.Store, key string, version uint64) {
	t.Helper()
	bz, err := store.GetStateCommitment().Get(actorName, version, []byte(key))
	require.NoError(t, err)
	require.NotEqual(t, len(bz), 0)
	require.Equal(t, bz, []byte(key))
}

func stateCommitmentNoHas(t *testing.T, store types.Store, key string, version uint64) {
	t.Helper()
	bz, err := store.GetStateCommitment().Get(actorName, version, []byte(key))
	// if not committed, should return version does not exist
	require.Error(t, err)
	require.Contains(t, err.Error(), "version does not exist")
	require.Equal(t, len(bz), 0)
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
