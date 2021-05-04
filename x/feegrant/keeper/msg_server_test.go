package keeper_test

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"

	"github.com/cosmos/cosmos-sdk/x/feegrant/types"
)

func (suite *KeeperTestSuite) TestGrantFeeAllowance() {
	testCases := []struct {
		name      string
		req       func() *types.MsgGrantAllowance
		expectErr bool
		errMsg    string
	}{
		{
			"invalid granter address",
			func() *types.MsgGrantAllowance {
				any, err := codectypes.NewAnyWithValue(&types.BasicAllowance{})
				suite.Require().NoError(err)
				return &types.MsgGrantAllowance{
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
			func() *types.MsgGrantAllowance {
				any, err := codectypes.NewAnyWithValue(&types.BasicAllowance{})
				suite.Require().NoError(err)
				return &types.MsgGrantAllowance{
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
			func() *types.MsgGrantAllowance {
				any, err := codectypes.NewAnyWithValue(&types.BasicAllowance{
					SpendLimit: suite.atom,
					Expiration: types.ExpiresAtTime(suite.sdkCtx.BlockTime().AddDate(1, 0, 0)),
				})
				suite.Require().NoError(err)
				return &types.MsgGrantAllowance{
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
			func() *types.MsgGrantAllowance {
				any, err := codectypes.NewAnyWithValue(&types.BasicAllowance{
					SpendLimit: suite.atom,
					Expiration: types.ExpiresAtTime(suite.sdkCtx.BlockTime().AddDate(1, 0, 0)),
				})
				suite.Require().NoError(err)
				return &types.MsgGrantAllowance{
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
			func() *types.MsgGrantAllowance {
				any, err := codectypes.NewAnyWithValue(&types.PeriodicAllowance{
					Basic: types.BasicAllowance{
						SpendLimit: suite.atom,
						Expiration: types.ExpiresAtTime(suite.sdkCtx.BlockTime().AddDate(1, 0, 0)),
					},
				})
				suite.Require().NoError(err)
				return &types.MsgGrantAllowance{
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
			func() *types.MsgGrantAllowance {
				any, err := codectypes.NewAnyWithValue(&types.PeriodicAllowance{
					Basic: types.BasicAllowance{
						SpendLimit: suite.atom,
						Expiration: types.ExpiresAtTime(suite.sdkCtx.BlockTime().AddDate(1, 0, 0)),
					},
				})
				suite.Require().NoError(err)
				return &types.MsgGrantAllowance{
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
			_, err := suite.msgSrvr.GrantAllowance(suite.ctx, tc.req())
			if tc.expectErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.errMsg)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestRevokeFeeAllowance() {

	testCases := []struct {
		name      string
		request   *types.MsgRevokeAllowance
		preRun    func()
		expectErr bool
		errMsg    string
	}{
		{
			"error: invalid granter",
			&types.MsgRevokeAllowance{
				Granter: "invalid-granter",
				Grantee: suite.addrs[1].String(),
			},
			func() {},
			true,
			"decoding bech32 failed",
		},
		{
			"error: invalid grantee",
			&types.MsgRevokeAllowance{
				Granter: suite.addrs[0].String(),
				Grantee: "invalid-grantee",
			},
			func() {},
			true,
			"decoding bech32 failed",
		},
		{
			"error: fee allowance not found",
			&types.MsgRevokeAllowance{
				Granter: suite.addrs[0].String(),
				Grantee: suite.addrs[1].String(),
			},
			func() {},
			true,
			"fee-grant not found",
		},
		{
			"success: revoke fee allowance",
			&types.MsgRevokeAllowance{
				Granter: suite.addrs[0].String(),
				Grantee: suite.addrs[1].String(),
			},
			func() {
				// removing fee allowance from previous tests if exists
				suite.msgSrvr.RevokeAllowance(suite.ctx, &types.MsgRevokeAllowance{
					Granter: suite.addrs[0].String(),
					Grantee: suite.addrs[1].String(),
				})
				any, err := codectypes.NewAnyWithValue(&types.PeriodicAllowance{
					Basic: types.BasicAllowance{
						SpendLimit: suite.atom,
						Expiration: types.ExpiresAtTime(suite.sdkCtx.BlockTime().AddDate(1, 0, 0)),
					},
				})
				suite.Require().NoError(err)
				req := &types.MsgGrantAllowance{
					Granter:   suite.addrs[0].String(),
					Grantee:   suite.addrs[1].String(),
					Allowance: any,
				}
				_, err = suite.msgSrvr.GrantAllowance(suite.ctx, req)
				suite.Require().NoError(err)
			},
			false,
			"",
		},
		{
			"error: check fee allowance revoked",
			&types.MsgRevokeAllowance{
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
			_, err := suite.msgSrvr.RevokeAllowance(suite.ctx, tc.request)
			if tc.expectErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.errMsg)
			}
		})
	}

}
