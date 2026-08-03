package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
)

const (
	invalidGrantee = "invalid-grantee"
	invalidGranter = "invalid-granter"
)

func (suite *KeeperTestSuite) TestFeeAllowance() {
	testCases := []struct {
		name      string
		req       *feegrant.QueryAllowanceRequest
		expectErr bool
		preRun    func()
		postRun   func(_ *feegrant.QueryAllowanceResponse)
	}{
		{
			"nil request",
			nil,
			true,
			func() {},
			func(*feegrant.QueryAllowanceResponse) {},
		},
		{
			"fail: invalid granter",
			&feegrant.QueryAllowanceRequest{
				Granter: invalidGranter,
				Grantee: suite.addrs[0].String(),
			},
			true,
			func() {},
			func(*feegrant.QueryAllowanceResponse) {},
		},
		{
			"fail: invalid grantee",
			&feegrant.QueryAllowanceRequest{
				Granter: suite.addrs[0].String(),
				Grantee: invalidGrantee,
			},
			true,
			func() {},
			func(*feegrant.QueryAllowanceResponse) {},
		},
		{
			"fail: no grants",
			&feegrant.QueryAllowanceRequest{
				Granter: suite.addrs[0].String(),
				Grantee: suite.addrs[1].String(),
			},
			true,
			func() {},
			func(*feegrant.QueryAllowanceResponse) {},
		},
		{
			"non existed grant",
			&feegrant.QueryAllowanceRequest{
				Granter: invalidGranter,
				Grantee: invalidGrantee,
			},
			true,
			func() {},
			func(*feegrant.QueryAllowanceResponse) {},
		},
		{
			"valid query: expect single grant",
			&feegrant.QueryAllowanceRequest{
				Granter: suite.addrs[0].String(),
				Grantee: suite.addrs[1].String(),
			},
			false,
			func() {
				suite.grantFeeAllowance(suite.addrs[0], suite.addrs[1])
			},
			func(response *feegrant.QueryAllowanceResponse) {
				suite.Require().Equal(response.Allowance.Granter, suite.addrs[0].String())
				suite.Require().Equal(response.Allowance.Grantee, suite.addrs[1].String())
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			tc.preRun()
			resp, err := suite.feegrantKeeper.Allowance(suite.ctx, tc.req)
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
		req       *feegrant.QueryAllowancesRequest
		expectErr bool
		preRun    func()
		postRun   func(_ *feegrant.QueryAllowancesResponse)
	}{
		{
			"nil request",
			nil,
			true,
			func() {},
			func(*feegrant.QueryAllowancesResponse) {},
		},
		{
			"fail: invalid grantee",
			&feegrant.QueryAllowancesRequest{
				Grantee: invalidGrantee,
			},
			true,
			func() {},
			func(*feegrant.QueryAllowancesResponse) {},
		},
		{
			"no grants",
			&feegrant.QueryAllowancesRequest{
				Grantee: suite.addrs[1].String(),
			},
			false,
			func() {},
			func(resp *feegrant.QueryAllowancesResponse) {
				suite.Require().Equal(len(resp.Allowances), 0)
			},
		},
		{
			"valid query: expect single grant",
			&feegrant.QueryAllowancesRequest{
				Grantee: suite.addrs[1].String(),
			},
			false,
			func() {
				suite.grantFeeAllowance(suite.addrs[0], suite.addrs[1])
			},
			func(resp *feegrant.QueryAllowancesResponse) {
				suite.Require().Equal(len(resp.Allowances), 1)
				suite.Require().Equal(resp.Allowances[0].Granter, suite.addrs[0].String())
				suite.Require().Equal(resp.Allowances[0].Grantee, suite.addrs[1].String())
			},
		},
		{
			"valid query: expect multi grant",
			&feegrant.QueryAllowancesRequest{
				Grantee: suite.addrs[0].String(),
			},
			false,
			func() {
				suite.grantFeeAllowance(suite.addrs[1], suite.addrs[0])
				suite.grantFeeAllowance(suite.addrs[2], suite.addrs[0])
				suite.grantFeeAllowance(suite.addrs[3], suite.addrs[0])
			},
			func(resp *feegrant.QueryAllowancesResponse) {
				suite.Require().Equal(len(resp.Allowances), 3)
				for i, addr := range suite.addrs[1:4] {
					resp.Allowances[i].Granter = suite.addrs[0].String()
					resp.Allowances[i].Granter = addr.String()
				}
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			tc.preRun()
			resp, err := suite.feegrantKeeper.Allowances(suite.ctx, tc.req)
			if tc.expectErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
				tc.postRun(resp)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestFeeAllowancesByGranter() {
	testCases := []struct {
		name      string
		req       *feegrant.QueryAllowancesByGranterRequest
		expectErr bool
		preRun    func()
		postRun   func(_ *feegrant.QueryAllowancesByGranterResponse)
	}{
		{
			"nil request",
			nil,
			true,
			func() {},
			func(*feegrant.QueryAllowancesByGranterResponse) {},
		},
		{
			"fail: invalid grantee",
			&feegrant.QueryAllowancesByGranterRequest{
				Granter: invalidGrantee,
			},
			true,
			func() {},
			func(*feegrant.QueryAllowancesByGranterResponse) {},
		},
		{
			"no grants",
			&feegrant.QueryAllowancesByGranterRequest{
				Granter: suite.addrs[0].String(),
			},
			false,
			func() {},
			func(resp *feegrant.QueryAllowancesByGranterResponse) {
				suite.Require().Equal(len(resp.Allowances), 0)
			},
		},
		{
			"valid query: expect single grant",
			&feegrant.QueryAllowancesByGranterRequest{
				Granter: suite.addrs[0].String(),
			},
			false,
			func() {
				suite.grantFeeAllowance(suite.addrs[0], suite.addrs[1])

				// adding this allowance to check whether the pagination working fine.
				suite.grantFeeAllowance(suite.addrs[1], suite.addrs[2])
			},
			func(resp *feegrant.QueryAllowancesByGranterResponse) {
				suite.Require().Equal(len(resp.Allowances), 1)
				suite.Require().Equal(resp.Allowances[0].Granter, suite.addrs[0].String())
				suite.Require().Equal(resp.Allowances[0].Grantee, suite.addrs[1].String())
				suite.Require().Equal(resp.Pagination.Total, uint64(1))
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			tc.preRun()
			resp, err := suite.feegrantKeeper.AllowancesByGranter(suite.ctx, tc.req)
			if tc.expectErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
				tc.postRun(resp)
			}
		})
	}
}

// TestFeeAllowancesPagination checks that the Allowances and AllowancesByGranter
// queries honor the PageRequest offset and count_total. The returned grants must come
// from the pagination transform, not from a side effect inside the predicate: the
// predicate is also invoked for offset-skipped items and for every remaining item during
// the count_total sweep, so appending there over-collects out-of-range grants.
func (suite *KeeperTestSuite) TestFeeAllowancesPagination() {
	grantee := suite.addrs[0]
	suite.grantFeeAllowance(suite.addrs[1], grantee)
	suite.grantFeeAllowance(suite.addrs[2], grantee)
	suite.grantFeeAllowance(suite.addrs[3], grantee)

	// offset must skip grants, not merely advance the page cursor
	resp, err := suite.feegrantKeeper.Allowances(suite.ctx, &feegrant.QueryAllowancesRequest{
		Grantee:    grantee.String(),
		Pagination: &query.PageRequest{Limit: 10, Offset: 1},
	})
	suite.Require().NoError(err)
	suite.Require().Len(resp.Allowances, 2)

	// a limited page with count_total must return only the page, not every match
	resp, err = suite.feegrantKeeper.Allowances(suite.ctx, &feegrant.QueryAllowancesRequest{
		Grantee:    grantee.String(),
		Pagination: &query.PageRequest{Limit: 2, CountTotal: true},
	})
	suite.Require().NoError(err)
	suite.Require().Len(resp.Allowances, 2)
	suite.Require().Equal(uint64(3), resp.Pagination.Total)

	// same guarantees on the granter-filtered query
	granter := suite.addrs[5]
	suite.grantFeeAllowance(granter, suite.addrs[6])
	suite.grantFeeAllowance(granter, suite.addrs[7])
	suite.grantFeeAllowance(granter, suite.addrs[8])

	byGranter, err := suite.feegrantKeeper.AllowancesByGranter(suite.ctx, &feegrant.QueryAllowancesByGranterRequest{
		Granter:    granter.String(),
		Pagination: &query.PageRequest{Limit: 2, CountTotal: true},
	})
	suite.Require().NoError(err)
	suite.Require().Len(byGranter.Allowances, 2)
	suite.Require().Equal(uint64(3), byGranter.Pagination.Total)
}

func (suite *KeeperTestSuite) grantFeeAllowance(granter, grantee sdk.AccAddress) {
	exp := suite.ctx.BlockTime().AddDate(1, 0, 0)
	err := suite.feegrantKeeper.GrantAllowance(suite.ctx, granter, grantee, &feegrant.BasicAllowance{
		SpendLimit: sdk.NewCoins(sdk.NewInt64Coin("atom", 555)),
		Expiration: &exp,
	})
	suite.Require().NoError(err)
}
