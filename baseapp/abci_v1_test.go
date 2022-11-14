package baseapp_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"cosmossdk.io/depinject"
	"github.com/cosmos/cosmos-sdk/baseapp"
	baseapptestutil "github.com/cosmos/cosmos-sdk/baseapp/testutil"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/mempool"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
)

type NoopCounterServerImpl struct{}

func (m NoopCounterServerImpl) IncrementCounter(
	_ context.Context,
	_ *baseapptestutil.MsgCounter,
) (*baseapptestutil.MsgCreateCounterResponse, error) {
	return &baseapptestutil.MsgCreateCounterResponse{}, nil
}

type ABCIv1TestSuite struct {
	suite.Suite
	baseApp  *baseapp.BaseApp
	mempool  mempool.Mempool
	txConfig client.TxConfig
	cdc      codec.ProtoCodecMarshaler
	options  []func(app *baseapp.BaseApp)
}

func TestABCIv1TestSuite(t *testing.T) {
	suite.Run(t, new(ABCIv1TestSuite))
}

func (s *ABCIv1TestSuite) SetupTest() {
	t := s.T()
	anteKey := []byte("ante-key")
	pool := mempool.NewNonceMempool()
	anteOpt := func(bapp *baseapp.BaseApp) {
		bapp.SetAnteHandler(anteHandlerTxTest(t, capKey1, anteKey))
	}

	var (
		appBuilder *runtime.AppBuilder
		cdc        codec.ProtoCodecMarshaler
		registry   codectypes.InterfaceRegistry
	)
	err := depinject.Inject(makeMinimalConfig(), &appBuilder, &cdc, &registry)
	require.NoError(t, err)

	s.options = append(s.options, baseapp.SetMempool(pool), anteOpt)
	app := setupBaseApp(t, s.options...)
	baseapptestutil.RegisterInterfaces(registry)
	app.SetMsgServiceRouter(baseapp.NewMsgServiceRouter())
	app.SetInterfaceRegistry(registry)

	baseapptestutil.RegisterKeyValueServer(app.MsgServiceRouter(), MsgKeyValueImpl{})
	baseapptestutil.RegisterCounterServer(app.MsgServiceRouter(), NoopCounterServerImpl{})
	header := tmproto.Header{Height: app.LastBlockHeight() + 1}

	app.InitChain(abci.RequestInitChain{
		ConsensusParams: &tmproto.ConsensusParams{},
	})

	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	// patch in TxConfig insted of using an output from x/auth/tx
	txConfig := authtx.NewTxConfig(cdc, authtx.DefaultSignModes)

	app.SetTxDecoder(txConfig.TxDecoder())
	app.SetTxEncoder(txConfig.TxEncoder())

	s.baseApp = app
	s.mempool = pool
	s.txConfig = txConfig
	s.cdc = cdc
}

func (s *ABCIv1TestSuite) TestABCIv1_HappyPath() {
	txConfig := s.txConfig
	t := s.T()

	tx := newTxCounter(txConfig, 0, 1)
	txBytes, err := txConfig.TxEncoder()(tx)
	require.NoError(t, err)

	reqCheckTx := abci.RequestCheckTx{
		Tx:   txBytes,
		Type: abci.CheckTxType_New,
	}
	s.baseApp.CheckTx(reqCheckTx)

	tx2 := newTxCounter(txConfig, 1, 1)

	tx2Bytes, err := txConfig.TxEncoder()(tx2)
	require.NoError(t, err)

	err = s.mempool.Insert(sdk.Context{}, tx2)
	require.NoError(t, err)
	reqPreparePropossal := abci.RequestPrepareProposal{
		MaxTxBytes: 1000,
	}
	resPreparePropossal := s.baseApp.PrepareProposal(reqPreparePropossal)

	require.Equal(t, 2, len(resPreparePropossal.Txs))

	var reqProposalTxBytes [2][]byte
	reqProposalTxBytes[0] = txBytes
	reqProposalTxBytes[1] = tx2Bytes
	reqProcessProposal := abci.RequestProcessProposal{
		Txs: reqProposalTxBytes[:],
	}

	resProcessProposal := s.baseApp.ProcessProposal(reqProcessProposal)
	require.Equal(t, abci.ResponseProcessProposal_ACCEPT, resProcessProposal.Status)

	res := s.baseApp.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
	require.Equal(t, 1, s.mempool.CountTx())

	require.NotEmpty(t, res.Events)
	require.True(t, res.IsOK(), fmt.Sprintf("%v", res))
}

func (s *ABCIv1TestSuite) TestABCIv1_PrepareProposal_ReachedMaxBytes() {
	txConfig := s.txConfig
	t := s.T()

	for i := 0; i < 100; i++ {
		tx2 := newTxCounter(txConfig, int64(i), int64(i))
		err := s.mempool.Insert(sdk.Context{}, tx2)
		require.NoError(t, err)
	}

	reqPreparePropossal := abci.RequestPrepareProposal{
		MaxTxBytes: 1500,
	}
	resPreparePropossal := s.baseApp.PrepareProposal(reqPreparePropossal)

	require.Equal(t, 10, len(resPreparePropossal.Txs))
}

func (s *ABCIv1TestSuite) TestABCIv1_PrepareProposal_BadEncoding() {
	txConfig := authtx.NewTxConfig(s.cdc, authtx.DefaultSignModes)

	t := s.T()

	tx := newTxCounter(txConfig, 0, 0)
	err := s.mempool.Insert(sdk.Context{}, tx)
	require.NoError(t, err)

	reqPrepareProposal := abci.RequestPrepareProposal{
		MaxTxBytes: 1000,
	}
	resPrepareProposal := s.baseApp.PrepareProposal(reqPrepareProposal)

	require.Equal(t, 1, len(resPrepareProposal.Txs))
}

func (s *ABCIv1TestSuite) TestABCIv1_PrepareProposal_Failures() {
	tx := newTxCounter(s.txConfig, 0, 0)
	txBytes, err := s.txConfig.TxEncoder()(tx)
	s.NoError(err)

	reqCheckTx := abci.RequestCheckTx{
		Tx:   txBytes,
		Type: abci.CheckTxType_New,
	}
	checkTxRes := s.baseApp.CheckTx(reqCheckTx)
	s.True(checkTxRes.IsOK())

	failTx := newTxCounter(s.txConfig, 1, 1)
	failTx = setFailOnAnte(s.txConfig, failTx, true)
	err = s.mempool.Insert(sdk.Context{}, failTx)
	s.NoError(err)
	s.Equal(2, s.mempool.CountTx())

	req := abci.RequestPrepareProposal{
		MaxTxBytes: 1000,
	}
	res := s.baseApp.PrepareProposal(req)
	s.Equal(1, len(res.Txs))
}
