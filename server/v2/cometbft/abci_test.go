package cometbft

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"io"
	"strings"
	"testing"
	"time"

	abciproto "github.com/cometbft/cometbft/api/cometbft/abci/v1"
	v1 "github.com/cometbft/cometbft/api/cometbft/types/v1"
	"github.com/cosmos/gogoproto/proto"
	gogotypes "github.com/cosmos/gogoproto/types"
	"github.com/stretchr/testify/require"

	appmodulev2 "cosmossdk.io/core/appmodule/v2"
	"cosmossdk.io/core/store"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/log"
	am "cosmossdk.io/server/v2/appmanager"
	"cosmossdk.io/server/v2/cometbft/handlers"
	cometmock "cosmossdk.io/server/v2/cometbft/internal/mock"
	"cosmossdk.io/server/v2/cometbft/mempool"
	"cosmossdk.io/server/v2/cometbft/types"
	"cosmossdk.io/server/v2/stf"
	"cosmossdk.io/server/v2/stf/branch"
	"cosmossdk.io/server/v2/stf/mock"
	consensustypes "cosmossdk.io/x/consensus/types"
)

var (
	sum                   = sha256.Sum256([]byte("test-hash"))
	DefaulConsensusParams = &v1.ConsensusParams{
		Block: &v1.BlockParams{
			MaxGas: 5000000,
		},
	}
	mockTx = mock.Tx{
		Sender:   []byte("sender"),
		Msg:      &gogotypes.BoolValue{Value: true},
		GasLimit: 100_000,
	}
	invalidMockTx = mock.Tx{
		Sender:   []byte("sender"),
		Msg:      &gogotypes.BoolValue{Value: true},
		GasLimit: 0,
	}
	actorName = []byte("cookies")
)

func getQueryRouterBuilder[T any, PT interface {
	*T
	proto.Message
},
	U any, UT interface {
		*U
		proto.Message
	}](
	t *testing.T,
	handler func(ctx context.Context, msg PT) (UT, error),
) *stf.MsgRouterBuilder {
	t.Helper()
	queryRouterBuilder := stf.NewMsgRouterBuilder()
	err := queryRouterBuilder.RegisterHandler(
		proto.MessageName(PT(new(T))),
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

	return queryRouterBuilder
}

func getMsgRouterBuilder[T any, PT interface {
	*T
	transaction.Msg
},
	U any, UT interface {
		*U
		transaction.Msg
	}](
	t *testing.T,
	handler func(ctx context.Context, msg PT) (UT, error),
) *stf.MsgRouterBuilder {
	t.Helper()
	msgRouterBuilder := stf.NewMsgRouterBuilder()
	err := msgRouterBuilder.RegisterHandler(
		proto.MessageName(PT(new(T))),
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

	return msgRouterBuilder
}

func TestConsensus_InitChain_Without_UpdateParam(t *testing.T) {
	c := setUpConsensus(t, 100_000, mempool.NoOpMempool[mock.Tx]{})
	mockStore := c.store
	_, err := c.InitChain(context.Background(), &abciproto.InitChainRequest{
		Time:          time.Now(),
		ChainId:       "test",
		InitialHeight: 1,
	})
	require.NoError(t, err)
	assertStoreLatestVersion(t, mockStore, 0)

	_, err = c.FinalizeBlock(context.Background(), &abciproto.FinalizeBlockRequest{
		Time:   time.Now(),
		Height: 1,
	})
	require.NoError(t, err)
	assertStoreLatestVersion(t, mockStore, 1)
}

func TestConsensus_InitChain_With_UpdateParam(t *testing.T) {
	c := setUpConsensus(t, 100_000, mempool.NoOpMempool[mock.Tx]{})
	mockStore := c.store
	_, err := c.InitChain(context.Background(), &abciproto.InitChainRequest{
		Time:            time.Now(),
		ChainId:         "test",
		ConsensusParams: DefaulConsensusParams,
		InitialHeight:   1,
	})
	require.NoError(t, err)
	assertStoreLatestVersion(t, mockStore, 0)

	_, err = c.FinalizeBlock(context.Background(), &abciproto.FinalizeBlockRequest{
		Time:   time.Now(),
		Height: 1,
	})
	require.NoError(t, err)

	assertStoreLatestVersion(t, mockStore, 1)
}

func TestConsensus_InitChain_Invalid_Height(t *testing.T) {
	c := setUpConsensus(t, 100_000, mempool.NoOpMempool[mock.Tx]{})
	mockStore := c.store
	_, err := c.InitChain(context.Background(), &abciproto.InitChainRequest{
		Time:          time.Now(),
		ChainId:       "test",
		InitialHeight: 2,
	})
	require.NoError(t, err)
	assertStoreLatestVersion(t, mockStore, 0)

	// Shouldn't be able to commit genesis block 2
	_, err = c.FinalizeBlock(context.Background(), &abciproto.FinalizeBlockRequest{
		Time:   time.Now(),
		Height: 2,
	})
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "unable to commit the changeset"))
}

