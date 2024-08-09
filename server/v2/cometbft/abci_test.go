package cometbft

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	appmodulev2 "cosmossdk.io/core/appmodule/v2"
	"cosmossdk.io/log"
	am "cosmossdk.io/server/v2/appmanager"
	"cosmossdk.io/server/v2/cometbft/mempool"
	cometmock "cosmossdk.io/server/v2/cometbft/mock"
	"cosmossdk.io/server/v2/cometbft/types"
	"cosmossdk.io/server/v2/cometbft/handlers"
	"cosmossdk.io/server/v2/stf"
	"cosmossdk.io/server/v2/stf/branch"
	"cosmossdk.io/server/v2/stf/mock"
	abciproto "github.com/cometbft/cometbft/api/cometbft/abci/v1"
	v1 "github.com/cometbft/cometbft/api/cometbft/types/v1"

	"github.com/cosmos/gogoproto/proto"

	"encoding/json"

	consensustypes "cosmossdk.io/x/consensus/types"
	gogotypes "github.com/cosmos/gogoproto/types"
	"github.com/stretchr/testify/require"
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

func addQueryHandlerToSTF[T any, PT interface {
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
	queryRouterBuilder := stf.NewMsgRouterBuilder()
	err := queryRouterBuilder.RegisterHandler(
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

	queryRouter, err := queryRouterBuilder.Build()
	require.NoError(t, err)
	stf.SetQueryRouter(s, queryRouter)
}

func TestConsensus_InitChain_Without_UpdateParam(t *testing.T) {
	c, _ := setUpConsensus(t, 100_000, mempool.NoOpMempool[mock.Tx]{})
	mockStore := c.store
	_, err := c.InitChain(context.Background(), &abciproto.InitChainRequest{
		Time:          time.Now(),
		ChainId:       "test",
		InitialHeight: 1,
	})
	require.NoError(t, err)
	stateStorageHas(t, mockStore, "init-chain", 1)
	stateStorageHas(t, mockStore, "end-block", 1)

	stateCommitmentNoHas(t, mockStore, "init-chain", 1)
	stateCommitmentNoHas(t, mockStore, "end-block", 1)

	_, err = c.FinalizeBlock(context.Background(), &abciproto.FinalizeBlockRequest{
		Time:   time.Now(),
		Height: 1,
	})
	require.NoError(t, err)

	stateCommitmentHas(t, mockStore, "init-chain", 1)
	stateCommitmentHas(t, mockStore, "end-block", 1)
}

func TestConsensus_InitChain_With_UpdateParam(t *testing.T) {
	c, s := setUpConsensus(t, 100_000, mempool.NoOpMempool[mock.Tx]{})
	mockStore := c.store
	addMsgHandlerToSTF(t, s, func(ctx context.Context, msg *consensustypes.MsgUpdateParams) (*consensustypes.MsgUpdateParams, error) {
		kvSet(t, ctx, "exec")
		return nil, nil
	})
	_, err := c.InitChain(context.Background(), &abciproto.InitChainRequest{
		Time:            time.Now(),
		ChainId:         "test",
		ConsensusParams: DefaulConsensusParams,
		InitialHeight:   1,
	})
	require.NoError(t, err)
	stateStorageHas(t, mockStore, "init-chain", 1)
	stateStorageHas(t, mockStore, "exec", 1)
	stateStorageHas(t, mockStore, "end-block", 1)

	stateCommitmentNoHas(t, mockStore, "init-chain", 1)
	stateCommitmentNoHas(t, mockStore, "exec", 1)
	stateCommitmentNoHas(t, mockStore, "end-block", 1)

	_, err = c.FinalizeBlock(context.Background(), &abciproto.FinalizeBlockRequest{
		Time:   time.Now(),
		Height: 1,
	})
	require.NoError(t, err)

	stateCommitmentHas(t, mockStore, "init-chain", 1)
	stateCommitmentHas(t, mockStore, "exec", 1)
	stateCommitmentHas(t, mockStore, "end-block", 1)
}

func TestConsensus_InitChain_Invalid_Height(t *testing.T) {
	c, _ := setUpConsensus(t, 100_000, mempool.NoOpMempool[mock.Tx]{})
	mockStore := c.store
	_, err := c.InitChain(context.Background(), &abciproto.InitChainRequest{
		Time:          time.Now(),
		ChainId:       "test",
		InitialHeight: 2,
	})
	require.NoError(t, err)
	stateStorageHas(t, mockStore, "init-chain", 2)
	stateStorageHas(t, mockStore, "end-block", 2)

	// Shouldn't be able to commit genesis block 2
	_, err = c.FinalizeBlock(context.Background(), &abciproto.FinalizeBlockRequest{
		Time:   time.Now(),
		Height: 2,
	})
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "unable to commit the changeset"))
}

