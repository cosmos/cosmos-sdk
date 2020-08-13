package simulate_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/grpc/simulate"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

type IntegrationTestSuite struct {
	suite.Suite

	clientCtx   client.Context
	queryClient simulate.SimulateServiceClient
	txConfig    client.TxConfig
}

func (s *IntegrationTestSuite) SetupTest() {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, abci.Header{})
	app.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())

	srv := simulate.NewSimulateServer(*app.BaseApp)

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	simulate.RegisterSimulateServiceServer(queryHelper, srv)
	queryClient := simulate.NewSimulateServiceClient(queryHelper)

	// Set up TxConfig.
	encodingConfig := simapp.MakeEncodingConfig()

	s.clientCtx = client.Context{}.
		WithTxConfig(encodingConfig.TxConfig)

	s.queryClient = queryClient
}

func (s IntegrationTestSuite) TestSimulateService() {
	// Create a test x/bank MsgSend
	addr1 := sdk.AccAddress([]byte("from"))
	addr2 := sdk.AccAddress([]byte("to"))
	coins := sdk.NewCoins(sdk.NewInt64Coin("atom", 10))
	msg := banktypes.NewMsgSend(addr1, addr2, coins)
	feeAmount := testdata.NewTestFeeAmount()
	gasLimit := testdata.NewTestGasLimit()
	memo := "foo"

	txBuilder := s.clientCtx.TxConfig.NewTxBuilder()
	txBuilder.SetMsgs(msg)
	txBuilder.SetMemo(memo)
	txBuilder.SetFeeAmount(feeAmount)
	txBuilder.SetGasLimit(gasLimit)
	tx := txBuilder.GetProtoTx()

	res, err := s.queryClient.Simulate(
		context.Background(),
		&simulate.SimulateRequest{Tx: tx},
	)
	s.Require().NoError(err)

	fmt.Println(res)
}

func TestSimulateTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
