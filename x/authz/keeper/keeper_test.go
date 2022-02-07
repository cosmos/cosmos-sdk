package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	tmtime "github.com/tendermint/tendermint/libs/time"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

var bankSendAuthMsgType = banktypes.SendAuthorization{}.MsgTypeURL()

type TestSuite struct {
	suite.Suite

	app         *simapp.SimApp
	ctx         sdk.Context
	addrs       []sdk.AccAddress
	queryClient authz.QueryClient
}

func (s *TestSuite) SetupTest() {
	app := simapp.Setup(s.T(), false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	now := tmtime.Now()
	ctx = ctx.WithBlockHeader(tmproto.Header{Time: now})
	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	authz.RegisterQueryServer(queryHelper, app.AuthzKeeper)
	queryClient := authz.NewQueryClient(queryHelper)
	s.queryClient = queryClient

	s.app = app
	s.ctx = ctx
	s.queryClient = queryClient
	s.addrs = simapp.AddTestAddrsIncremental(app, ctx, 3, sdk.NewInt(30000000))
}

func (s *TestSuite) TestKeeper() {
	app, ctx, addrs := s.app, s.ctx, s.addrs

	granterAddr := addrs[0]
	granteeAddr := addrs[1]
	recipientAddr := addrs[2]

	s.T().Log("verify that no authorization returns nil")
	authorization, expiration := app.AuthzKeeper.GetCleanAuthorization(ctx, granteeAddr, granterAddr, bankSendAuthMsgType)
	s.Require().Nil(authorization)
	s.Require().Equal(expiration, time.Time{})
	now := s.ctx.BlockHeader().Time

	newCoins := sdk.NewCoins(sdk.NewInt64Coin("steak", 100))
	s.T().Log("verify if expired authorization is rejected")
	x := &banktypes.SendAuthorization{SpendLimit: newCoins}
	err := app.AuthzKeeper.SaveGrant(ctx, granterAddr, granteeAddr, x, now.Add(-1*time.Hour))
	s.Require().Error(err)
	authorization, _ = app.AuthzKeeper.GetCleanAuthorization(ctx, granteeAddr, granterAddr, bankSendAuthMsgType)
	s.Require().Nil(authorization)

	s.T().Log("verify if authorization is accepted")
	x = &banktypes.SendAuthorization{SpendLimit: newCoins}
	err = app.AuthzKeeper.SaveGrant(ctx, granteeAddr, granterAddr, x, now.Add(time.Hour))
	s.Require().NoError(err)
	authorization, _ = app.AuthzKeeper.GetCleanAuthorization(ctx, granteeAddr, granterAddr, bankSendAuthMsgType)
	s.Require().NotNil(authorization)
	s.Require().Equal(authorization.MsgTypeURL(), bankSendAuthMsgType)

	s.T().Log("verify fetching authorization with wrong msg type fails")
	authorization, _ = app.AuthzKeeper.GetCleanAuthorization(ctx, granteeAddr, granterAddr, sdk.MsgTypeURL(&banktypes.MsgMultiSend{}))
	s.Require().Nil(authorization)

	s.T().Log("verify fetching authorization with wrong grantee fails")
	authorization, _ = app.AuthzKeeper.GetCleanAuthorization(ctx, recipientAddr, granterAddr, bankSendAuthMsgType)
	s.Require().Nil(authorization)

	s.T().Log("verify revoke fails with wrong information")
	err = app.AuthzKeeper.DeleteGrant(ctx, recipientAddr, granterAddr, bankSendAuthMsgType)
	s.Require().Error(err)
	authorization, _ = app.AuthzKeeper.GetCleanAuthorization(ctx, recipientAddr, granterAddr, bankSendAuthMsgType)
	s.Require().Nil(authorization)

	s.T().Log("verify revoke executes with correct information")
	err = app.AuthzKeeper.DeleteGrant(ctx, granteeAddr, granterAddr, bankSendAuthMsgType)
	s.Require().NoError(err)
	authorization, _ = app.AuthzKeeper.GetCleanAuthorization(ctx, granteeAddr, granterAddr, bankSendAuthMsgType)
	s.Require().Nil(authorization)

}

func (s *TestSuite) TestKeeperIter() {
	app, ctx, addrs := s.app, s.ctx, s.addrs

	granterAddr := addrs[0]
	granteeAddr := addrs[1]

	s.T().Log("verify that no authorization returns nil")
	authorization, expiration := app.AuthzKeeper.GetCleanAuthorization(ctx, granteeAddr, granterAddr, "Abcd")
	s.Require().Nil(authorization)
	s.Require().Equal(time.Time{}, expiration)
	now := s.ctx.BlockHeader().Time.Add(time.Second)

	newCoins := sdk.NewCoins(sdk.NewInt64Coin("steak", 100))
	s.T().Log("verify if expired authorization is rejected")
	x := &banktypes.SendAuthorization{SpendLimit: newCoins}
	err := app.AuthzKeeper.SaveGrant(ctx, granteeAddr, granterAddr, x, now.Add(-1*time.Hour))
	s.Require().Error(err)
	authorization, _ = app.AuthzKeeper.GetCleanAuthorization(ctx, granteeAddr, granterAddr, "abcd")
	s.Require().Nil(authorization)

	app.AuthzKeeper.IterateGrants(ctx, func(granter, grantee sdk.AccAddress, grant authz.Grant) bool {
		s.Require().Equal(granter, granterAddr)
		s.Require().Equal(grantee, granteeAddr)
		return true
	})

}

func (s *TestSuite) TestKeeperFees() {
	app, addrs := s.app, s.addrs

	granterAddr := addrs[0]
	granteeAddr := addrs[1]
	recipientAddr := addrs[2]
	s.Require().NoError(testutil.FundAccount(app.BankKeeper, s.ctx, granterAddr, sdk.NewCoins(sdk.NewInt64Coin("steak", 10000))))
	expiration := s.ctx.BlockHeader().Time.Add(1 * time.Second)

	smallCoin := sdk.NewCoins(sdk.NewInt64Coin("steak", 20))
	someCoin := sdk.NewCoins(sdk.NewInt64Coin("steak", 123))

	msgs := authz.NewMsgExec(granteeAddr, []sdk.Msg{
		&banktypes.MsgSend{
			Amount:      sdk.NewCoins(sdk.NewInt64Coin("steak", 2)),
			FromAddress: granterAddr.String(),
			ToAddress:   recipientAddr.String(),
		},
	})

	s.Require().NoError(msgs.UnpackInterfaces(app.AppCodec()))

	s.T().Log("verify dispatch fails with invalid authorization")
	executeMsgs, err := msgs.GetMessages()
	s.Require().NoError(err)
	result, err := app.AuthzKeeper.DispatchActions(s.ctx, granteeAddr, executeMsgs)

	s.Require().Nil(result)
	s.Require().NotNil(err)

	s.T().Log("verify dispatch executes with correct information")
	// grant authorization
	err = app.AuthzKeeper.SaveGrant(s.ctx, granteeAddr, granterAddr, &banktypes.SendAuthorization{SpendLimit: smallCoin}, expiration)
	s.Require().NoError(err)
	authorization, _ := app.AuthzKeeper.GetCleanAuthorization(s.ctx, granteeAddr, granterAddr, bankSendAuthMsgType)
	s.Require().NotNil(authorization)

	s.Require().Equal(authorization.MsgTypeURL(), bankSendAuthMsgType)

	executeMsgs, err = msgs.GetMessages()
	s.Require().NoError(err)

	result, err = app.AuthzKeeper.DispatchActions(s.ctx, granteeAddr, executeMsgs)
	s.Require().NoError(err)
	s.Require().NotNil(result)

	authorization, _ = app.AuthzKeeper.GetCleanAuthorization(s.ctx, granteeAddr, granterAddr, bankSendAuthMsgType)
	s.Require().NotNil(authorization)

	s.T().Log("verify dispatch fails with overlimit")
	// grant authorization

	msgs = authz.NewMsgExec(granteeAddr, []sdk.Msg{
		&banktypes.MsgSend{
			Amount:      someCoin,
			FromAddress: granterAddr.String(),
			ToAddress:   recipientAddr.String(),
		},
	})

	s.Require().NoError(msgs.UnpackInterfaces(app.AppCodec()))
	executeMsgs, err = msgs.GetMessages()
	s.Require().NoError(err)

	result, err = app.AuthzKeeper.DispatchActions(s.ctx, granteeAddr, executeMsgs)
	s.Require().Nil(result)
	s.Require().NotNil(err)

	authorization, _ = app.AuthzKeeper.GetCleanAuthorization(s.ctx, granteeAddr, granterAddr, bankSendAuthMsgType)
	s.Require().NotNil(authorization)
}

// Tests that all msg events included in an authz MsgExec tx
// Ref: https://github.com/cosmos/cosmos-sdk/issues/9501
func (s *TestSuite) TestDispatchedEvents() {
	require := s.Require()
	app, addrs := s.app, s.addrs
	granterAddr := addrs[0]
	granteeAddr := addrs[1]
	recipientAddr := addrs[2]
	require.NoError(testutil.FundAccount(app.BankKeeper, s.ctx, granterAddr, sdk.NewCoins(sdk.NewInt64Coin("steak", 10000))))
	expiration := s.ctx.BlockHeader().Time.Add(1 * time.Second) // must be in the future

	smallCoin := sdk.NewCoins(sdk.NewInt64Coin("steak", 20))
	msgs := authz.NewMsgExec(granteeAddr, []sdk.Msg{
		&banktypes.MsgSend{
			Amount:      sdk.NewCoins(sdk.NewInt64Coin("steak", 2)),
			FromAddress: granterAddr.String(),
			ToAddress:   recipientAddr.String(),
		},
	})

	// grant authorization
	err := app.AuthzKeeper.SaveGrant(s.ctx, granteeAddr, granterAddr, &banktypes.SendAuthorization{SpendLimit: smallCoin}, expiration)
	require.NoError(err)
	authorization, _ := app.AuthzKeeper.GetCleanAuthorization(s.ctx, granteeAddr, granterAddr, bankSendAuthMsgType)
	require.NotNil(authorization)
	require.Equal(authorization.MsgTypeURL(), bankSendAuthMsgType)

	executeMsgs, err := msgs.GetMessages()
	require.NoError(err)

	result, err := app.AuthzKeeper.DispatchActions(s.ctx, granteeAddr, executeMsgs)
	require.NoError(err)
	require.NotNil(result)
	events := s.ctx.EventManager().Events()
	// get last 5 events (events that occur *after* the grant)
	events = events[len(events)-5:]
	requiredEvents := map[string]bool{
		"coin_spent":    false,
		"coin_received": false,
		"transfer":      false,
		"message":       false,
	}
	for _, e := range events {
		requiredEvents[e.Type] = true
	}
	for _, v := range requiredEvents {
		require.True(v)
	}
}

func TestTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
