package keeper_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant/keeper"
	"github.com/cosmos/cosmos-sdk/x/feegrant/types"
)

type KeeperTestSuite struct {
	suite.Suite

	app     *simapp.SimApp
	sdkCtx  sdk.Context
	addrs   []sdk.AccAddress
	msgSrvr types.MsgServer
	ctx     context.Context
	atom    sdk.Coins
	keeper  keeper.Keeper
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) SetupTest() {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	suite.app = app
	suite.sdkCtx = ctx
	suite.addrs = simapp.AddTestAddrsIncremental(app, ctx, 4, sdk.NewInt(30000000))
	suite.ctx = sdk.WrapSDKContext(ctx)
	suite.keeper = suite.app.FeeGrantKeeper
	suite.msgSrvr = keeper.NewMsgServerImpl(suite.keeper)
	suite.atom = sdk.NewCoins(sdk.NewCoin("atom", sdk.NewInt(555)))
}

func (suite *KeeperTestSuite) TestKeeperCrud() {
	// some helpers
	eth := sdk.NewCoins(sdk.NewInt64Coin("eth", 123))
	basic := &types.BasicAllowance{
		SpendLimit: suite.atom,
		Expiration: types.ExpiresAtHeight(334455),
	}

	basic2 := &types.BasicAllowance{
		SpendLimit: eth,
		Expiration: types.ExpiresAtHeight(172436),
	}

	// let's set up some initial state here
	err := suite.keeper.GrantAllowance(suite.sdkCtx, suite.addrs[0], suite.addrs[1], basic)
	suite.Require().NoError(err)

	err = suite.keeper.GrantAllowance(suite.sdkCtx, suite.addrs[0], suite.addrs[2], basic2)
	suite.Require().NoError(err)

	err = suite.keeper.GrantAllowance(suite.sdkCtx, suite.addrs[1], suite.addrs[2], basic)
	suite.Require().NoError(err)

	err = suite.keeper.GrantAllowance(suite.sdkCtx, suite.addrs[1], suite.addrs[3], basic)
	suite.Require().NoError(err)

	err = suite.keeper.GrantAllowance(suite.sdkCtx, suite.addrs[3], suite.addrs[0], basic2)
	suite.Require().NoError(err)

	// remove some, overwrite other
	_, err = suite.msgSrvr.RevokeAllowance(suite.ctx, &types.MsgRevokeAllowance{Granter: suite.addrs[0].String(), Grantee: suite.addrs[1].String()})
	suite.Require().NoError(err)
	_, err = suite.msgSrvr.RevokeAllowance(suite.ctx, &types.MsgRevokeAllowance{Granter: suite.addrs[0].String(), Grantee: suite.addrs[2].String()})
	suite.Require().NoError(err)

	// revoke non-exist fee allowance
	_, err = suite.msgSrvr.RevokeAllowance(suite.ctx, &types.MsgRevokeAllowance{Granter: suite.addrs[0].String(), Grantee: suite.addrs[2].String()})
	suite.Require().Error(err)

	err = suite.keeper.GrantAllowance(suite.sdkCtx, suite.addrs[0], suite.addrs[2], basic)
	suite.Require().NoError(err)

	err = suite.keeper.GrantAllowance(suite.sdkCtx, suite.addrs[1], suite.addrs[2], basic2)
	suite.Require().NoError(err)

	// end state:
	// addr -> addr3 (basic)
	// addr2 -> addr3 (basic2), addr4(basic)
	// addr4 -> addr (basic2)

	// then lots of queries
	cases := map[string]struct {
		grantee   sdk.AccAddress
		granter   sdk.AccAddress
		allowance types.FeeAllowanceI
	}{
		"addr revoked": {
			granter: suite.addrs[0],
			grantee: suite.addrs[1],
		},
		"addr revoked and added": {
			granter:   suite.addrs[0],
			grantee:   suite.addrs[2],
			allowance: basic,
		},
		"addr never there": {
			granter: suite.addrs[0],
			grantee: suite.addrs[3],
		},
		"addr modified": {
			granter:   suite.addrs[1],
			grantee:   suite.addrs[2],
			allowance: basic2,
		},
	}

	for name, tc := range cases {
		tc := tc
		suite.Run(name, func() {
			allow, _ := suite.keeper.GetAllowance(suite.sdkCtx, tc.granter, tc.grantee)

			if tc.allowance == nil {
				suite.Nil(allow)
				return
			}
			suite.NotNil(allow)
			suite.Equal(tc.allowance, allow)
		})
	}
	accAddr, err := sdk.AccAddressFromBech32("cosmos1rxr4mq58w3gtnx5tsc438mwjjafv3mja7k5pnu")
	suite.Require().NoError(err)

	// let's grant and revoke authorization to non existing account
	err = suite.keeper.GrantAllowance(suite.sdkCtx, suite.addrs[3], accAddr, basic2)
	suite.Require().NoError(err)

	_, err = suite.keeper.GetAllowance(suite.sdkCtx, suite.addrs[3], accAddr)
	suite.Require().NoError(err)

	_, err = suite.msgSrvr.RevokeAllowance(suite.ctx, &types.MsgRevokeAllowance{Granter: suite.addrs[3].String(), Grantee: accAddr.String()})
	suite.Require().NoError(err)

}