func TestConsensus_FinalizeBlock_Invalid_Height(t *testing.T) {
	c := setUpConsensus(t, 100_000, mempool.NoOpMempool[mock.Tx]{})
	_, err := c.InitChain(context.Background(), &abciproto.InitChainRequest{
		Time:          time.Now(),
		ChainId:       "test",
		InitialHeight: 1,
	})
	require.NoError(t, err)

	_, err = c.FinalizeBlock(context.Background(), &abciproto.FinalizeBlockRequest{
		Time:   time.Now(),
		Height: 1,
	})
	require.NoError(t, err)

	_, err = c.FinalizeBlock(context.Background(), &abciproto.FinalizeBlockRequest{
		Time:   time.Now(),
		Height: 3,
	})
	require.Error(t, err)
}

func TestConsensus_FinalizeBlock_NoTxs(t *testing.T) {
	c := setUpConsensus(t, 100_000, mempool.NoOpMempool[mock.Tx]{})
	mockStore := c.store

	_, err := c.InitChain(context.Background(), &abciproto.InitChainRequest{
		Time:          time.Now(),
		ChainId:       "test",
		InitialHeight: 1,
	})
	require.NoError(t, err)

	_, err = c.FinalizeBlock(context.Background(), &abciproto.FinalizeBlockRequest{
		Time:   time.Now(),
		Height: 1,
	})
	require.NoError(t, err)

	endBlock := 10
	for i := 2; i <= endBlock; i++ {
		_, err = c.FinalizeBlock(context.Background(), &abciproto.FinalizeBlockRequest{
			Time:   time.Now(),
			Height: int64(i),
			Hash:   sum[:],
		})
		require.NoError(t, err)

		assertStoreLatestVersion(t, mockStore, uint64(i))
	}
	require.Equal(t, int64(endBlock), c.lastCommittedHeight.Load())
}

func TestConsensus_FinalizeBlock_MultiTxs_OutOfGas(t *testing.T) {
	c := setUpConsensus(t, 100_000, mempool.NoOpMempool[mock.Tx]{})

	_, err := c.InitChain(context.Background(), &abciproto.InitChainRequest{
		Time:          time.Now(),
		ChainId:       "test",
		InitialHeight: 1,
	})
	require.NoError(t, err)

	_, err = c.FinalizeBlock(context.Background(), &abciproto.FinalizeBlockRequest{
		Time:   time.Now(),
		Height: 1,
	})
	require.NoError(t, err)

	endBlock := 10
	for i := 2; i <= endBlock; i++ {
		res, err := c.FinalizeBlock(context.Background(), &abciproto.FinalizeBlockRequest{
			Time:   time.Now(),
			Height: int64(i),
			Hash:   sum[:],
			Txs:    [][]byte{invalidMockTx.Bytes(), mockTx.Bytes()},
		})
		require.NoError(t, err)
		require.NotEqual(t, res.TxResults[0].Code, 0)
	}
	require.Equal(t, int64(endBlock), c.lastCommittedHeight.Load())
}

