package keeper_test

import (
	gocontext "context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
)

type SlashingTestSuite struct {
	suite.Suite

	app         *simapp.SimApp
	ctx         sdk.Context
	queryClient types.QueryClient
	addrDels    []sdk.AccAddress
}

func (suite *SlashingTestSuite) SetupTest() {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, abci.Header{})

	app.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())
	app.BankKeeper.SetSendEnabled(ctx, true)
	app.SlashingKeeper.SetParams(ctx, keeper.TestParams())

	addrDels := simapp.AddTestAddrsIncremental(app, ctx, 2, sdk.TokensFromConsensusPower(200))

	info1 := types.NewValidatorSigningInfo(sdk.ConsAddress(addrDels[0]), int64(4), int64(3),
		time.Unix(2, 0), false, int64(10))
	info2 := types.NewValidatorSigningInfo(sdk.ConsAddress(addrDels[1]), int64(5), int64(4),
		time.Unix(2, 0), false, int64(10))
	app.SlashingKeeper.SetValidatorSigningInfo(ctx, sdk.ConsAddress(addrDels[0]), info1)
	app.SlashingKeeper.SetValidatorSigningInfo(ctx, sdk.ConsAddress(addrDels[1]), info2)

	suite.app = app
	suite.ctx = ctx
	suite.addrDels = addrDels

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, app.SlashingKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	suite.queryClient = queryClient
}

func (suite *SlashingTestSuite) TestGRPCQueryParams() {
	queryClient := suite.queryClient
	var paramsResp, err = queryClient.Params(gocontext.Background(), nil)
	suite.Error(err)
	suite.Nil(paramsResp)

	paramsResp, err = queryClient.Params(gocontext.Background(), &types.QueryParamsRequest{})

	suite.NoError(err)
	suite.Equal(keeper.TestParams(), paramsResp.Params)
}

func (suite *SlashingTestSuite) TestGRPCSigningInfo() {
	queryClient := suite.queryClient

	infoResp, err := queryClient.SigningInfo(gocontext.Background(), &types.QuerySigningInfoRequest{ConsAddress: nil})
	suite.Error(err)
	suite.Nil(infoResp)

	consAddr := sdk.ConsAddress(suite.addrDels[0])
	info, found := suite.app.SlashingKeeper.GetValidatorSigningInfo(suite.ctx, consAddr)
	suite.True(found)

	infoResp, err = queryClient.SigningInfo(gocontext.Background(),
		&types.QuerySigningInfoRequest{ConsAddress: consAddr})
	suite.NoError(err)
	suite.Equal(&info, infoResp.ValSigningInfo)
}

func TestSlashingTestSuite(t *testing.T) {
	suite.Run(t, new(SlashingTestSuite))
}
