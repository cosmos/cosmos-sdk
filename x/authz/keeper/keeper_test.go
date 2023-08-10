package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"

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
	s.addrs = simapp.AddTestAddrsIncremental(app, ctx, 7, sdk.NewInt(30000000))
}

func (s *TestSuite) TestKeeper() {
	app, ctx, addrs := s.app, s.ctx, s.addrs
	now := ctx.BlockTime()
	require := s.Require()

	granterAddr := addrs[0]
	granteeAddr := addrs[1]

	s.T().Log("verify that no authorization returns nil")
	authorizations, err := app.AuthzKeeper.GetAuthorizations(ctx, granteeAddr, granterAddr)
	require.NoError(err)
	require.Len(authorizations, 0)

	s.T().Log("verify save, get and delete")
	sendAutz := &banktypes.SendAuthorization{SpendLimit: coins100}
	expire := now.AddDate(1, 0, 0)
	err = app.AuthzKeeper.SaveGrant(ctx, granteeAddr, granterAddr, sendAutz, &expire)
	require.NoError(err)

	authorizations, err = app.AuthzKeeper.GetAuthorizations(ctx, granteeAddr, granterAddr)
	require.NoError(err)
	require.Len(authorizations, 1)

	err = app.AuthzKeeper.DeleteGrant(ctx, granteeAddr, granterAddr, sendAutz.MsgTypeURL())
	require.NoError(err)

	authorizations, err = app.AuthzKeeper.GetAuthorizations(ctx, granteeAddr, granterAddr)
	require.NoError(err)
	require.Len(authorizations, 0)

	s.T().Log("verify granting same authorization overwrite existing authorization")
	err = app.AuthzKeeper.SaveGrant(ctx, granteeAddr, granterAddr, sendAutz, &expire)
	require.NoError(err)

	authorizations, err = app.AuthzKeeper.GetAuthorizations(ctx, granteeAddr, granterAddr)
	require.NoError(err)
	require.Len(authorizations, 1)

	sendAutz = &banktypes.SendAuthorization{SpendLimit: coins1000}
	err = app.AuthzKeeper.SaveGrant(ctx, granteeAddr, granterAddr, sendAutz, &expire)
	require.NoError(err)
	authorizations, err = app.AuthzKeeper.GetAuthorizations(ctx, granteeAddr, granterAddr)
	require.NoError(err)
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
	e := ctx.BlockTime().AddDate(1, 0, 0)

	s.app.AuthzKeeper.SaveGrant(ctx, granteeAddr, granterAddr, banktypes.NewSendAuthorization(coins100), &e)
	s.app.AuthzKeeper.SaveGrant(ctx, granteeAddr, granter2Addr, banktypes.NewSendAuthorization(coins100), &e)

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
	a := banktypes.NewSendAuthorization(coins100)

	require.NoError(testutil.FundAccount(app.BankKeeper, s.ctx, granterAddr, coins1000))

	testCases := []struct {
		name      string
		req       authz.MsgExec
		expectErr bool
		errMsg    string
		preRun    func() sdk.Context
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
			func() sdk.Context {
				// remove any existing authorizations
				app.AuthzKeeper.DeleteGrant(s.ctx, granteeAddr, granterAddr, bankSendAuthMsgType)
				return s.ctx
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
			func() sdk.Context {
				e := now.AddDate(0, 0, 1)
				err := app.AuthzKeeper.SaveGrant(s.ctx, granteeAddr, granterAddr, a, &e)
				require.NoError(err)
				return s.ctx.WithBlockTime(s.ctx.BlockTime().AddDate(0, 0, 2))
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
			func() sdk.Context {
				e := now.AddDate(0, 1, 0)
				err := app.AuthzKeeper.SaveGrant(s.ctx, granteeAddr, granterAddr, a, &e)
				require.NoError(err)
				return s.ctx
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
			func() sdk.Context {
				e := now.AddDate(0, 1, 0)
				err := app.AuthzKeeper.SaveGrant(s.ctx, granteeAddr, granterAddr, a, &e)
				require.NoError(err)
				return s.ctx
			},
			func() {
				authzs, err := app.AuthzKeeper.GetAuthorizations(s.ctx, granteeAddr, granterAddr)
				require.NoError(err)
				require.Len(authzs, 1)
				authorization := authzs[0].(*banktypes.SendAuthorization)
				require.NotNil(authorization)
				require.Equal(authorization.SpendLimit, coins100.Sub(coins10...))
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
			func() sdk.Context {
				e := now.AddDate(0, 1, 0)
				err := app.AuthzKeeper.SaveGrant(s.ctx, granteeAddr, granterAddr, a, &e)
				require.NoError(err)
				return s.ctx
			},
			func() {
				authzs, err := app.AuthzKeeper.GetAuthorizations(s.ctx, granteeAddr, granterAddr)
				require.NoError(err)
				require.Len(authzs, 0)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			ctx := tc.preRun()
			executeMsgs, err := tc.req.GetMessages()
			require.NoError(err)
			result, err := app.AuthzKeeper.DispatchActions(ctx, granteeAddr, executeMsgs)
			if tc.expectErr {
				require.Error(err)
				require.Nil(result)
				require.Contains(err.Error(), tc.errMsg)
			} else {
				require.NoError(err)
				require.NotNil(result)
			}
			tc.postRun()
		})
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
	expiration := s.ctx.BlockTime().Add(1 * time.Second) // must be in the future

	msgs := authz.NewMsgExec(granteeAddr, []sdk.Msg{
		&banktypes.MsgSend{
			Amount:      coins10,
			FromAddress: granterAddr.String(),
			ToAddress:   recipientAddr.String(),
		},
	})

	// grant authorization
	err := app.AuthzKeeper.SaveGrant(s.ctx, granteeAddr, granterAddr, &banktypes.SendAuthorization{SpendLimit: coins10}, &expiration)
	require.NoError(err)
	authorizations, err := app.AuthzKeeper.GetAuthorizations(s.ctx, granteeAddr, granterAddr)
	require.NoError(err)
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

func (s *TestSuite) TestDequeueAllGrantsQueue() {
	require := s.Require()
	app, addrs := s.app, s.addrs
	granter := addrs[0]
	grantee := addrs[1]
	grantee1 := addrs[2]
	exp := s.ctx.BlockTime().AddDate(0, 0, 1)
	a := banktypes.SendAuthorization{SpendLimit: coins100}

	// create few authorizations
	err := app.AuthzKeeper.SaveGrant(s.ctx, grantee, granter, &a, &exp)
	require.NoError(err)

	err = app.AuthzKeeper.SaveGrant(s.ctx, grantee1, granter, &a, &exp)
	require.NoError(err)

	exp2 := exp.AddDate(0, 1, 0)
	err = app.AuthzKeeper.SaveGrant(s.ctx, granter, grantee1, &a, &exp2)
	require.NoError(err)

	exp2 = exp.AddDate(2, 0, 0)
	err = app.AuthzKeeper.SaveGrant(s.ctx, granter, grantee, &a, &exp2)
	require.NoError(err)

	newCtx := s.ctx.WithBlockTime(exp.AddDate(1, 0, 0))
	err = app.AuthzKeeper.DequeueAndDeleteExpiredGrants(newCtx)
	require.NoError(err)

	s.T().Log("verify expired grants are pruned from the state")
	authzs, err := app.AuthzKeeper.GetAuthorizations(newCtx, grantee, granter)
	require.NoError(err)
	require.Len(authzs, 0)

	authzs, err = app.AuthzKeeper.GetAuthorizations(newCtx, granter, grantee1)
	require.NoError(err)
	require.Len(authzs, 0)

	authzs, err = app.AuthzKeeper.GetAuthorizations(newCtx, grantee1, granter)
	require.NoError(err)
	require.Len(authzs, 0)

	authzs, err = app.AuthzKeeper.GetAuthorizations(newCtx, granter, grantee)
	require.NoError(err)
	require.Len(authzs, 1)
}

func (s *TestSuite) TestGetAuthorization() {
	addr1 := s.addrs[3]
	addr2 := s.addrs[4]
	addr3 := s.addrs[5]
	addr4 := s.addrs[6]

	genAuthMulti := authz.NewGenericAuthorization(sdk.MsgTypeURL(&banktypes.MsgMultiSend{}))
	genAuthSend := authz.NewGenericAuthorization(sdk.MsgTypeURL(&banktypes.MsgSend{}))
	sendAuth := banktypes.NewSendAuthorization(coins10)

	start := s.ctx.BlockHeader().Time
	expired := start.Add(time.Duration(1) * time.Second)
	notExpired := start.Add(time.Duration(5) * time.Hour)

	s.Require().NoError(s.app.AuthzKeeper.SaveGrant(s.ctx, addr1, addr2, genAuthMulti, nil), "creating grant 1->2")
	s.Require().NoError(s.app.AuthzKeeper.SaveGrant(s.ctx, addr1, addr3, genAuthSend, &expired), "creating grant 1->3")
	s.Require().NoError(s.app.AuthzKeeper.SaveGrant(s.ctx, addr1, addr4, sendAuth, &notExpired), "creating grant 1->4")
	// Without access to private keeper methods, I don't know how to save a grant with an invalid authorization.
	newCtx := s.ctx.WithBlockTime(start.Add(time.Duration(1) * time.Minute))

	tests := []struct {
		name    string
		grantee sdk.AccAddress
		granter sdk.AccAddress
		msgType string
		expAuth authz.Authorization
		expExp  *time.Time
	}{
		{
			name:    "grant has nil exp and is returned",
			grantee: addr1,
			granter: addr2,
			msgType: genAuthMulti.MsgTypeURL(),
			expAuth: genAuthMulti,
			expExp:  nil,
		},
		{
			name:    "grant is expired not returned",
			grantee: addr1,
			granter: addr3,
			msgType: genAuthSend.MsgTypeURL(),
			expAuth: nil,
			expExp:  nil,
		},
		{
			name:    "grant is not expired and is returned",
			grantee: addr1,
			granter: addr4,
			msgType: sendAuth.MsgTypeURL(),
			expAuth: sendAuth,
			expExp:  &notExpired,
		},
		{
			name:    "grant is not expired but wrong msg type returns nil",
			grantee: addr1,
			granter: addr4,
			msgType: genAuthMulti.MsgTypeURL(),
			expAuth: nil,
			expExp:  nil,
		},
		{
			name:    "no grant exists between the two",
			grantee: addr2,
			granter: addr3,
			msgType: genAuthSend.MsgTypeURL(),
			expAuth: nil,
			expExp:  nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			actAuth, actExp := s.app.AuthzKeeper.GetAuthorization(newCtx, tc.grantee, tc.granter, tc.msgType)
			s.Assert().Equal(tc.expAuth, actAuth, "authorization")
			s.Assert().Equal(tc.expExp, actExp, "expiration")
		})
	}
}

func (s *TestSuite) TestGetAuthorizations() {
	require := s.Require()
	addr1 := s.addrs[1]
	addr2 := s.addrs[2]

	genAuthMulti := authz.NewGenericAuthorization(sdk.MsgTypeURL(&banktypes.MsgMultiSend{}))
	genAuthSend := authz.NewGenericAuthorization(sdk.MsgTypeURL(&banktypes.MsgSend{}))

	start := s.ctx.BlockHeader().Time
	expired := start.Add(time.Duration(1) * time.Second)

	s.Require().NoError(s.app.AuthzKeeper.SaveGrant(s.ctx, addr1, addr2, genAuthMulti, &expired), "creating multi send grant 1->2")
	s.Require().NoError(s.app.AuthzKeeper.SaveGrant(s.ctx, addr1, addr2, genAuthSend, &expired), "creating send grant 1->2")

	authzs, err := s.app.AuthzKeeper.GetAuthorizations(s.ctx, addr1, addr2)
	require.NoError(err)
	require.Len(authzs, 2)
	require.Equal(sdk.MsgTypeURL(&banktypes.MsgMultiSend{}), authzs[0].MsgTypeURL())
	require.Equal(sdk.MsgTypeURL(&banktypes.MsgMultiSend{}), authzs[1].MsgTypeURL())
}

func TestTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
