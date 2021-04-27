package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant/types"
)

type KeeperTestSuite struct {
	suite.Suite

	app   *simapp.SimApp
	ctx   sdk.Context
	addrs []sdk.AccAddress
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) SetupTest() {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	suite.app = app
	suite.ctx = ctx
	suite.addrs = simapp.AddTestAddrsIncremental(app, ctx, 4, sdk.NewInt(30000000))
}

func (suite *KeeperTestSuite) TestKeeperCrud() {
	ctx := suite.ctx
	k := suite.app.FeeGrantKeeper

	// some helpers
	atom := sdk.NewCoins(sdk.NewInt64Coin("atom", 555))
	eth := sdk.NewCoins(sdk.NewInt64Coin("eth", 123))
	exp := ctx.BlockTime().AddDate(1, 0, 0)
	basic := &types.BasicFeeAllowance{
		SpendLimit: atom,
		Expiration: &exp,
	}

	basic2 := &types.BasicFeeAllowance{
		SpendLimit: eth,
		Expiration: &exp,
	}

	// let's set up some initial state here
	err := k.GrantFeeAllowance(ctx, suite.addrs[0], suite.addrs[1], basic)
	suite.Require().NoError(err)

	err = k.GrantFeeAllowance(ctx, suite.addrs[0], suite.addrs[2], basic2)
	suite.Require().NoError(err)

	err = k.GrantFeeAllowance(ctx, suite.addrs[1], suite.addrs[2], basic)
	suite.Require().NoError(err)

	err = k.GrantFeeAllowance(ctx, suite.addrs[1], suite.addrs[3], basic)
	suite.Require().NoError(err)

	err = k.GrantFeeAllowance(ctx, suite.addrs[3], suite.addrs[0], basic2)
	suite.Require().NoError(err)

	// remove some, overwrite other
	k.RevokeFeeAllowance(ctx, suite.addrs[0], suite.addrs[1])
	k.RevokeFeeAllowance(ctx, suite.addrs[0], suite.addrs[2])

	err = k.GrantFeeAllowance(ctx, suite.addrs[0], suite.addrs[2], basic)
	suite.Require().NoError(err)

	err = k.GrantFeeAllowance(ctx, suite.addrs[1], suite.addrs[2], basic2)
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
			allow, _ := k.GetFeeAllowance(ctx, tc.granter, tc.grantee)

			if tc.allowance == nil {
				suite.Nil(allow)
				return
			}
			suite.NotNil(allow)
			suite.Equal(tc.allowance, allow)
		})
	}

	grant1, err := types.NewFeeAllowanceGrant(suite.addrs[3], suite.addrs[0], basic2)
	suite.NoError(err)

	grant2, err := types.NewFeeAllowanceGrant(suite.addrs[1], suite.addrs[2], basic2)
	suite.NoError(err)

	grant3, err := types.NewFeeAllowanceGrant(suite.addrs[0], suite.addrs[2], basic)
	suite.NoError(err)

	allCases := map[string]struct {
		grantee sdk.AccAddress
		grants  []types.FeeAllowanceGrant
	}{
		"addr2 has none": {
			grantee: suite.addrs[1],
		},
		"addr has one": {
			grantee: suite.addrs[0],
			grants: []types.FeeAllowanceGrant{
				grant1,
			},
		},
		"addr3 has two": {
			grantee: suite.addrs[2],
			grants: []types.FeeAllowanceGrant{
				grant3,
				grant2,
			},
		},
	}

	for name, tc := range allCases {
		tc := tc
		suite.Run(name, func() {
			var grants []types.FeeAllowanceGrant
			err := k.IterateAllGranteeFeeAllowances(ctx, tc.grantee, func(grant types.FeeAllowanceGrant) bool {
				grants = append(grants, grant)
				return false
			})
			suite.NoError(err)
			suite.Equal(tc.grants, grants)
		})
	}
}

func (suite *KeeperTestSuite) TestUseGrantedFee() {
	ctx := suite.ctx
	k := suite.app.FeeGrantKeeper

	// some helpers
	atom := sdk.NewCoins(sdk.NewInt64Coin("atom", 555))
	eth := sdk.NewCoins(sdk.NewInt64Coin("eth", 123))

	exp := ctx.BlockTime().AddDate(1, 0, 0)
	future := &types.BasicFeeAllowance{
		SpendLimit: atom,
		Expiration: &exp,
	}

	expired := &types.BasicFeeAllowance{
		SpendLimit: eth,
		Expiration: &exp,
	}

	// for testing limits of the contract
	hugeAtom := sdk.NewCoins(sdk.NewInt64Coin("atom", 9999))
	_ = hugeAtom
	smallAtom := sdk.NewCoins(sdk.NewInt64Coin("atom", 1))
	_ = smallAtom
	futureAfterSmall := &types.BasicFeeAllowance{
		SpendLimit: sdk.NewCoins(sdk.NewInt64Coin("atom", 554)),
		Expiration: &exp,
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
			fee:     atom,
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

			err := k.GrantFeeAllowance(ctx, suite.addrs[0], suite.addrs[1], future)
			suite.Require().NoError(err)

			err = k.GrantFeeAllowance(ctx, suite.addrs[0], suite.addrs[3], expired)
			suite.Require().NoError(err)

			err = k.UseGrantedFees(ctx, tc.granter, tc.grantee, tc.fee, []sdk.Msg{})
			if tc.allowed {
				suite.NoError(err)
			} else {
				suite.Error(err)
			}

			loaded, _ := k.GetFeeAllowance(ctx, tc.granter, tc.grantee)

			suite.Equal(tc.final, loaded)
		})
	}
}
