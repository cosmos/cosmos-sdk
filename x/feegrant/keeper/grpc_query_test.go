package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant/types"
)

func (suite *KeeperTestSuite) TestFeeAllowance() {

	testCases := []struct {
		name      string
		req       *types.QueryFeeAllowanceRequest
		expectErr bool
		preRun    func()
		postRun   func(_ *types.QueryFeeAllowanceResponse)
	}{
		{
			"nil request",
			nil,
			true,
			func() {},
			func(*types.QueryFeeAllowanceResponse) {},
		},
		{
			"fail: invalid granter",
			&types.QueryFeeAllowanceRequest{
				Granter: "invalid_granter",
				Grantee: suite.addrs[0].String(),
			},
			true,
			func() {},
			func(*types.QueryFeeAllowanceResponse) {},
		},
		{
			"fail: invalid grantee",
			&types.QueryFeeAllowanceRequest{
				Granter: suite.addrs[0].String(),
				Grantee: "invalid_grantee",
			},
			true,
			func() {},
			func(*types.QueryFeeAllowanceResponse) {},
		},
		{
			"fail: no grants",
			&types.QueryFeeAllowanceRequest{
				Granter: suite.addrs[0].String(),
				Grantee: suite.addrs[1].String(),
			},
			true,
			func() {},
			func(*types.QueryFeeAllowanceResponse) {},
		},
		{
			"valid query: expect single grant",
			&types.QueryFeeAllowanceRequest{
				Granter: suite.addrs[0].String(),
				Grantee: suite.addrs[1].String(),
			},
			false,
			func() {
				grantFeeAllowance(suite)
			},
			func(allowance *types.QueryFeeAllowanceResponse) {
				suite.Require().Equal(allowance.FeeAllowance.Granter, suite.addrs[0].String())
				suite.Require().Equal(allowance.FeeAllowance.Grantee, suite.addrs[1].String())
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			tc.preRun()
			resp, err := suite.keeper.FeeAllowance(suite.ctx, tc.req)
			if tc.expectErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
				tc.postRun(resp)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestFeeAllowances() {
	testCases := []struct {
		name      string
		req       *types.QueryFeeAllowancesRequest
		expectErr bool
		preRun    func()
		postRun   func(_ *types.QueryFeeAllowancesResponse)
	}{
		{
			"nil request",
			nil,
			true,
			func() {},
			func(*types.QueryFeeAllowancesResponse) {},
		},
		{
			"fail: invalid grantee",
			&types.QueryFeeAllowancesRequest{
				Grantee: "invalid_grantee",
			},
			true,
			func() {},
			func(*types.QueryFeeAllowancesResponse) {},
		},
		{
			"no grants",
			&types.QueryFeeAllowancesRequest{
				Grantee: suite.addrs[1].String(),
			},
			false,
			func() {},
			func(resp *types.QueryFeeAllowancesResponse) {
				suite.Require().Equal(len(resp.FeeAllowances), 0)
			},
		},
		{
			"valid query: expect single grant",
			&types.QueryFeeAllowancesRequest{
				Grantee: suite.addrs[1].String(),
			},
			false,
			func() {
				grantFeeAllowance(suite)
			},
			func(resp *types.QueryFeeAllowancesResponse) {
				suite.Require().Equal(len(resp.FeeAllowances), 1)
				suite.Require().Equal(resp.FeeAllowances[0].Granter, suite.addrs[0].String())
				suite.Require().Equal(resp.FeeAllowances[0].Grantee, suite.addrs[1].String())
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			tc.preRun()
			resp, err := suite.keeper.FeeAllowances(suite.ctx, tc.req)
			if tc.expectErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
				tc.postRun(resp)
			}
		})
	}
}

func grantFeeAllowance(suite *KeeperTestSuite) {
	err := suite.app.FeeGrantKeeper.GrantFeeAllowance(suite.sdkCtx, suite.addrs[0], suite.addrs[1], &types.BasicFeeAllowance{
		SpendLimit: sdk.NewCoins(sdk.NewInt64Coin("atom", 555)),
		Expiration: types.ExpiresAtHeight(334455),
	})
	suite.Require().NoError(err)
}