func TestConsensus_FinalizeBlock_MultiTxs(t *testing.T) {
	c := setUpConsensus(t, 100_000, mempool.NoOpMempool[mock.Tx]{})
	mockStore := c.store

	_, err := c.InitChain(context.Background(), &abciproto.InitChainRequest{
		Time:          time.Now(),
		ChainId:       "test",
		InitialHeight: 1,
	})
	require.NoError(t, err)

	_, err = c.FinalizeBlock(context.Background(), &abciproto.FinalizeBlockRequest{
		Time:   time.Now(),
		Height: 1,
	})
	require.NoError(t, err)

	endBlock := 10
	for i := 2; i <= endBlock; i++ {
		_, err = c.FinalizeBlock(context.Background(), &abciproto.FinalizeBlockRequest{
			Time:   time.Now(),
			Height: int64(i),
			Hash:   sum[:],
			Txs:    [][]byte{mockTx.Bytes(), mockTx.Bytes()},
		})
		require.NoError(t, err)
		assertStoreLatestVersion(t, mockStore, uint64(i))
	}
	require.Equal(t, int64(endBlock), c.lastCommittedHeight.Load())
}

func TestConsensus_CheckTx(t *testing.T) {
	c := setUpConsensus(t, 0, mempool.NoOpMempool[mock.Tx]{})

	_, err := c.InitChain(context.Background(), &abciproto.InitChainRequest{
		Time:          time.Now(),
		ChainId:       "test",
		InitialHeight: 1,
	})
	require.NoError(t, err)

	// empty byte
	_, err = c.CheckTx(context.Background(), &abciproto.CheckTxRequest{
		Tx: []byte{},
	})
	require.Error(t, err)

	// out of gas
	res, err := c.CheckTx(context.Background(), &abciproto.CheckTxRequest{
		Tx: mock.Tx{
			Sender:   []byte("sender"),
			Msg:      &gogotypes.BoolValue{Value: true},
			GasLimit: 100_000,
		}.Bytes(),
	})
	require.NoError(t, err)
	require.NotEqual(t, res.Code, 0)

	c = setUpConsensus(t, 100_000, mempool.NoOpMempool[mock.Tx]{})
	res, err = c.CheckTx(context.Background(), &abciproto.CheckTxRequest{
		Tx: mock.Tx{
			Sender:   []byte("sender"),
			Msg:      &gogotypes.BoolValue{Value: true},
			GasLimit: 100_000,
		}.Bytes(),
	})
	require.NoError(t, err)
	require.NotEqual(t, res.GasUsed, 0)
}

func TestConsensus_ExtendVote(t *testing.T) {
	c := setUpConsensus(t, 100_000, mempool.NoOpMempool[mock.Tx]{})

	_, err := c.InitChain(context.Background(), &abciproto.InitChainRequest{
		Time:          time.Now(),
		ChainId:       "test",
		InitialHeight: 1,
		ConsensusParams: &v1.ConsensusParams{
			Block: &v1.BlockParams{
				MaxGas: 5000000,
			},
			Feature: &v1.FeatureParams{
				VoteExtensionsEnableHeight: &gogotypes.Int64Value{Value: 2},
			},
		},
	})
	require.NoError(t, err)

	// Votes not enabled yet
	_, err = c.ExtendVote(context.Background(), &abciproto.ExtendVoteRequest{
		Height: 1,
	})
	require.ErrorContains(t, err, "vote extensions are not enabled")

	// Empty extendVote handler
	_, err = c.ExtendVote(context.Background(), &abciproto.ExtendVoteRequest{
		Height: 2,
	})
	require.ErrorContains(t, err, "no extend function was set")

	// Use NoOp handler
	c.extendVote = DefaultServerOptions[mock.Tx]().ExtendVoteHandler
	res, err := c.ExtendVote(context.Background(), &abciproto.ExtendVoteRequest{
		Height: 2,
	})
	require.NoError(t, err)
	require.Equal(t, len(res.VoteExtension), 0)
}