func (suite *KeeperTestSuite) TestUseGrantedFee() {
	eth := sdk.NewCoins(sdk.NewInt64Coin("eth", 123))
	future := &types.BasicAllowance{
		SpendLimit: suite.atom,
		Expiration: types.ExpiresAtHeight(5678),
	}

	expired := &types.BasicAllowance{
		SpendLimit: eth,
		Expiration: types.ExpiresAtHeight(55),
	}

	// for testing limits of the contract
	hugeAtom := sdk.NewCoins(sdk.NewInt64Coin("atom", 9999))
	_ = hugeAtom
	smallAtom := sdk.NewCoins(sdk.NewInt64Coin("atom", 1))
	_ = smallAtom
	futureAfterSmall := &types.BasicAllowance{
		SpendLimit: sdk.NewCoins(sdk.NewInt64Coin("atom", 554)),
		Expiration: types.ExpiresAtHeight(5678),
	}

	// then lots of queries
	cases := map[string]struct {
		grantee sdk.AccAddress
		granter sdk.AccAddress
		fee     sdk.Coins
		allowed bool
		final   types.FeeAllowanceI
	}{
		"use entire pot": {
			granter: suite.addrs[0],
			grantee: suite.addrs[1],
			fee:     suite.atom,
			allowed: true,
			final:   nil,
		},
		"expired and removed": {
			granter: suite.addrs[0],
			grantee: suite.addrs[2],
			fee:     eth,
			allowed: false,
			final:   nil,
		},
		"too high": {
			granter: suite.addrs[0],
			grantee: suite.addrs[1],
			fee:     hugeAtom,
			allowed: false,
			final:   future,
		},
		"use a little": {
			granter: suite.addrs[0],
			grantee: suite.addrs[1],
			fee:     smallAtom,
			allowed: true,
			final:   futureAfterSmall,
		},
	}

	for name, tc := range cases {
		tc := tc
		suite.Run(name, func() {
			// let's set up some initial state here
			// addr -> addr2 (future)
			// addr -> addr3 (expired)

			err := suite.keeper.GrantAllowance(suite.sdkCtx, suite.addrs[0], suite.addrs[1], future)
			suite.Require().NoError(err)

			err = suite.keeper.GrantAllowance(suite.sdkCtx, suite.addrs[0], suite.addrs[3], expired)
			suite.Require().NoError(err)

			err = suite.keeper.UseGrantedFees(suite.sdkCtx, tc.granter, tc.grantee, tc.fee, []sdk.Msg{})
			if tc.allowed {
				suite.NoError(err)
			} else {
				suite.Error(err)
			}

			loaded, _ := suite.keeper.GetAllowance(suite.sdkCtx, tc.granter, tc.grantee)

			suite.Equal(tc.final, loaded)
		})
	}
}

func (suite *KeeperTestSuite) TestIterateGrants() {
	eth := sdk.NewCoins(sdk.NewInt64Coin("eth", 123))
	allowance := &types.BasicAllowance{
		SpendLimit: suite.atom,
		Expiration: types.ExpiresAtHeight(5678),
	}

	allowance1 := &types.BasicAllowance{
		SpendLimit: eth,
		Expiration: types.ExpiresAtTime(suite.sdkCtx.BlockTime().Add(24 * time.Hour)),
	}

	suite.keeper.GrantAllowance(suite.sdkCtx, suite.addrs[0], suite.addrs[1], allowance)
	suite.keeper.GrantAllowance(suite.sdkCtx, suite.addrs[2], suite.addrs[1], allowance1)

	suite.keeper.IterateAllFeeAllowances(suite.sdkCtx, func(grant types.Grant) bool {
		suite.Require().Equal(suite.addrs[1].String(), grant.Grantee)
		suite.Require().Contains([]string{suite.addrs[0].String(), suite.addrs[2].String()}, grant.Granter)
		return true
	})

}
