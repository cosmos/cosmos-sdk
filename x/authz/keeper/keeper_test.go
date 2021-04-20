package keeper_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

type TestSuite struct {
	suite.Suite

	app         *simapp.SimApp
	ctx         sdk.Context
	addrs       []sdk.AccAddress
	queryClient types.QueryClient
}

func (s *TestSuite) SetupTest() {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	now := tmtime.Now()
	ctx = ctx.WithBlockHeader(tmproto.Header{Time: now})
	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, app.AuthzKeeper)
	queryClient := types.NewQueryClient(queryHelper)
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
	authorization, expiration := app.AuthzKeeper.GetOrRevokeAuthorization(ctx, granteeAddr, granterAddr, banktypes.SendAuthorization{}.MethodName())
	s.Require().Nil(authorization)
	s.Require().Equal(expiration, time.Time{})
	now := s.ctx.BlockHeader().Time
	s.Require().NotNil(now)

	newCoins := sdk.NewCoins(sdk.NewInt64Coin("steak", 100))
	s.T().Log("verify if expired authorization is rejected")
	x := &banktypes.SendAuthorization{SpendLimit: newCoins}
	err := app.AuthzKeeper.Grant(ctx, granterAddr, granteeAddr, x, now.Add(-1*time.Hour))
	s.Require().NoError(err)
	authorization, _ = app.AuthzKeeper.GetOrRevokeAuthorization(ctx, granteeAddr, granterAddr, banktypes.SendAuthorization{}.MethodName())
	s.Require().Nil(authorization)

	s.T().Log("verify if authorization is accepted")
	x = &banktypes.SendAuthorization{SpendLimit: newCoins}
	err = app.AuthzKeeper.Grant(ctx, granteeAddr, granterAddr, x, now.Add(time.Hour))
	s.Require().NoError(err)
	authorization, _ = app.AuthzKeeper.GetOrRevokeAuthorization(ctx, granteeAddr, granterAddr, banktypes.SendAuthorization{}.MethodName())
	s.Require().NotNil(authorization)
	s.Require().Equal(authorization.MethodName(), banktypes.SendAuthorization{}.MethodName())

	s.T().Log("verify fetching authorization with wrong msg type fails")
	authorization, _ = app.AuthzKeeper.GetOrRevokeAuthorization(ctx, granteeAddr, granterAddr, sdk.MsgName(&banktypes.MsgMultiSend{}))
	s.Require().Nil(authorization)

	s.T().Log("verify fetching authorization with wrong grantee fails")
	authorization, _ = app.AuthzKeeper.GetOrRevokeAuthorization(ctx, recipientAddr, granterAddr, banktypes.SendAuthorization{}.MethodName())
	s.Require().Nil(authorization)

	s.T().Log("verify revoke fails with wrong information")
	err = app.AuthzKeeper.Revoke(ctx, recipientAddr, granterAddr, banktypes.SendAuthorization{}.MethodName())
	s.Require().Error(err)
	authorization, _ = app.AuthzKeeper.GetOrRevokeAuthorization(ctx, recipientAddr, granterAddr, banktypes.SendAuthorization{}.MethodName())
	s.Require().Nil(authorization)

	s.T().Log("verify revoke executes with correct information")
	err = app.AuthzKeeper.Revoke(ctx, granteeAddr, granterAddr, banktypes.SendAuthorization{}.MethodName())
	s.Require().NoError(err)
	authorization, _ = app.AuthzKeeper.GetOrRevokeAuthorization(ctx, granteeAddr, granterAddr, banktypes.SendAuthorization{}.MethodName())
	s.Require().Nil(authorization)

}

func (s *TestSuite) TestKeeperIter() {
	app, ctx, addrs := s.app, s.ctx, s.addrs

	granterAddr := addrs[0]
	granteeAddr := addrs[1]

	s.T().Log("verify that no authorization returns nil")
	authorization, expiration := app.AuthzKeeper.GetOrRevokeAuthorization(ctx, granteeAddr, granterAddr, "Abcd")
	s.Require().Nil(authorization)
	s.Require().Equal(time.Time{}, expiration)
	now := s.ctx.BlockHeader().Time
	s.Require().NotNil(now)

	newCoins := sdk.NewCoins(sdk.NewInt64Coin("steak", 100))
	s.T().Log("verify if expired authorization is rejected")
	x := &banktypes.SendAuthorization{SpendLimit: newCoins}
	err := app.AuthzKeeper.Grant(ctx, granteeAddr, granterAddr, x, now.Add(-1*time.Hour))
	s.Require().NoError(err)
	authorization, _ = app.AuthzKeeper.GetOrRevokeAuthorization(ctx, granteeAddr, granterAddr, "abcd")
	s.Require().Nil(authorization)

	app.AuthzKeeper.IterateGrants(ctx, func(granter, grantee sdk.AccAddress, grant types.AuthorizationGrant) bool {
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
	s.Require().NoError(simapp.FundAccount(app, s.ctx, granterAddr, sdk.NewCoins(sdk.NewInt64Coin("steak", 10000))))
	now := s.ctx.BlockHeader().Time
	s.Require().NotNil(now)

	smallCoin := sdk.NewCoins(sdk.NewInt64Coin("steak", 20))
	someCoin := sdk.NewCoins(sdk.NewInt64Coin("steak", 123))

	msgs := types.NewMsgExecAuthorized(granteeAddr, []sdk.Msg{
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
	err = app.AuthzKeeper.Grant(s.ctx, granteeAddr, granterAddr, &banktypes.SendAuthorization{SpendLimit: smallCoin}, now)
	s.Require().NoError(err)
	authorization, _ := app.AuthzKeeper.GetOrRevokeAuthorization(s.ctx, granteeAddr, granterAddr, banktypes.SendAuthorization{}.MethodName())
	s.Require().NotNil(authorization)

	s.Require().Equal(authorization.MethodName(), banktypes.SendAuthorization{}.MethodName())

	executeMsgs, err = msgs.GetMessages()
	s.Require().NoError(err)

	fmt.Println("TEST granteeAddr=", granteeAddr, "executeMsgs=", executeMsgs)
	result, err = app.AuthzKeeper.DispatchActions(s.ctx, granteeAddr, executeMsgs)
	s.Require().NoError(err)
	s.Require().NotNil(result)

	authorization, _ = app.AuthzKeeper.GetOrRevokeAuthorization(s.ctx, granteeAddr, granterAddr, banktypes.SendAuthorization{}.MethodName())
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

	s.Require().NoError(msgs.UnpackInterfaces(app.AppCodec()))
	executeMsgs, err = msgs.GetMessages()
	s.Require().NoError(err)

	result, err = app.AuthzKeeper.DispatchActions(s.ctx, granteeAddr, executeMsgs)
	s.Require().Nil(result)
	s.Require().NotNil(err)

	authorization, _ = app.AuthzKeeper.GetOrRevokeAuthorization(s.ctx, granteeAddr, granterAddr, banktypes.SendAuthorization{}.MethodName())
	s.Require().NotNil(authorization)
}

func TestTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
