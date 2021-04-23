package keeper_test

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant/keeper"

	"github.com/cosmos/cosmos-sdk/x/feegrant/types"
)

func (suite *KeeperTestSuite) TestGrantFeeAllowance() {
	ctx := suite.ctx
	wrapCtx := sdk.WrapSDKContext(ctx)
	k := suite.app.FeeGrantKeeper
	impl := keeper.NewMsgServerImpl(k)
	atoms := sdk.NewCoins(sdk.NewCoin("atom", sdk.NewInt(1000)))

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
					SpendLimit: atoms,
					Expiration: types.ExpiresAtTime(suite.ctx.BlockTime().AddDate(1, 0, 0)),
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
					SpendLimit: atoms,
					Expiration: types.ExpiresAtTime(suite.ctx.BlockTime().AddDate(1, 0, 0)),
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
						SpendLimit: atoms,
						Expiration: types.ExpiresAtTime(suite.ctx.BlockTime().AddDate(1, 0, 0)),
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
						SpendLimit: atoms,
						Expiration: types.ExpiresAtTime(suite.ctx.BlockTime().AddDate(1, 0, 0)),
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
			_, err := impl.GrantFeeAllowance(wrapCtx, tc.req())
			if tc.expectErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.errMsg)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestRevokeFeeAllowance() {
	ctx := suite.ctx
	wrapCtx := sdk.WrapSDKContext(ctx)
	k := suite.app.FeeGrantKeeper
	impl := keeper.NewMsgServerImpl(k)
	atoms := sdk.NewCoins(sdk.NewCoin("atom", sdk.NewInt(1000)))

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
				impl.RevokeFeeAllowance(wrapCtx, &types.MsgRevokeFeeAllowance{
					Granter: suite.addrs[0].String(),
					Grantee: suite.addrs[1].String(),
				})
				any, err := codectypes.NewAnyWithValue(&types.PeriodicFeeAllowance{
					Basic: types.BasicFeeAllowance{
						SpendLimit: atoms,
						Expiration: types.ExpiresAtTime(suite.ctx.BlockTime().AddDate(1, 0, 0)),
					},
				})
				suite.Require().NoError(err)
				req := &types.MsgGrantFeeAllowance{
					Granter:   suite.addrs[0].String(),
					Grantee:   suite.addrs[1].String(),
					Allowance: any,
				}
				_, err = impl.GrantFeeAllowance(wrapCtx, req)
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
			_, err := impl.RevokeFeeAllowance(wrapCtx, tc.request)
			if tc.expectErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.errMsg)
			}
		})
	}

}