func TestConsensus_VerifyVoteExtension(t *testing.T) {
	c := setUpConsensus(t, 100_000, mempool.NoOpMempool[mock.Tx]{})

	_, err := c.InitChain(context.Background(), &abciproto.InitChainRequest{
		Time:          time.Now(),
		ChainId:       "test",
		InitialHeight: 1,
		ConsensusParams: &v1.ConsensusParams{
			Block: &v1.BlockParams{
				MaxGas: 5000000,
			},
			Feature: &v1.FeatureParams{
				VoteExtensionsEnableHeight: &gogotypes.Int64Value{Value: 2},
			},
		},
	})
	require.NoError(t, err)

	// Votes not enabled yet
	_, err = c.VerifyVoteExtension(context.Background(), &abciproto.VerifyVoteExtensionRequest{
		Height: 1,
	})
	require.ErrorContains(t, err, "vote extensions are not enabled")

	// Empty verifyVote handler
	_, err = c.VerifyVoteExtension(context.Background(), &abciproto.VerifyVoteExtensionRequest{
		Height: 2,
	})
	require.ErrorContains(t, err, "no verify function was set")

	// Use NoOp handler
	c.verifyVoteExt = DefaultServerOptions[mock.Tx]().VerifyVoteExtensionHandler
	res, err := c.VerifyVoteExtension(context.Background(), &abciproto.VerifyVoteExtensionRequest{
		Height: 2,
		Hash:   []byte("test"),
	})
	require.NoError(t, err)
	require.Equal(t, res.Status, abciproto.VERIFY_VOTE_EXTENSION_STATUS_ACCEPT)
}

func TestConsensus_PrepareProposal(t *testing.T) {
	c := setUpConsensus(t, 100_000, mempool.NoOpMempool[mock.Tx]{})

	// Invalid height
	_, err := c.PrepareProposal(context.Background(), &abciproto.PrepareProposalRequest{
		Height: 0,
	})
	require.Error(t, err)

	// empty handler
	_, err = c.PrepareProposal(context.Background(), &abciproto.PrepareProposalRequest{
		Height: 1,
	})
	require.Error(t, err)

	// NoOp handler
	c.prepareProposalHandler = DefaultServerOptions[mock.Tx]().PrepareProposalHandler
	_, err = c.PrepareProposal(context.Background(), &abciproto.PrepareProposalRequest{
		Height: 1,
		Txs:    [][]byte{mockTx.Bytes()},
	})
	require.NoError(t, err)
}

func TestConsensus_PrepareProposal_With_Handler_NoOpMempool(t *testing.T) {
	c := setUpConsensus(t, 100_000, mempool.NoOpMempool[mock.Tx]{})

	c.prepareProposalHandler = handlers.NewDefaultProposalHandler(c.mempool).PrepareHandler()

	// zero MaxTxBytes
	res, err := c.PrepareProposal(context.Background(), &abciproto.PrepareProposalRequest{
		Height:     1,
		MaxTxBytes: 0,
		Txs:        [][]byte{mockTx.Bytes()},
	})
	require.NoError(t, err)
	require.Equal(t, len(res.Txs), 0)

	// have tx exceed MaxTxBytes
	// each mock tx has 128 bytes, should select 2 txs
	res, err = c.PrepareProposal(context.Background(), &abciproto.PrepareProposalRequest{
		Height:     1,
		MaxTxBytes: 300,
		Txs:        [][]byte{mockTx.Bytes(), mockTx.Bytes(), mockTx.Bytes()},
	})
	require.NoError(t, err)
	require.Equal(t, len(res.Txs), 2)

	// reach MaxTxBytes
	res, err = c.PrepareProposal(context.Background(), &abciproto.PrepareProposalRequest{
		Height:     1,
		MaxTxBytes: 256,
		Txs:        [][]byte{mockTx.Bytes(), mockTx.Bytes()},
	})
	require.NoError(t, err)
	require.Equal(t, len(res.Txs), 2)

	// Over gas, under MaxTxBytes
	// 300_000 gas limit, should only take 3 txs
	res, err = c.PrepareProposal(context.Background(), &abciproto.PrepareProposalRequest{
		Height:     1,
		MaxTxBytes: 1000,
		Txs:        [][]byte{mockTx.Bytes(), mockTx.Bytes(), mockTx.Bytes(), mockTx.Bytes()},
	})
	require.NoError(t, err)
	require.Equal(t, len(res.Txs), 3)

	// Reach max gas
	res, err = c.PrepareProposal(context.Background(), &abciproto.PrepareProposalRequest{
		Height:     1,
		MaxTxBytes: 1000,
		Txs:        [][]byte{mockTx.Bytes(), mockTx.Bytes(), mockTx.Bytes()},
	})
	require.NoError(t, err)
	require.Equal(t, len(res.Txs), 3)

	// have a bad encoding tx
	res, err = c.PrepareProposal(context.Background(), &abciproto.PrepareProposalRequest{
		Height:     1,
		MaxTxBytes: 1000,
		Txs:        [][]byte{mockTx.Bytes(), append(mockTx.Bytes(), []byte("bad")...), mockTx.Bytes()},
	})
	require.NoError(t, err)
	require.Equal(t, len(res.Txs), 2)
}