func TestConsensus_FinalizeBlock_Invalid_Height(t *testing.T) {
	c, _ := setUpConsensus(t, 100_000, mempool.NoOpMempool[mock.Tx]{})
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
	fmt.Println(err)
}

func TestConsensus_FinalizeBlock_NoTxs(t *testing.T) {
	c, s := setUpConsensus(t, 100_000, mempool.NoOpMempool[mock.Tx]{})
	mockStore := c.store

	addQueryHandlerToSTF(t, s, func(ctx context.Context, q *consensustypes.QueryParamsRequest) (*consensustypes.QueryParamsResponse, error) {
		cParams := DefaulConsensusParams
		kvSet(t, ctx, "query")
		return &consensustypes.QueryParamsResponse{
			Params: cParams,
		}, nil
	})

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

		stateCommitmentHas(t, mockStore, "begin-block", uint64(i))
		stateCommitmentNoHas(t, mockStore, "exec", uint64(i))
		stateCommitmentHas(t, mockStore, "end-block", uint64(i))
	}
	require.Equal(t, int64(endBlock), c.lastCommittedHeight.Load())
}

func TestConsensus_FinalizeBlock_MultiTxs_OutOfGas(t *testing.T) {
	c, s := setUpConsensus(t, 100_000, mempool.NoOpMempool[mock.Tx]{})

	addQueryHandlerToSTF(t, s, func(ctx context.Context, q *consensustypes.QueryParamsRequest) (*consensustypes.QueryParamsResponse, error) {
		cParams := DefaulConsensusParams
		kvSet(t, ctx, "query")
		return &consensustypes.QueryParamsResponse{
			Params: cParams,
		}, nil
	})
	addMsgHandlerToSTF(t, s, func(ctx context.Context, msg *gogotypes.BoolValue) (*gogotypes.BoolValue, error) {
		kvSet(t, ctx, "exec")
		return nil, nil
	})

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
			Txs:    [][]byte{invalidMockTx.Bytes()},
		})
		fmt.Println("res", res, err)
		require.NoError(t, err)
		require.NotEqual(t, res.TxResults[0].Code, 0)
		require.Contains(t, res.TxResults[0].Log, "out of gas")
	}
	require.Equal(t, int64(endBlock), c.lastCommittedHeight.Load())
}

func TestConsensus_FinalizeBlock_MultiTxs(t *testing.T) {
	c, s := setUpConsensus(t, 100_000, mempool.NoOpMempool[mock.Tx]{})
	mockStore := c.store

	addQueryHandlerToSTF(t, s, func(ctx context.Context, q *consensustypes.QueryParamsRequest) (*consensustypes.QueryParamsResponse, error) {
		cParams := DefaulConsensusParams
		kvSet(t, ctx, "query")
		return &consensustypes.QueryParamsResponse{
			Params: cParams,
		}, nil
	})
	addMsgHandlerToSTF(t, s, func(ctx context.Context, msg *gogotypes.BoolValue) (*gogotypes.BoolValue, error) {
		kvSet(t, ctx, "exec")
		return nil, nil
	})

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
			Txs:    [][]byte{mockTx.Bytes()},
		})
		require.NoError(t, err)
		stateCommitmentHas(t, mockStore, "init-chain", uint64(i))
		stateCommitmentHas(t, mockStore, "exec", uint64(i))
		stateCommitmentHas(t, mockStore, "end-block", uint64(i))
	}
	require.Equal(t, int64(endBlock), c.lastCommittedHeight.Load())
}

