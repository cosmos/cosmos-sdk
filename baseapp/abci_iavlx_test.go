package baseapp_test

import (
	"context"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/store/rootmulti"

	"github.com/cosmos/cosmos-sdk/baseapp"
	baseapptestutil "github.com/cosmos/cosmos-sdk/baseapp/testutil"
	iavlx "github.com/cosmos/cosmos-sdk/iavl"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestABCI_Query_IAVLXQueryableProof(t *testing.T) {
	key, value := []byte("hello"), []byte("goodbye")

	var cms *iavlx.CommitMultiTree
	setCMSOpt := func(bapp *baseapp.BaseApp) {
		var err error
		cms, err = iavlx.LoadCommitMultiTree(t.TempDir(), iavlx.Options{})
		require.NoError(t, err)
		bapp.SetCMS(cms)
	}
	anteOpt := func(bapp *baseapp.BaseApp) {
		bapp.SetAnteHandler(func(ctx sdk.Context, tx sdk.Tx, simulate bool) (newCtx sdk.Context, err error) {
			ctx.KVStore(capKey1).Set(key, value)
			return ctx, nil
		})
	}

	suite := NewBaseAppSuite(t, setCMSOpt, anteOpt)
	t.Cleanup(func() {
		if cms != nil {
			_ = cms.Close()
		}
	})

	baseapptestutil.RegisterCounterServer(suite.baseApp.MsgServiceRouter(), CounterServerImplGasMeterOnly{})
	_, err := suite.baseApp.InitChain(&abci.RequestInitChain{
		ConsensusParams: &cmtproto.ConsensusParams{},
	})
	require.NoError(t, err)

	tx := newTxCounter(t, suite.txConfig, 0, 0)
	bz, err := suite.txConfig.TxEncoder()(tx)
	require.NoError(t, err)

	_, err = suite.baseApp.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: 1,
		Txs:    [][]byte{bz},
	})
	require.NoError(t, err)
	_, err = suite.baseApp.Commit()
	require.NoError(t, err)

	_, err = suite.baseApp.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: 2,
		Txs:    [][]byte{bz},
	})
	require.NoError(t, err)
	_, err = suite.baseApp.Commit()
	require.NoError(t, err)

	res, err := suite.baseApp.Query(context.TODO(), &abci.RequestQuery{
		Path:   "/store/key1/key",
		Data:   key,
		Height: 2,
		Prove:  true,
	})
	require.NoError(t, err)
	require.Equal(t, value, res.Value)
	require.NotNil(t, res.ProofOps)
	require.Len(t, res.ProofOps.Ops, 2)
	require.Empty(t, res.Log)

	prt := rootmulti.DefaultProofRuntime()
	require.NoError(t, prt.VerifyValue(res.ProofOps, suite.baseApp.LastCommitID().Hash, "/key1/hello", value))
}