func TestConsensus_ProcessProposal(t *testing.T) {
	c := setUpConsensus(t, 100_000, mempool.NoOpMempool[mock.Tx]{})

	// Invalid height
	_, err := c.ProcessProposal(context.Background(), &abciproto.ProcessProposalRequest{
		Height: 0,
	})
	require.Error(t, err)

	// empty handler
	_, err = c.ProcessProposal(context.Background(), &abciproto.ProcessProposalRequest{
		Height: 1,
	})
	require.Error(t, err)

	// NoOp handler
	c.processProposalHandler = DefaultServerOptions[mock.Tx]().ProcessProposalHandler
	_, err = c.ProcessProposal(context.Background(), &abciproto.ProcessProposalRequest{
		Height: 1,
		Txs:    [][]byte{mockTx.Bytes()},
	})
	require.NoError(t, err)
}

func TestConsensus_ProcessProposal_With_Handler(t *testing.T) {
	c := setUpConsensus(t, 100_000, cometmock.MockMempool[mock.Tx]{})

	c.processProposalHandler = handlers.NewDefaultProposalHandler(c.mempool).ProcessHandler()

	// exceed max gas
	res, err := c.ProcessProposal(context.Background(), &abciproto.ProcessProposalRequest{
		Height: 1,
		Txs:    [][]byte{mockTx.Bytes(), mockTx.Bytes(), mockTx.Bytes(), mockTx.Bytes()},
	})
	require.NoError(t, err)
	require.Equal(t, res.Status, abciproto.PROCESS_PROPOSAL_STATUS_REJECT)

	// have bad encode tx
	// should reject
	res, err = c.ProcessProposal(context.Background(), &abciproto.ProcessProposalRequest{
		Height: 1,
		Txs:    [][]byte{mockTx.Bytes(), append(mockTx.Bytes(), []byte("bad")...), mockTx.Bytes(), mockTx.Bytes()},
	})
	require.NoError(t, err)
	require.Equal(t, res.Status, abciproto.PROCESS_PROPOSAL_STATUS_REJECT)
}

func TestConsensus_Info(t *testing.T) {
	c := setUpConsensus(t, 100_000, cometmock.MockMempool[mock.Tx]{})

	// Version 0
	res, err := c.Info(context.Background(), &abciproto.InfoRequest{})
	require.NoError(t, err)
	require.Equal(t, res.LastBlockHeight, int64(0))

	// Commit store to version 1
	_, err = c.InitChain(context.Background(), &abciproto.InitChainRequest{
		Time:          time.Now(),
		ChainId:       "test",
		InitialHeight: 1,
	})
	require.NoError(t, err)

	_, err = c.FinalizeBlock(context.Background(), &abciproto.FinalizeBlockRequest{
		Time:   time.Now(),
		Height: 1,
	})
	require.NoError(t, err)

	res, err = c.Info(context.Background(), &abciproto.InfoRequest{})
	require.NoError(t, err)
	require.Equal(t, res.LastBlockHeight, int64(1))
}

