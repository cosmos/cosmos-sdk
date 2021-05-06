package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant/types"
)

func (suite *KeeperTestSuite) TestFeeAllowance() {

	testCases := []struct {
		name      string
		req       *types.QueryAllowanceRequest
		expectErr bool
		preRun    func()
		postRun   func(_ *types.QueryAllowanceResponse)
	}{
		{
			"nil request",
			nil,
			true,
			func() {},
			func(*types.QueryAllowanceResponse) {},
		},
		{
			"fail: invalid granter",
			&types.QueryAllowanceRequest{
				Granter: "invalid_granter",
				Grantee: suite.addrs[0].String(),
			},
			true,
			func() {},
			func(*types.QueryAllowanceResponse) {},
		},
		{
			"fail: invalid grantee",
			&types.QueryAllowanceRequest{
				Granter: suite.addrs[0].String(),
				Grantee: "invalid_grantee",
			},
			true,
			func() {},
			func(*types.QueryAllowanceResponse) {},
		},
		{
			"fail: no grants",
			&types.QueryAllowanceRequest{
				Granter: suite.addrs[0].String(),
				Grantee: suite.addrs[1].String(),
			},
			true,
			func() {},
			func(*types.QueryAllowanceResponse) {},
		},
		{
			"valid query: expect single grant",
			&types.QueryAllowanceRequest{
				Granter: suite.addrs[0].String(),
				Grantee: suite.addrs[1].String(),
			},
			false,
			func() {
				grantFeeAllowance(suite)
			},
			func(response *types.QueryAllowanceResponse) {
				suite.Require().Equal(response.Allowance.Granter, suite.addrs[0].String())
				suite.Require().Equal(response.Allowance.Grantee, suite.addrs[1].String())
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			tc.preRun()
			resp, err := suite.keeper.Allowance(suite.ctx, tc.req)
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
		req       *types.QueryAllowancesRequest
		expectErr bool
		preRun    func()
		postRun   func(_ *types.QueryAllowancesResponse)
	}{
		{
			"nil request",
			nil,
			true,
			func() {},
			func(*types.QueryAllowancesResponse) {},
		},
		{
			"fail: invalid grantee",
			&types.QueryAllowancesRequest{
				Grantee: "invalid_grantee",
			},
			true,
			func() {},
			func(*types.QueryAllowancesResponse) {},
		},
		{
			"no grants",
			&types.QueryAllowancesRequest{
				Grantee: suite.addrs[1].String(),
			},
			false,
			func() {},
			func(resp *types.QueryAllowancesResponse) {
				suite.Require().Equal(len(resp.Allowances), 0)
			},
		},
		{
			"valid query: expect single grant",
			&types.QueryAllowancesRequest{
				Grantee: suite.addrs[1].String(),
			},
			false,
			func() {
				grantFeeAllowance(suite)
			},
			func(resp *types.QueryAllowancesResponse) {
				suite.Require().Equal(len(resp.Allowances), 1)
				suite.Require().Equal(resp.Allowances[0].Granter, suite.addrs[0].String())
				suite.Require().Equal(resp.Allowances[0].Grantee, suite.addrs[1].String())
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			tc.preRun()
			resp, err := suite.keeper.Allowances(suite.ctx, tc.req)
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
	exp := suite.sdkCtx.BlockTime().AddDate(1, 0, 0)
	err := suite.app.FeeGrantKeeper.GrantAllowance(suite.sdkCtx, suite.addrs[0], suite.addrs[1], &types.BasicAllowance{
		SpendLimit: sdk.NewCoins(sdk.NewInt64Coin("atom", 555)),
		Expiration: &exp,
	})
	suite.Require().NoError(err)
}
