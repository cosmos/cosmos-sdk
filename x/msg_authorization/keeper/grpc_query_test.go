package keeper_test

import (
	gocontext "context"
	"time"
	"fmt"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	tmtime "github.com/tendermint/tendermint/types/time"
	"github.com/cosmos/cosmos-sdk/x/msg_authorization/types"
)


type MsgAuthTestSuite struct {
	suite.Suite

	app         *simapp.SimApp
	ctx         sdk.Context
	queryClient types.QueryClient
	accAddrs    []sdk.AccAddress
}


func (suite *MsgAuthTestSuite) SetupTest() {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	now := tmtime.Now()
	ctx = ctx.WithBlockHeader(tmproto.Header{Time: now})
	
	accAddrs := simapp.AddTestAddrsIncremental(app, ctx, 2, sdk.TokensFromConsensusPower(200))
	suite.app = app
	suite.ctx = ctx
	suite.accAddrs = accAddrs

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, app.MsgAuthKeeper)
	queryClient := types.NewQueryClient(queryHelper)
	suite.queryClient = queryClient
}

func (suite *MsgAuthTestSuite) TestGRPCQueryAuthorization() {
	suite.T().Log("Here")
	queryClient := suite.queryClient

	now := suite.ctx.BlockHeader().Time
	newCoins := sdk.NewCoins(sdk.NewInt64Coin("steak", 100))
	x := &types.SendAuthorization{SpendLimit: newCoins}
	suite.app.MsgAuthKeeper.Grant(suite.ctx, suite.accAddrs[0], suite.accAddrs[1],x,now.Add(time.Hour) )

	paramsResp, err := queryClient.Authorization(gocontext.Background(), &types.QueryAuthorizationRequest{})

	suite.NoError(err)

	authorization, _ := suite.app.MsgAuthKeeper.GetAuthorization(suite.ctx, suite.accAddrs[0], suite.accAddrs[1],x.MsgType())
	suite.NotNil(authorization)
	fmt.Println(authorization)
	suite.Equal(x, paramsResp.GetAuthorization())
}