func TestConsensus_CheckTx(t *testing.T) {
	c, s := setUpConsensus(t, 0, mempool.NoOpMempool[mock.Tx]{})

	addMsgHandlerToSTF(t, s, func(ctx context.Context, msg *gogotypes.BoolValue) (*gogotypes.BoolValue, error) {
		kvSet(t, ctx, "exec")
		return nil, nil
	})

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
	require.Contains(t, res.Log, "out of gas")
	require.NotEqual(t, res.Code, 0)

	c, _ = setUpConsensus(t, 100_000, mempool.NoOpMempool[mock.Tx]{})
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
	c, s := setUpConsensus(t, 100_000, mempool.NoOpMempool[mock.Tx]{})

	addQueryHandlerToSTF(t, s, func(ctx context.Context, q *consensustypes.QueryParamsRequest) (*consensustypes.QueryParamsResponse, error) {
		cParams := &v1.ConsensusParams{
			Block: &v1.BlockParams{
				MaxGas: 5000000,
			},
			Abci: &v1.ABCIParams{
				VoteExtensionsEnableHeight: 2,
			},
			Feature: &v1.FeatureParams{
				VoteExtensionsEnableHeight: &gogotypes.Int64Value{Value: 2},
			},
		}

		kvSet(t, ctx, "query")
		return &consensustypes.QueryParamsResponse{
			Params: cParams,
		}, nil
	})
	addMsgHandlerToSTF(t, s, func(ctx context.Context, msg *consensustypes.MsgUpdateParams) (*consensustypes.MsgUpdateParams, error) {
		kvSet(t, ctx, "exec")
		return nil, nil
	})

	_, err := c.InitChain(context.Background(), &abciproto.InitChainRequest{
		Time:          time.Now(),
		ChainId:       "test",
		InitialHeight: 1,
		ConsensusParams: &v1.ConsensusParams{
			Block: &v1.BlockParams{
				MaxGas: 5000000,
			},
			Abci: &v1.ABCIParams{
				VoteExtensionsEnableHeight: 2,
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
	require.ErrorContains(t, err, "verify function was set")

	// Use NoOp handler
	c.extendVote = DefaultServerOptions[mock.Tx]().ExtendVoteHandler
	res, err := c.ExtendVote(context.Background(), &abciproto.ExtendVoteRequest{
		Height: 2,
	})
	require.NoError(t, err)
	require.Equal(t, len(res.VoteExtension), 0)
}

func TestConsensus_VerifyVoteExtension(t *testing.T) {
	c, s := setUpConsensus(t, 100_000, mempool.NoOpMempool[mock.Tx]{})

	addQueryHandlerToSTF(t, s, func(ctx context.Context, q *consensustypes.QueryParamsRequest) (*consensustypes.QueryParamsResponse, error) {
		cParams := &v1.ConsensusParams{
			Block: &v1.BlockParams{
				MaxGas: 5000000,
			},
			Abci: &v1.ABCIParams{
				VoteExtensionsEnableHeight: 2,
			},
			Feature: &v1.FeatureParams{
				VoteExtensionsEnableHeight: &gogotypes.Int64Value{Value: 2},
			},
		}
		return &consensustypes.QueryParamsResponse{
			Params: cParams,
		}, nil
	})
	addMsgHandlerToSTF(t, s, func(ctx context.Context, msg *consensustypes.MsgUpdateParams) (*consensustypes.MsgUpdateParams, error) {
		kvSet(t, ctx, "exec")
		return nil, nil
	})

	_, err := c.InitChain(context.Background(), &abciproto.InitChainRequest{
		Time:          time.Now(),
		ChainId:       "test",
		InitialHeight: 1,
		ConsensusParams: &v1.ConsensusParams{
			Block: &v1.BlockParams{
				MaxGas: 5000000,
			},
			Abci: &v1.ABCIParams{
				VoteExtensionsEnableHeight: 2,
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
	require.ErrorContains(t, err, "verify function was set")

	// Use NoOp handler
	c.verifyVoteExt = DefaultServerOptions[mock.Tx]().VerifyVoteExtensionHandler
	res, err := c.VerifyVoteExtension(context.Background(), &abciproto.VerifyVoteExtensionRequest{
		Height: 2,
		Hash: []byte("test"),
	})
	fmt.Println("res", res, err)
	require.NoError(t, err)
	require.Equal(t, res.Status, abciproto.VERIFY_VOTE_EXTENSION_STATUS_ACCEPT)
}

func TestConsensus_PrepareProposal(t *testing.T) {
	c, _ := setUpConsensus(t, 100_000, mempool.NoOpMempool[mock.Tx]{})

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
		Txs: [][]byte{mockTx.Bytes()},
	})
	require.NoError(t, err)
}

func TestConsensus_PrepareProposal_With_Handler_NoOpMempool(t *testing.T) {
	c, s := setUpConsensus(t, 100_000, mempool.NoOpMempool[mock.Tx]{})

	addQueryHandlerToSTF(t, s, func(ctx context.Context, q *consensustypes.QueryParamsRequest) (*consensustypes.QueryParamsResponse, error) {
		cParams := &v1.ConsensusParams{
			Block: &v1.BlockParams{
				MaxGas: 300_000,
			},
		}
		return &consensustypes.QueryParamsResponse{
			Params: cParams,
		}, nil
	})

	c.prepareProposalHandler = handlers.NewDefaultProposalHandler(c.mempool).PrepareHandler()
	
	// zero MaxTxBytes
	res, err := c.PrepareProposal(context.Background(), &abciproto.PrepareProposalRequest{
		Height: 1,
		MaxTxBytes: 0,
		Txs: [][]byte{mockTx.Bytes()},
	})
	require.NoError(t, err)
	require.Equal(t, len(res.Txs), 0)

	// have tx exeed MaxTxBytes
	// each mock tx has 128 bytes, should select 2 txs
	res, err = c.PrepareProposal(context.Background(), &abciproto.PrepareProposalRequest{
		Height: 1,
		MaxTxBytes: 300,
		Txs: [][]byte{mockTx.Bytes(), mockTx.Bytes(), mockTx.Bytes()},
	})
	require.NoError(t, err)
	require.Equal(t, len(res.Txs), 2)

	// reach MaxTxBytes
	res, err = c.PrepareProposal(context.Background(), &abciproto.PrepareProposalRequest{
		Height: 1,
		MaxTxBytes: 256,
		Txs: [][]byte{mockTx.Bytes(), mockTx.Bytes()},
	})
	require.NoError(t, err)
	require.Equal(t, len(res.Txs), 2)

	// Over gas, under MaxTxBytes
	// 300_000 gas limit, should only take 3 txs
	res, err = c.PrepareProposal(context.Background(), &abciproto.PrepareProposalRequest{
		Height: 1,
		MaxTxBytes: 1000,
		Txs: [][]byte{mockTx.Bytes(), mockTx.Bytes(), mockTx.Bytes(), mockTx.Bytes()},
	})
	require.NoError(t, err)
	require.Equal(t, len(res.Txs), 3)

	// Reach max gas
	res, err = c.PrepareProposal(context.Background(), &abciproto.PrepareProposalRequest{
		Height: 1,
		MaxTxBytes: 1000,
		Txs: [][]byte{mockTx.Bytes(), mockTx.Bytes(), mockTx.Bytes()},
	})
	require.NoError(t, err)
	require.Equal(t, len(res.Txs), 3)

	// have a bad encoding tx
	res, err = c.PrepareProposal(context.Background(), &abciproto.PrepareProposalRequest{
		Height: 1,
		MaxTxBytes: 1000,
		Txs: [][]byte{mockTx.Bytes(), append(mockTx.Bytes(), []byte("bad")...), mockTx.Bytes()},
	})
	require.NoError(t, err)
	require.Equal(t, len(res.Txs), 2)
}

func TestConsensus_ProcessProposal(t *testing.T) {
	c, _ := setUpConsensus(t, 100_000, mempool.NoOpMempool[mock.Tx]{})

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
		Txs: [][]byte{mockTx.Bytes()},
	})
	require.NoError(t, err)

	// have bad encode tx
	_, err = c.ProcessProposal(context.Background(), &abciproto.ProcessProposalRequest{
		Height: 1,
		Txs: [][]byte{mockTx.Bytes(), append(mockTx.Bytes(), []byte("bad")...), mockTx.Bytes()},
	})
	fmt.Println("err bad encode", err)
	require.Error(t, err)
}

func TestConsensus_ProcessProposal_With_Handler_OutOfGas(t *testing.T) {
	c, s := setUpConsensus(t, 0, cometmock.MockMempool[mock.Tx]{})
	c.processProposalHandler = handlers.NewDefaultProposalHandler(c.mempool).ProcessHandler()
	addQueryHandlerToSTF(t, s, func(ctx context.Context, q *consensustypes.QueryParamsRequest) (*consensustypes.QueryParamsResponse, error) {
		cParams := &v1.ConsensusParams{
			Block: &v1.BlockParams{
				MaxGas: 300_000,
			},
		}
		return &consensustypes.QueryParamsResponse{
			Params: cParams,
		}, nil
	})
	res, err := c.ProcessProposal(context.Background(), &abciproto.ProcessProposalRequest{
		Height: 1,
		Txs: [][]byte{invalidMockTx.Bytes(), invalidMockTx.Bytes(), invalidMockTx.Bytes()},
	})
	require.NoError(t, err)
	require.Equal(t, res.Status, abciproto.PROCESS_PROPOSAL_STATUS_REJECT)
}

func TestConsensus_ProcessProposal_With_Handler(t *testing.T) {
	c, s := setUpConsensus(t, 100_000, cometmock.MockMempool[mock.Tx]{})

	addQueryHandlerToSTF(t, s, func(ctx context.Context, q *consensustypes.QueryParamsRequest) (*consensustypes.QueryParamsResponse, error) {
		cParams := &v1.ConsensusParams{
			Block: &v1.BlockParams{
				MaxGas: 300_000,
			},
		}
		return &consensustypes.QueryParamsResponse{
			Params: cParams,
		}, nil
	})

	c.processProposalHandler = handlers.NewDefaultProposalHandler(c.mempool).ProcessHandler()

	// exeed max gas
	res, err := c.ProcessProposal(context.Background(), &abciproto.ProcessProposalRequest{
		Height: 1,
		Txs: [][]byte{mockTx.Bytes(), mockTx.Bytes(), mockTx.Bytes(), mockTx.Bytes()},
	})
	require.NoError(t, err)
	require.Equal(t, res.Status, abciproto.PROCESS_PROPOSAL_STATUS_REJECT)

	// have bad encode tx
	// should remove that tx and accept
	res, err = c.ProcessProposal(context.Background(), &abciproto.ProcessProposalRequest{
		Height: 1,
		Txs: [][]byte{mockTx.Bytes(), append(mockTx.Bytes(), []byte("bad")...), mockTx.Bytes(), mockTx.Bytes()},
	})
	fmt.Println(res, err)
	require.Equal(t, res.Status, abciproto.PROCESS_PROPOSAL_STATUS_ACCEPT)
}

func setUpConsensus(t *testing.T, gasLimit uint64, mempool mempool.Mempool[mock.Tx]) (*Consensus[mock.Tx], *stf.STF[mock.Tx]) {
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
			return kvSet(t, ctx, "init-chain")
		},
	}

	am, err := b.Build()
	require.NoError(t, err)

	return NewConsensus[mock.Tx](log.NewNopLogger(), "testing-app", "authority", am, mempool, map[string]struct{}{}, nil, mockStore, Config{AppTomlConfig: DefaultAppTomlConfig()}, mock.TxCodec{}, "test"), s
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
	bz, _ := store.GetStateCommitment().Get(actorName, version, []byte(key))
	require.Equal(t, len(bz), 0)
}
