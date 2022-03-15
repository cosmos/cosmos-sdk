package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/tieredfee/types"
)

type IntegrationTestSuite struct {
	suite.Suite

	app         *simapp.SimApp
	ctx         sdk.Context
	queryClient types.QueryClient
	clientCtx   client.Context
}

func (suite *IntegrationTestSuite) SetupTest() {
	app := simapp.Setup(suite.T(), false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{Time: time.Now()})

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, app.TieredFeeKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	suite.app = app
	suite.ctx = ctx
	suite.queryClient = queryClient

	// Set up TxConfig.
	encodingConfig := simapp.MakeTestEncodingConfig()
	suite.clientCtx = client.Context{}.
		WithTxConfig(encodingConfig.TxConfig).
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithCodec(codec.NewAminoCodec(encodingConfig.Amino))
}

func (suite *IntegrationTestSuite) TestSimple() {
	suite.SetupTest() // reset
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
