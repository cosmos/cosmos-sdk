package keeper_test

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"

	"github.com/cosmos/cosmos-sdk/x/feegrant/types"
)

func (suite *KeeperTestSuite) TestGrantFeeAllowance() {
	oneYear := suite.sdkCtx.BlockTime().AddDate(1, 0, 0)

	testCases := []struct {
		name      string
		req       func() *types.MsgGrantFeeAllowance
		expectErr bool
		errMsg    string
	}{
		{
			"invalid granter address",
			func() *types.MsgGrantFeeAllowance {
				any, err := codectypes.NewAnyWithValue(&types.BasicFeeAllowance{})
				suite.Require().NoError(err)
				return &types.MsgGrantFeeAllowance{
					Granter:   "invalid-granter",
					Grantee:   suite.addrs[1].String(),
					Allowance: any,
				}
			},
			true,
			"decoding bech32 failed",
		},
		{
			"invalid grantee address",
			func() *types.MsgGrantFeeAllowance {
				any, err := codectypes.NewAnyWithValue(&types.BasicFeeAllowance{})
				suite.Require().NoError(err)
				return &types.MsgGrantFeeAllowance{
					Granter:   suite.addrs[0].String(),
					Grantee:   "invalid-grantee",
					Allowance: any,
				}
			},
			true,
			"decoding bech32 failed",
		},
		{
			"valid: basic fee allowance",
			func() *types.MsgGrantFeeAllowance {
				any, err := codectypes.NewAnyWithValue(&types.BasicFeeAllowance{
					SpendLimit: suite.atom,
					Expiration: &oneYear,
				})
				suite.Require().NoError(err)
				return &types.MsgGrantFeeAllowance{
					Granter:   suite.addrs[0].String(),
					Grantee:   suite.addrs[1].String(),
					Allowance: any,
				}
			},
			false,
			"",
		},
		{
			"fail: fee allowance exists",
			func() *types.MsgGrantFeeAllowance {
				any, err := codectypes.NewAnyWithValue(&types.BasicFeeAllowance{
					SpendLimit: suite.atom,
					Expiration: &oneYear,
				})
				suite.Require().NoError(err)
				return &types.MsgGrantFeeAllowance{
					Granter:   suite.addrs[0].String(),
					Grantee:   suite.addrs[1].String(),
					Allowance: any,
				}
			},
			true,
			"fee allowance already exists",
		},
		{
			"valid: periodic fee allowance",
			func() *types.MsgGrantFeeAllowance {
				any, err := codectypes.NewAnyWithValue(&types.PeriodicFeeAllowance{
					Basic: types.BasicFeeAllowance{
						SpendLimit: suite.atom,
						Expiration: &oneYear,
					},
				})
				suite.Require().NoError(err)
				return &types.MsgGrantFeeAllowance{
					Granter:   suite.addrs[1].String(),
					Grantee:   suite.addrs[2].String(),
					Allowance: any,
				}
			},
			false,
			"",
		},
		{
			"error: fee allowance exists",
			func() *types.MsgGrantFeeAllowance {
				any, err := codectypes.NewAnyWithValue(&types.PeriodicFeeAllowance{
					Basic: types.BasicFeeAllowance{
						SpendLimit: suite.atom,
						Expiration: &oneYear,
					},
				})
				suite.Require().NoError(err)
				return &types.MsgGrantFeeAllowance{
					Granter:   suite.addrs[1].String(),
					Grantee:   suite.addrs[2].String(),
					Allowance: any,
				}
			},
			true,
			"fee allowance already exists",
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			_, err := suite.msgSrvr.GrantFeeAllowance(suite.ctx, tc.req())
			if tc.expectErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.errMsg)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestRevokeFeeAllowance() {
	oneYear := suite.sdkCtx.BlockTime().AddDate(1, 0, 0)

	testCases := []struct {
		name      string
		request   *types.MsgRevokeFeeAllowance
		preRun    func()
		expectErr bool
		errMsg    string
	}{
		{
			"error: invalid granter",
			&types.MsgRevokeFeeAllowance{
				Granter: "invalid-granter",
				Grantee: suite.addrs[1].String(),
			},
			func() {},
			true,
			"decoding bech32 failed",
		},
		{
			"error: invalid grantee",
			&types.MsgRevokeFeeAllowance{
				Granter: suite.addrs[0].String(),
				Grantee: "invalid-grantee",
			},
			func() {},
			true,
			"decoding bech32 failed",
		},
		{
			"error: fee allowance not found",
			&types.MsgRevokeFeeAllowance{
				Granter: suite.addrs[0].String(),
				Grantee: suite.addrs[1].String(),
			},
			func() {},
			true,
			"fee-grant not found",
		},
		{
			"success: revoke fee allowance",
			&types.MsgRevokeFeeAllowance{
				Granter: suite.addrs[0].String(),
				Grantee: suite.addrs[1].String(),
			},
			func() {
				// removing fee allowance from previous tests if exists
				suite.msgSrvr.RevokeFeeAllowance(suite.ctx, &types.MsgRevokeFeeAllowance{
					Granter: suite.addrs[0].String(),
					Grantee: suite.addrs[1].String(),
				})
				any, err := codectypes.NewAnyWithValue(&types.PeriodicFeeAllowance{
					Basic: types.BasicFeeAllowance{
						SpendLimit: suite.atom,
						Expiration: &oneYear,
					},
				})
				suite.Require().NoError(err)
				req := &types.MsgGrantFeeAllowance{
					Granter:   suite.addrs[0].String(),
					Grantee:   suite.addrs[1].String(),
					Allowance: any,
				}
				_, err = suite.msgSrvr.GrantFeeAllowance(suite.ctx, req)
				suite.Require().NoError(err)
			},
			false,
			"",
		},
		{
			"error: check fee allowance revoked",
			&types.MsgRevokeFeeAllowance{
				Granter: suite.addrs[0].String(),
				Grantee: suite.addrs[1].String(),
			},
			func() {},
			true,
			"fee-grant not found",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			tc.preRun()
			_, err := suite.msgSrvr.RevokeFeeAllowance(suite.ctx, tc.request)
			if tc.expectErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.errMsg)
			}
		})
	}

}
