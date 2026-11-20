package baseapp_test

import (
	"context"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"

	baseapptestutil "github.com/cosmos/cosmos-sdk/baseapp/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type internalServer struct{ calls *int }

func (s internalServer) DoInternal(_ context.Context, _ *baseapptestutil.MsgInternal) (*baseapptestutil.MsgInternalResponse, error) {
	*s.calls++
	return &baseapptestutil.MsgInternalResponse{}, nil
}

func setupInternalMsgSuite(t *testing.T) (*BaseAppSuite, *int) {
	t.Helper()

	suite := NewBaseAppSuite(t)
	calls := new(int)
	baseapptestutil.RegisterInternalServer(suite.baseApp.MsgServiceRouter(), internalServer{calls: calls})
	baseapptestutil.RegisterCounterServer(suite.baseApp.MsgServiceRouter(), CounterServerImplGasMeterOnly{})

	_, err := suite.baseApp.InitChain(&abci.RequestInitChain{ConsensusParams: &cmtproto.ConsensusParams{}})
	require.NoError(t, err)

	return suite, calls
}

func internalTxBytes(t *testing.T, suite *BaseAppSuite) []byte {
	t.Helper()

	_, _, addr := testdata.KeyTestPubAddr()
	builder := suite.txConfig.NewTxBuilder()
	require.NoError(t, builder.SetMsgs(&baseapptestutil.MsgInternal{Signer: addr.String()}))
	builder.SetMemo("counter=0&failOnAnte=false")
	setTxSignature(t, builder, 0)

	bz, err := suite.txConfig.TxEncoder()(builder.GetTx())
	require.NoError(t, err)

	return bz
}

func TestInternalMsg_AnnotationIsRead(t *testing.T) {
	suite, _ := setupInternalMsgSuite(t)
	router := suite.baseApp.MsgServiceRouter()

	require.True(t, router.IsInternal(sdk.MsgTypeURL(&baseapptestutil.MsgInternal{})))
	require.False(t, router.IsInternal(sdk.MsgTypeURL(&baseapptestutil.MsgCounter{})))
}

func TestInternalMsg_RejectedAtTxBoundary(t *testing.T) {
	t.Run("CheckTx", func(t *testing.T) {
		suite, calls := setupInternalMsgSuite(t)
		res, err := suite.baseApp.CheckTx(&abci.RequestCheckTx{Tx: internalTxBytes(t, suite)})
		require.NoError(t, err)
		require.NotZero(t, res.Code)
		require.Contains(t, res.Log, "internal-only")
		require.Zero(t, *calls)
	})

	t.Run("Simulate", func(t *testing.T) {
		suite, calls := setupInternalMsgSuite(t)
		_, _, err := suite.baseApp.Simulate(internalTxBytes(t, suite))
		require.ErrorContains(t, err, "internal-only")
		require.Zero(t, *calls)
	})

	t.Run("FinalizeBlock", func(t *testing.T) {
		suite, calls := setupInternalMsgSuite(t)
		res, err := suite.baseApp.FinalizeBlock(&abci.RequestFinalizeBlock{
			Height: 1,
			Txs:    [][]byte{internalTxBytes(t, suite)},
		})
		require.NoError(t, err)
		require.Len(t, res.TxResults, 1)
		require.NotZero(t, res.TxResults[0].Code)
		require.Contains(t, res.TxResults[0].Log, "internal-only")
		require.Zero(t, *calls)
	})

	t.Run("control: external msg passes CheckTx", func(t *testing.T) {
		suite, _ := setupInternalMsgSuite(t)
		txBytes, err := suite.txConfig.TxEncoder()(newTxCounter(t, suite.txConfig, 0, 0))
		require.NoError(t, err)
		res, err := suite.baseApp.CheckTx(&abci.RequestCheckTx{Tx: txBytes})
		require.NoError(t, err)
		require.Zero(t, res.Code, res.Log)
	})
}
