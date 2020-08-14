package simulate_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/grpc/simulate"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/std"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

type IntegrationTestSuite struct {
	suite.Suite

	app         *simapp.SimApp
	clientCtx   client.Context
	queryClient simulate.SimulateServiceClient
	sdkCtx      sdk.Context
}

func (s *IntegrationTestSuite) SetupTest() {
	app := simapp.Setup(false)
	sdkCtx := app.BaseApp.NewContext(false, abci.Header{})
	app.AccountKeeper.SetParams(sdkCtx, authtypes.DefaultParams())

	// Set up TxConfig.
	encodingConfig := simapp.MakeEncodingConfig()
	pubKeyCodec := std.DefaultPublicKeyCodec{}
	clientCtx := client.Context{}.WithTxConfig(encodingConfig.TxConfig)

	// Create new simulation server.
	srv := simulate.NewSimulateServer(*app.BaseApp, pubKeyCodec, clientCtx.TxConfig)

	queryHelper := baseapp.NewQueryServerTestHelper(sdkCtx, app.InterfaceRegistry())
	simulate.RegisterSimulateServiceServer(queryHelper, srv)
	queryClient := simulate.NewSimulateServiceClient(queryHelper)

	s.app = app
	s.clientCtx = clientCtx
	s.queryClient = queryClient
	s.sdkCtx = sdkCtx
}

func (s IntegrationTestSuite) TestSimulateService() {
	// Create an account with some funds.
	priv, _, addr := testdata.KeyTestPubAddr()
	acc := s.app.AccountKeeper.NewAccountWithAddress(s.sdkCtx, addr)
	err := acc.SetAccountNumber(0)
	s.Require().NoError(err)
	s.app.AccountKeeper.SetAccount(s.sdkCtx, acc)
	s.app.BankKeeper.SetBalances(s.sdkCtx, addr, sdk.Coins{
		sdk.NewInt64Coin("atom", 10000000),
	})

	// Create a test x/bank MsgSend.
	addr1 := sdk.AccAddress([]byte("from"))
	addr2 := sdk.AccAddress([]byte("to"))
	coins := sdk.NewCoins(sdk.NewInt64Coin("atom", 10))
	msg := banktypes.NewMsgSend(addr1, addr2, coins)
	feeAmount := testdata.NewTestFeeAmount()
	gasLimit := testdata.NewTestGasLimit()
	memo := "foo"

	// Create a txBuilder.
	txBuilder := s.clientCtx.TxConfig.NewTxBuilder()
	txBuilder.SetMsgs(msg)
	txBuilder.SetMemo(memo)
	txBuilder.SetFeeAmount(feeAmount)
	txBuilder.SetGasLimit(gasLimit)
	// 1st round: set empty signature
	sigV2 := signing.SignatureV2{
		PubKey: priv.PubKey(),
		Data: &signing.SingleSignatureData{
			SignMode:  s.clientCtx.TxConfig.SignModeHandler().DefaultMode(),
			Signature: nil,
		},
	}
	txBuilder.SetSignatures(sigV2)
	// 2nd round: actually sign
	sigV2, err = tx.SignWithPrivKey(
		s.clientCtx.TxConfig.SignModeHandler().DefaultMode(),
		authsigning.SignerData{ChainID: s.sdkCtx.ChainID(), AccountNumber: 0, AccountSequence: 0},
		txBuilder, priv, s.clientCtx.TxConfig,
	)
	txBuilder.SetSignatures(sigV2)

	res, err := s.queryClient.Simulate(
		context.Background(),
		&simulate.SimulateRequest{Tx: txBuilder.GetProtoTx()},
	)
	s.Require().NoError(err)

	// TODO Better test
	s.Require().NotEmpty(res)
}

func TestSimulateTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
