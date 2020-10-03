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
	tmtime "github.com/tendermint/tendermint/types/time"

)

type TestSuite struct {
	suite.Suite

	app *simapp.SimApp
	ctx sdk.Context
	// queryClient types.QueryClient
	addrs []sdk.AccAddress
}

func (s *TestSuite) SetupTest() {
	s.app = simapp.Setup(false)
	ctx := s.app.BaseApp.NewContext(false, tmproto.Header{})
	now := tmtime.Now()
	ctx = ctx.WithBlockHeader(tmproto.Header{Time: now})
	s.ctx = ctx
	s.addrs = simapp.AddTestAddrsIncremental(s.app, s.ctx, 3, sdk.NewInt(20000001))

	// s.ctx, s.accountKeeper, s.paramsKeeper, s.bankKeeper, s.keeper, s.router = SetupTestInput()
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
	x := &types.SendAuthorization{SpendLimit: newCoins}
	app.MsgAuthKeeper.Grant(ctx, granterAddr, granteeAddr, x, now.Add(-1*time.Hour))
	authorization, _ = app.MsgAuthKeeper.GetAuthorization(ctx, granteeAddr, granterAddr, banktypes.MsgSend{}.Type())
	s.Require().Nil(authorization)

	s.T().Log("verify if authorization is accepted")
	x = &types.SendAuthorization{SpendLimit: newCoins}
	app.MsgAuthKeeper.Grant(ctx, granteeAddr, granterAddr, x, now.Add(time.Hour))
	authorization, _ = app.MsgAuthKeeper.GetAuthorization(ctx, granteeAddr, granterAddr, banktypes.MsgSend{}.Type())
	s.Require().NotNil(authorization)
	s.Require().Equal(authorization.MsgType(), banktypes.MsgSend{}.Type())

	s.T().Log("verify fetching authorization with wrong msg type fails")
	authorization, _ = app.MsgAuthKeeper.GetAuthorization(ctx, granteeAddr, granterAddr, banktypes.MsgMultiSend{}.Type())
	s.Require().Nil(authorization)

	s.T().Log("verify fetching authorization with wrong grantee fails")
	authorization, _ = app.MsgAuthKeeper.GetAuthorization(ctx, recipientAddr, granterAddr, banktypes.MsgMultiSend{}.Type())
	s.Require().Nil(authorization)

	s.T().Log("verify revoke fails with wrong information")
	err = app.MsgAuthKeeper.Revoke(ctx, recipientAddr, granterAddr, banktypes.MsgMultiSend{}.Type())
	s.Require().Error(err)
	authorization, _ = app.MsgAuthKeeper.GetAuthorization(ctx, recipientAddr, granterAddr, banktypes.MsgSend{}.Type())
	s.Require().Nil(authorization)

	s.T().Log("verify revoke executes with correct information")
	err = app.MsgAuthKeeper.Revoke(ctx, granteeAddr, granterAddr, banktypes.MsgSend{}.Type())
	s.Require().NoError(err)
	authorization, _ = app.MsgAuthKeeper.GetAuthorization(ctx, granteeAddr, granterAddr, banktypes.MsgSend{}.Type())
	s.Require().Nil(authorization)

}

func (s *TestSuite) TestKeeperFees() {
	app, addrs := s.app, s.addrs

	granterAddr := addrs[0]
	granteeAddr := addrs[1]
	recipientAddr := addrs[2]
	err := app.BankKeeper.SetBalances(s.ctx, granterAddr, sdk.NewCoins(sdk.NewInt64Coin("steak", 10000)))
	s.Require().Nil(err)
	s.Require().True(app.BankKeeper.GetBalance(s.ctx, granterAddr, "steak").IsEqual(sdk.NewCoin("steak", sdk.NewInt(10000))))

	now := s.ctx.BlockHeader().Time
	s.Require().NotNil(now)

	smallCoin := sdk.NewCoins(sdk.NewInt64Coin("steak", 20))
	someCoin := sdk.NewCoins(sdk.NewInt64Coin("steak", 123))
	//lotCoin := sdk.NewCoins(sdk.NewInt64Coin("steak", 4567))

	msgs := types.NewMsgExecAuthorized(granteeAddr, []sdk.Msg{
		&banktypes.MsgSend{
			Amount:      sdk.NewCoins(sdk.NewInt64Coin("steak", 2)),
			FromAddress: granterAddr.String(),
			ToAddress:   recipientAddr.String(),
		},
	})

	s.T().Log("verify dispatch fails with invalid authorization")
	executeMsgs, err := msgs.GetMsgs()
	s.Require().NoError(err)
	result, error := app.MsgAuthKeeper.DispatchActions(s.ctx, granteeAddr, executeMsgs)

	s.Require().Nil(result)
	s.Require().NotNil(error)

	s.T().Log("verify dispatch executes with correct information")
	// grant authorization
	app.MsgAuthKeeper.Grant(s.ctx, granteeAddr, granterAddr, &types.SendAuthorization{SpendLimit: smallCoin}, now)
	authorization, expiration := app.MsgAuthKeeper.GetAuthorization(s.ctx, granteeAddr, granterAddr, banktypes.MsgSend{}.Type())
	s.Require().NotNil(authorization)
	s.Require().Zero(expiration)
	s.Require().Equal(authorization.MsgType(), banktypes.MsgSend{}.Type())

	executeMsgs, err = msgs.GetMsgs()
	s.Require().NoError(err)

	result, error = app.MsgAuthKeeper.DispatchActions(s.ctx, granteeAddr, executeMsgs)
	s.Require().NotNil(result)
	s.Require().Nil(error)

	authorization, _ = app.MsgAuthKeeper.GetAuthorization(s.ctx, granteeAddr, granterAddr, banktypes.MsgSend{}.Type())
	s.Require().NotNil(authorization)

	s.T().Log("verify dispatch fails with overlimit")
	// grant authorization

	msgs = types.NewMsgExecAuthorized(granteeAddr, []sdk.Msg{
		&banktypes.MsgSend{
			Amount:      someCoin,
			FromAddress: granterAddr.String(),
			ToAddress:   recipientAddr.String(),
		},
	})

	executeMsgs, err = msgs.GetMsgs()
	s.Require().NoError(err)

	result, error = app.MsgAuthKeeper.DispatchActions(s.ctx, granteeAddr, executeMsgs)
	s.Require().Nil(result)
	s.Require().NotNil(error)

	authorization, _ = app.MsgAuthKeeper.GetAuthorization(s.ctx, granteeAddr, granterAddr, banktypes.MsgSend{}.Type())
	s.Require().NotNil(authorization)
}

func TestTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}