package keeper_test

import (
	"testing"

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

var (
	bankSendAuthMsgType = banktypes.SendAuthorization{}.MsgTypeURL()
	coins10             = sdk.NewCoins(sdk.NewInt64Coin("stake", 10))
	coins100            = sdk.NewCoins(sdk.NewInt64Coin("stake", 100))
	coins1000           = sdk.NewCoins(sdk.NewInt64Coin("stake", 1000))
)

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
	now := ctx.BlockTime()
	require := s.Require()

	granterAddr := addrs[0]
	granteeAddr := addrs[1]

	s.T().Log("verify that no authorization returns nil")
	authorizations := app.AuthzKeeper.GetAuthorizations(ctx, granteeAddr, granterAddr)
	require.Len(authorizations, 0)

	s.T().Log("verify save, get and delete")
	sendAutz := &banktypes.SendAuthorization{SpendLimit: coins100}
	err := app.AuthzKeeper.SaveGrant(ctx, granteeAddr, granterAddr, sendAutz, now.AddDate(1, 0, 0))
	require.NoError(err)
	authorizations = app.AuthzKeeper.GetAuthorizations(ctx, granteeAddr, granterAddr)
	require.Len(authorizations, 1)

	err = app.AuthzKeeper.DeleteGrant(ctx, granteeAddr, granterAddr, sendAutz.MsgTypeURL())
	require.NoError(err)

	authorizations = app.AuthzKeeper.GetAuthorizations(ctx, granteeAddr, granterAddr)
	require.Len(authorizations, 0)

	s.T().Log("verify granting same authorization overwrite existing authorization")
	err = app.AuthzKeeper.SaveGrant(ctx, granteeAddr, granterAddr, sendAutz, now.AddDate(1, 0, 0))
	require.NoError(err)
	authorizations = app.AuthzKeeper.GetAuthorizations(ctx, granteeAddr, granterAddr)
	require.Len(authorizations, 1)

	sendAutz = &banktypes.SendAuthorization{SpendLimit: coins1000}
	err = app.AuthzKeeper.SaveGrant(ctx, granteeAddr, granterAddr, sendAutz, now.AddDate(1, 0, 0))
	require.NoError(err)
	authorizations = app.AuthzKeeper.GetAuthorizations(ctx, granteeAddr, granterAddr)
	require.Len(authorizations, 1)
	authorization := authorizations[0]
	sendAuth := authorization.(*banktypes.SendAuthorization)
	require.Equal(sendAuth.SpendLimit, sendAutz.SpendLimit)
	require.Equal(sendAuth.MsgTypeURL(), sendAutz.MsgTypeURL())

	s.T().Log("verify removing non existing authorization returns error")
	err = app.AuthzKeeper.DeleteGrant(ctx, granterAddr, granteeAddr, "abcd")
	s.Require().Error(err)
}

func (s *TestSuite) TestKeeperIter() {
	app, ctx, addrs := s.app, s.ctx, s.addrs

	granterAddr := addrs[0]
	granteeAddr := addrs[1]
	granter2Addr := addrs[2]

	s.app.AuthzKeeper.SaveGrant(ctx, granteeAddr, granterAddr, banktypes.NewSendAuthorization(coins100), ctx.BlockTime().AddDate(1, 0, 0))
	s.app.AuthzKeeper.SaveGrant(ctx, granteeAddr, granter2Addr, banktypes.NewSendAuthorization(coins100), ctx.BlockTime().AddDate(1, 0, 0))

	app.AuthzKeeper.IterateGrants(ctx, func(granter, grantee sdk.AccAddress, grant authz.Grant) bool {
		s.Require().Equal(granteeAddr, grantee)
		s.Require().Contains([]sdk.AccAddress{granterAddr, granter2Addr}, granter)
		return true
	})

}

