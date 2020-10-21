package keeper_test

import (
	gocontext "context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/slashing/testslashing"
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
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	app.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())
	app.BankKeeper.SetParams(ctx, banktypes.DefaultParams())
	app.SlashingKeeper.SetParams(ctx, testslashing.TestParams())

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
	paramsResp, err := queryClient.Params(gocontext.Background(), &types.QueryParamsRequest{})

	suite.NoError(err)
	suite.Equal(testslashing.TestParams(), paramsResp.Params)
}

func (suite *SlashingTestSuite) TestGRPCSigningInfo() {
	queryClient := suite.queryClient

	infoResp, err := queryClient.SigningInfo(gocontext.Background(), &types.QuerySigningInfoRequest{ConsAddress: ""})
	suite.Error(err)
	suite.Nil(infoResp)

	consAddr := sdk.ConsAddress(suite.addrDels[0])
	info, found := suite.app.SlashingKeeper.GetValidatorSigningInfo(suite.ctx, consAddr)
	suite.True(found)

	infoResp, err = queryClient.SigningInfo(gocontext.Background(),
		&types.QuerySigningInfoRequest{ConsAddress: consAddr.String()})
	suite.NoError(err)
	suite.Equal(info, infoResp.ValSigningInfo)
}

func (suite *SlashingTestSuite) TestGRPCSigningInfos() {
	queryClient := suite.queryClient

	var signingInfos []types.ValidatorSigningInfo

	suite.app.SlashingKeeper.IterateValidatorSigningInfos(suite.ctx, func(consAddr sdk.ConsAddress, info types.ValidatorSigningInfo) (stop bool) {
		signingInfos = append(signingInfos, info)
		return false
	})

	// verify all values are returned without pagination
	var infoResp, err = queryClient.SigningInfos(gocontext.Background(),
		&types.QuerySigningInfosRequest{Pagination: nil})
	suite.NoError(err)
	suite.Equal(signingInfos, infoResp.Info)

	infoResp, err = queryClient.SigningInfos(gocontext.Background(),
		&types.QuerySigningInfosRequest{Pagination: &query.PageRequest{Limit: 1, CountTotal: true}})
	suite.NoError(err)
	suite.Len(infoResp.Info, 1)
	suite.Equal(signingInfos[0], infoResp.Info[0])
	suite.NotNil(infoResp.Pagination.NextKey)
	suite.Equal(uint64(2), infoResp.Pagination.Total)
}

func TestSlashingTestSuite(t *testing.T) {
	suite.Run(t, new(SlashingTestSuite))
}
