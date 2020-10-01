package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/msg_authorization/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

// type TestSuite struct {
// 	suite.Suite
// 	ctx           sdk.Context
// 	accountKeeper authkeeper.AccountKeeper
// 	paramsKeeper  params.Keeper
// 	bankKeeper    bankkeeper.Keeper
// 	keeper        Keeper
// 	router        baseapp.Router
// }

type TestSuite struct {
	suite.Suite

	app *simapp.SimApp
	ctx sdk.Context
	// queryClient types.QueryClient
	addrs []sdk.AccAddress
}

func (s *TestSuite) SetupTest() {
	s.app = simapp.Setup(false)
	s.ctx = s.app.BaseApp.NewContext(false, tmproto.Header{})
	s.addrs = simapp.AddTestAddrsIncremental(s.app, s.ctx, 3, sdk.NewInt(20000001))
}

func (s *TestSuite) TestKeeper() {
	app, ctx, addrs := s.app, s.ctx, s.addrs
	granterAddr := addrs[0]
	granteeAddr := addrs[1]
	recipientAddr := addrs[2]
	err := app.BankKeeper.SetBalances(ctx, granterAddr, sdk.NewCoins(sdk.NewInt64Coin("steak", 10000)))
	s.Require().Nil(err)
	s.Require().True(app.BankKeeper.GetBalance(ctx, granterAddr, "steak").IsEqual(sdk.NewCoin("steak", sdk.NewInt(10000))))

	s.T().Log("verify that no authorization returns nil")
	authorization, expiration := app.MsgAuthKeeper.GetAuthorization(ctx, granteeAddr, granterAddr, banktypes.MsgSend{}.Type())
	s.Require().Nil(authorization)
	s.Require().Zero(expiration)
	now := s.ctx.BlockHeader().Time
	s.Require().NotNil(now)

	newCoins := sdk.NewCoins(sdk.NewInt64Coin("steak", 100))
	s.T().Log("verify if expired authorization is rejected")
	x := types.SendAuthorization{SpendLimit: newCoins}
	app.MsgAuthKeeper.Grant(ctx, granterAddr, granteeAddr, &x, now.Add(-1*time.Hour))
	authorization, _ = app.MsgAuthKeeper.GetAuthorization(ctx, granteeAddr, granterAddr, banktypes.MsgSend{}.Type())
	s.Require().Nil(authorization)
	s.T().Log("verify if authorization is accepted")
	x = types.SendAuthorization{SpendLimit: newCoins}
	app.MsgAuthKeeper.Grant(ctx, granteeAddr, granterAddr, &x, now.Add(time.Hour))
	authorization, exp := app.MsgAuthKeeper.GetAuthorization(ctx, granteeAddr, granterAddr, banktypes.MsgSend{}.Type())
	s.Require().NotNil(authorization)
	s.Require().Equal(authorization.MsgType(), banktypes.MsgSend{}.Type())
	s.Require().Equal(exp, now.Add(time.Hour).Unix())

	s.T().Log("verify fetching authorization with wrong msg type fails")
	authorization, _ = app.MsgAuthKeeper.GetAuthorization(ctx, granteeAddr, granterAddr, banktypes.MsgMultiSend{}.Type())
	s.Require().Nil(authorization)

	s.T().Log("verify fetching authorization with wrong grantee fails")
	authorization, _ = app.MsgAuthKeeper.GetAuthorization(ctx, recipientAddr, granterAddr, banktypes.MsgMultiSend{}.Type())
	s.Require().Nil(authorization)

	s.T().Log("")

	s.T().Log("verify revoke fails with wrong information")
	app.MsgAuthKeeper.Revoke(ctx, recipientAddr, granterAddr, banktypes.MsgSend{}.Type())
	authorization, _ = app.MsgAuthKeeper.GetAuthorization(ctx, recipientAddr, granterAddr, banktypes.MsgSend{}.Type())
	s.Require().Nil(authorization)

	s.T().Log("verify revoke executes with correct information")
	app.MsgAuthKeeper.Revoke(ctx, recipientAddr, granterAddr, banktypes.MsgSend{}.Type())
	authorization, _ = app.MsgAuthKeeper.GetAuthorization(ctx, granteeAddr, granterAddr, banktypes.MsgSend{}.Type())
	s.Require().NotNil(authorization)

}

func TestTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