func (s *TestSuite) TestDispatchAction() {
	app, addrs := s.app, s.addrs
	require := s.Require()
	now := s.ctx.BlockTime()

	granterAddr := addrs[0]
	granteeAddr := addrs[1]
	recipientAddr := addrs[2]

	require.NoError(testutil.FundAccount(app.BankKeeper, s.ctx, granterAddr, coins1000))

	testCases := []struct {
		name      string
		req       authz.MsgExec
		expectErr bool
		errMsg    string
		preRun    func()
		postRun   func()
	}{
		{
			"expect error authorization not found",
			authz.NewMsgExec(granteeAddr, []sdk.Msg{
				&banktypes.MsgSend{
					Amount:      coins10,
					FromAddress: granterAddr.String(),
					ToAddress:   recipientAddr.String(),
				},
			}),
			true,
			"authorization not found",
			func() {
				// remove any existing authorizations
				app.AuthzKeeper.DeleteGrant(s.ctx, granteeAddr, granterAddr, bankSendAuthMsgType)
			},
			func() {},
		},
		{
			"expect error expired authorization",
			authz.NewMsgExec(granteeAddr, []sdk.Msg{
				&banktypes.MsgSend{
					Amount:      coins10,
					FromAddress: granterAddr.String(),
					ToAddress:   recipientAddr.String(),
				},
			}),
			true,
			"authorization expired",
			func() {
				err := app.AuthzKeeper.SaveGrant(s.ctx, granteeAddr, granterAddr, banktypes.NewSendAuthorization(coins100), now.AddDate(0, -1, 0))
				require.NoError(err)
			},
			func() {},
		},
		{
			"expect error over spent limit",
			authz.NewMsgExec(granteeAddr, []sdk.Msg{
				&banktypes.MsgSend{
					Amount:      coins1000,
					FromAddress: granterAddr.String(),
					ToAddress:   recipientAddr.String(),
				},
			}),
			true,
			"requested amount is more than spend limit",
			func() {
				err := app.AuthzKeeper.SaveGrant(s.ctx, granteeAddr, granterAddr, banktypes.NewSendAuthorization(coins100), now.AddDate(0, 1, 0))
				require.NoError(err)
			},
			func() {},
		},
		{
			"valid test verify amount left",
			authz.NewMsgExec(granteeAddr, []sdk.Msg{
				&banktypes.MsgSend{
					Amount:      coins10,
					FromAddress: granterAddr.String(),
					ToAddress:   recipientAddr.String(),
				},
			}),
			false,
			"",
			func() {
				err := app.AuthzKeeper.SaveGrant(s.ctx, granteeAddr, granterAddr, banktypes.NewSendAuthorization(coins100), now.AddDate(0, 1, 0))
				require.NoError(err)
			},
			func() {
				authzs := app.AuthzKeeper.GetAuthorizations(s.ctx, granteeAddr, granterAddr)
				require.Len(authzs, 1)
				authorization := authzs[0].(*banktypes.SendAuthorization)
				require.NotNil(authorization)
				require.Equal(authorization.SpendLimit, coins100.Sub(coins10))
			},
		},
		{
			"valid test verify authorization is removed when it is used up",
			authz.NewMsgExec(granteeAddr, []sdk.Msg{
				&banktypes.MsgSend{
					Amount:      coins100,
					FromAddress: granterAddr.String(),
					ToAddress:   recipientAddr.String(),
				},
			}),
			false,
			"",
			func() {
				err := app.AuthzKeeper.SaveGrant(s.ctx, granteeAddr, granterAddr, banktypes.NewSendAuthorization(coins100), now.AddDate(0, 1, 0))
				require.NoError(err)
			},
			func() {
				authzs := app.AuthzKeeper.GetAuthorizations(s.ctx, granteeAddr, granterAddr)
				require.Len(authzs, 0)
			},
		},
	}

	for _, tc := range testCases {
		tc.preRun()
		executeMsgs, err := tc.req.GetMessages()
		s.Require().NoError(err)
		result, err := app.AuthzKeeper.DispatchActions(s.ctx, granteeAddr, executeMsgs)
		if tc.expectErr {
			require.Error(err)
			require.Nil(result)
			require.Contains(err.Error(), tc.errMsg)
		} else {
			require.NoError(err)
			require.NotNil(result)
		}
		tc.postRun()
	}

}

// Tests that all msg events included in an authz MsgExec tx
// Ref: https://github.com/cosmos/cosmos-sdk/issues/9501
func (s *TestSuite) TestDispatchedEvents() {
	require := s.Require()
	app, addrs := s.app, s.addrs
	granterAddr := addrs[0]
	granteeAddr := addrs[1]
	recipientAddr := addrs[2]
	require.NoError(testutil.FundAccount(app.BankKeeper, s.ctx, granterAddr, coins1000))
	now := s.ctx.BlockHeader().Time
	require.NotNil(now)

	msgs := authz.NewMsgExec(granteeAddr, []sdk.Msg{
		&banktypes.MsgSend{
			Amount:      coins10,
			FromAddress: granterAddr.String(),
			ToAddress:   recipientAddr.String(),
		},
	})

	// grant authorization
	err := app.AuthzKeeper.SaveGrant(s.ctx, granteeAddr, granterAddr, &banktypes.SendAuthorization{SpendLimit: coins10}, now)
	require.NoError(err)
	authorizations := app.AuthzKeeper.GetAuthorizations(s.ctx, granteeAddr, granterAddr)
	require.Len(authorizations, 1)
	authorization := authorizations[0].(*banktypes.SendAuthorization)
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
