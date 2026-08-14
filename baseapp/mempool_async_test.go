package baseapp_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/cockroachdb/errors"
	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/baseapp"
	baseapptestutil "github.com/cosmos/cosmos-sdk/baseapp/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/mempool"
)

// removeFaultyMempool is a minimal mempool whose removal always fails, either
// by panicking or by returning an error, simulating a broken or custom mempool
// implementation. Insert is a no-op so txs can pass CheckTx.
type removeFaultyMempool struct {
	failMode string // "panic" or "error"
}

func (m *removeFaultyMempool) Insert(context.Context, sdk.Tx, mempool.InsertOption) error {
	return nil
}

func (m *removeFaultyMempool) Select(context.Context, [][]byte) mempool.Iterator {
	return nil
}

func (m *removeFaultyMempool) SelectBy(context.Context, [][]byte, func(mempool.PooledTx) bool) {}

func (m *removeFaultyMempool) CountTx() int {
	return 0
}

func (m *removeFaultyMempool) Remove(sdk.Tx) error {
	switch m.failMode {
	case "panic":
		panic("mempool remove panic")
	case "error":
		return errors.New("mempool remove error")
	default:
		return nil
	}
}

func (m *removeFaultyMempool) RemoveWithReason(context.Context, sdk.Tx, mempool.RemoveReason) error {
	return m.Remove(nil)
}

// TestMempoolRemoveAsync_Panic verifies that a panic in a mempool's Remove
// during FinalizeBlock is contained by the async removal worker and cannot fail
// block execution (which would otherwise cause a consensus failure).
func TestMempoolRemoveAsync_Panic(t *testing.T) {
	suite := NewBaseAppSuite(t, baseapp.SetMempool(&removeFaultyMempool{failMode: "panic"}))
	baseapptestutil.RegisterCounterServer(suite.baseApp.MsgServiceRouter(), NoopCounterServerImpl{})

	_, err := suite.baseApp.InitChain(&abci.RequestInitChain{ConsensusParams: &cmtproto.ConsensusParams{}})
	require.NoError(t, err)

	tx := newTxCounter(t, suite.txConfig, 0, 1)
	txBytes, err := suite.txConfig.TxEncoder()(tx)
	require.NoError(t, err)

	// A panicking mempool removal must not fail block execution.
	res, err := suite.baseApp.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: 1,
		Txs:    [][]byte{txBytes},
	})
	require.NoError(t, err)
	require.True(t, res.TxResults[0].IsOK(), fmt.Sprintf("%v", res))

	// The app must remain functional for subsequent blocks.
	_, err = suite.baseApp.Commit()
	require.NoError(t, err)
	_, err = suite.baseApp.FinalizeBlock(&abci.RequestFinalizeBlock{Height: 2})
	require.NoError(t, err)
}

// TestMempoolRemoveAsync_Error verifies that an error returned by a mempool's
// Remove during FinalizeBlock cannot fail block execution.
func TestMempoolRemoveAsync_Error(t *testing.T) {
	suite := NewBaseAppSuite(t, baseapp.SetMempool(&removeFaultyMempool{failMode: "error"}))
	baseapptestutil.RegisterCounterServer(suite.baseApp.MsgServiceRouter(), NoopCounterServerImpl{})

	_, err := suite.baseApp.InitChain(&abci.RequestInitChain{ConsensusParams: &cmtproto.ConsensusParams{}})
	require.NoError(t, err)

	tx := newTxCounter(t, suite.txConfig, 0, 1)
	txBytes, err := suite.txConfig.TxEncoder()(tx)
	require.NoError(t, err)

	res, err := suite.baseApp.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: 1,
		Txs:    [][]byte{txBytes},
	})
	require.NoError(t, err)
	require.True(t, res.TxResults[0].IsOK(), fmt.Sprintf("%v", res))
}

// TestMempoolRemoveAsync_RecheckError verifies that a mempool removal error
// triggered by a failing recheck ante handler is not surfaced in the CheckTx
// response.
func TestMempoolRemoveAsync_RecheckError(t *testing.T) {
	anteOpt := func(bapp *baseapp.BaseApp) {
		bapp.SetAnteHandler(func(ctx sdk.Context, _ sdk.Tx, _ bool) (sdk.Context, error) {
			if ctx.IsReCheckTx() {
				return ctx, errors.New("recheck failed in ante handler")
			}
			return ctx, nil
		})
	}

	suite := NewBaseAppSuite(t, anteOpt, baseapp.SetMempool(&removeFaultyMempool{failMode: "error"}))
	baseapptestutil.RegisterCounterServer(suite.baseApp.MsgServiceRouter(), NoopCounterServerImpl{})

	_, err := suite.baseApp.InitChain(&abci.RequestInitChain{ConsensusParams: &cmtproto.ConsensusParams{}})
	require.NoError(t, err)

	tx := newTxCounter(t, suite.txConfig, 0, 1)
	txBytes, err := suite.txConfig.TxEncoder()(tx)
	require.NoError(t, err)

	_, err = suite.baseApp.CheckTx(&abci.RequestCheckTx{Tx: txBytes, Type: abci.CheckTxType_New})
	require.NoError(t, err)

	// The ante error is the only error surfaced: the mempool removal error is
	// handled asynchronously and must not be joined into the response.
	resp, err := suite.baseApp.CheckTx(&abci.RequestCheckTx{Tx: txBytes, Type: abci.CheckTxType_Recheck})
	require.NoError(t, err)
	require.True(t, resp.IsErr())
	require.Equal(t, "recheck failed in ante handler", resp.Log)
}
