package reflection_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/grpc/reflection"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/simapp"
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
	queryClient reflection.
	sdkCtx      sdk.Context
}

func (s *IntegrationTestSuite) SetupTest() {
	app := simapp.Setup(false)
	sdkCtx := app.BaseApp.NewContext(false, abci.Header{})

	srv := 

	queryHelper := baseapp.NewQueryServerTestHelper(sdkCtx, app.InterfaceRegistry())
	reflection.RegisterSimulateServiceServer(queryHelper, srv)
	queryClient := reflection.NewSimulateServiceClient(queryHelper)

	s.app = app
	s.clientCtx = client.Context{}.
		WithTxConfig(encodingConfig.TxConfig)
	s.queryClient = queryClient
	s.sdkCtx = sdkCtx
}

func (s IntegrationTestSuite) TestSimulateService() {
	
	res, err := s.queryClient.Simulate(
		context.Background(),
		&reflection.SimulateRequest{Tx: txBuilder.GetProtoTx()},
	)
	s.Require().NoError(err)

	fmt.Println(res)
}

func TestSimulateTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}