// TODO:
//   - GRPC request
//   - app request
//   - p2p request
func TestConsensus_Query(t *testing.T) {
	c := setUpConsensus(t, 100_000, cometmock.MockMempool[mock.Tx]{})

	// Write data to state storage
	err := c.store.GetStateStorage().ApplyChangeset(1, &store.Changeset{
		Changes: []store.StateChanges{
			{
				Actor: actorName,
				StateChanges: []store.KVPair{
					{
						Key:    []byte("key"),
						Value:  []byte("value"),
						Remove: false,
					},
				},
			},
		},
	})
	require.NoError(t, err)

	_, err = c.InitChain(context.Background(), &abciproto.InitChainRequest{
		Time:          time.Now(),
		ChainId:       "test",
		InitialHeight: 1,
	})
	require.NoError(t, err)

	_, err = c.FinalizeBlock(context.Background(), &abciproto.FinalizeBlockRequest{
		Time:   time.Now(),
		Height: 1,
		Txs:    [][]byte{mockTx.Bytes()},
	})
	require.NoError(t, err)

	// empty request
	res, err := c.Query(context.Background(), &abciproto.QueryRequest{})
	require.NoError(t, err)
	require.Equal(t, res.Code, uint32(1))
	require.Contains(t, res.Log, "no query path provided")

	// Query store
	res, err = c.Query(context.Background(), &abciproto.QueryRequest{
		Path:   "store/cookies/",
		Data:   []byte("key"),
		Height: 1,
	})
	require.NoError(t, err)
	require.Equal(t, string(res.Value), "value")

	// Query store with no value
	res, err = c.Query(context.Background(), &abciproto.QueryRequest{
		Path:   "store/cookies/",
		Data:   []byte("exec"),
		Height: 1,
	})
	require.NoError(t, err)
	require.Equal(t, res.Value, []byte(nil))
}

func setUpConsensus(t *testing.T, gasLimit uint64, mempool mempool.Mempool[mock.Tx]) *Consensus[mock.Tx] {
	t.Helper()

	msgRouterBuilder := getMsgRouterBuilder(t, func(ctx context.Context, msg *gogotypes.BoolValue) (*gogotypes.BoolValue, error) {
		return nil, nil
	})

	queryRouterBuilder := getQueryRouterBuilder(t, func(ctx context.Context, q *consensustypes.QueryParamsRequest) (*consensustypes.QueryParamsResponse, error) {
		cParams := &v1.ConsensusParams{
			Block: &v1.BlockParams{
				MaxGas: 300000,
			},
			Feature: &v1.FeatureParams{
				VoteExtensionsEnableHeight: &gogotypes.Int64Value{Value: 2},
			},
		}
		return &consensustypes.QueryParamsResponse{
			Params: cParams,
		}, nil
	})

	s, err := stf.NewSTF(
		log.NewNopLogger().With("module", "stf"),
		msgRouterBuilder,
		queryRouterBuilder,
		func(ctx context.Context, txs []mock.Tx) error { return nil },
		func(ctx context.Context) error {
			return nil
		},
		func(ctx context.Context) error {
			return nil
		},
		func(ctx context.Context, tx mock.Tx) error {
			return nil
		},
		func(ctx context.Context) ([]appmodulev2.ValidatorUpdate, error) { return nil, nil },
		func(ctx context.Context, tx mock.Tx, success bool) error {
			return nil
		},
		branch.DefaultNewWriterMap,
	)
	require.NoError(t, err)

	ss := cometmock.NewMockStorage(log.NewNopLogger(), t.TempDir())
	sc := cometmock.NewMockCommiter(log.NewNopLogger(), string(actorName), "stf")
	mockStore := cometmock.NewMockStore(ss, sc)

	b := am.Builder[mock.Tx]{
		STF:                s,
		DB:                 mockStore,
		ValidateTxGasLimit: gasLimit,
		QueryGasLimit:      gasLimit,
		SimulationGasLimit: gasLimit,
		InitGenesis: func(ctx context.Context, src io.Reader, txHandler func(json.RawMessage) error) error {
			return nil
		},
	}

	am, err := b.Build()
	require.NoError(t, err)

	return NewConsensus[mock.Tx](log.NewNopLogger(), "testing-app", am, mempool, map[string]struct{}{}, nil, mockStore, Config{AppTomlConfig: DefaultAppTomlConfig()}, mock.TxCodec{}, "test")
}

// Check target version same with store's latest version
// And should have commit info of target version
// If block 0, commitInfo returned should be nil
func assertStoreLatestVersion(t *testing.T, store types.Store, target uint64) {
	t.Helper()
	version, err := store.GetLatestVersion()
	require.NoError(t, err)
	require.Equal(t, version, target)
	commitInfo, err := store.GetStateCommitment().GetCommitInfo(version)
	require.NoError(t, err)
	if target != 0 {
		require.Equal(t, commitInfo.Version, target)
	} else {
		require.Nil(t, commitInfo)
	}
}
