package keeper_test

import (
	"time"

	"github.com/golang/mock/gomock"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/header"
	authtypes "cosmossdk.io/x/auth/types"
	"cosmossdk.io/x/feegrant"

	codecaddress "github.com/cosmos/cosmos-sdk/codec/address"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types"
)

func (suite *KeeperTestSuite) TestGrantAllowance() {
	ctx := suite.ctx.WithHeaderInfo(header.Info{Time: time.Now()})
	oneYear := ctx.HeaderInfo().Time.AddDate(1, 0, 0)
	yesterday := ctx.HeaderInfo().Time.AddDate(0, 0, -1)

	addressCodec := codecaddress.NewBech32Codec("cosmos")

	testCases := []struct {
		name      string
		req       func() *feegrant.MsgGrantAllowance
		expectErr bool
		errMsg    string
	}{
		{
			"invalid granter address",
			func() *feegrant.MsgGrantAllowance {
				any, err := codectypes.NewAnyWithValue(&feegrant.BasicAllowance{})
				suite.Require().NoError(err)
				invalid := "invalid-granter"
				return &feegrant.MsgGrantAllowance{
					Granter:   invalid,
					Grantee:   suite.encodedAddrs[1],
					Allowance: any,
				}
			},
			true,
			"decoding bech32 failed",
		},
		{
			"invalid grantee address",
			func() *feegrant.MsgGrantAllowance {
				any, err := codectypes.NewAnyWithValue(&feegrant.BasicAllowance{})
				suite.Require().NoError(err)
				invalid := "invalid-grantee"
				return &feegrant.MsgGrantAllowance{
					Granter:   suite.encodedAddrs[0],
					Grantee:   invalid,
					Allowance: any,
				}
			},
			true,
			"decoding bech32 failed",
		},
		{
			"valid: grantee account doesn't exist",
			func() *feegrant.MsgGrantAllowance {
				grantee := "cosmos139f7kncmglres2nf3h4hc4tade85ekfr8sulz5"
				granteeAccAddr, err := addressCodec.StringToBytes(grantee)
				suite.Require().NoError(err)
				any, err := codectypes.NewAnyWithValue(&feegrant.BasicAllowance{
					SpendLimit: suite.atom,
					Expiration: &oneYear,
				})
				suite.Require().NoError(err)

				suite.accountKeeper.EXPECT().GetAccount(gomock.Any(), granteeAccAddr).Return(nil).AnyTimes()

				acc := authtypes.NewBaseAccountWithAddress(granteeAccAddr)
				add, err := addressCodec.StringToBytes(grantee)
				suite.Require().NoError(err)

				suite.accountKeeper.EXPECT().NewAccountWithAddress(gomock.Any(), add).Return(acc).AnyTimes()

				suite.Require().NoError(err)
				return &feegrant.MsgGrantAllowance{
					Granter:   suite.encodedAddrs[0],
					Grantee:   grantee,
					Allowance: any,
				}
			},
			false,
			"",
		},
		{
			"invalid: past expiry",
			func() *feegrant.MsgGrantAllowance {
				any, err := codectypes.NewAnyWithValue(&feegrant.BasicAllowance{
					SpendLimit: suite.atom,
					Expiration: &yesterday,
				})
				suite.Require().NoError(err)
				return &feegrant.MsgGrantAllowance{
					Granter:   suite.encodedAddrs[0],
					Grantee:   suite.encodedAddrs[1],
					Allowance: any,
				}
			},
			true,
			"expiration is before current block time",
		},
		{
			"valid: basic fee allowance",
			func() *feegrant.MsgGrantAllowance {
				any, err := codectypes.NewAnyWithValue(&feegrant.BasicAllowance{
					SpendLimit: suite.atom,
					Expiration: &oneYear,
				})
				suite.Require().NoError(err)
				return &feegrant.MsgGrantAllowance{
					Granter:   suite.encodedAddrs[0],
					Grantee:   suite.encodedAddrs[1],
					Allowance: any,
				}
			},
			false,
			"",
		},
		{
			"fail: fee allowance exists",
			func() *feegrant.MsgGrantAllowance {
				any, err := codectypes.NewAnyWithValue(&feegrant.BasicAllowance{
					SpendLimit: suite.atom,
					Expiration: &oneYear,
				})
				suite.Require().NoError(err)
				return &feegrant.MsgGrantAllowance{
					Granter:   suite.encodedAddrs[0],
					Grantee:   suite.encodedAddrs[1],
					Allowance: any,
				}
			},
			true,
			"fee allowance already exists",
		},
		{
			"valid: periodic fee allowance",
			func() *feegrant.MsgGrantAllowance {
				any, err := codectypes.NewAnyWithValue(&feegrant.PeriodicAllowance{
					Basic: feegrant.BasicAllowance{
						SpendLimit: suite.atom,
						Expiration: &oneYear,
					},
					PeriodSpendLimit: suite.atom,
				})
				suite.Require().NoError(err)
				return &feegrant.MsgGrantAllowance{
					Granter:   suite.encodedAddrs[1],
					Grantee:   suite.encodedAddrs[2],
					Allowance: any,
				}
			},
			false,
			"",
		},
		{
			"valid: with period reset",
			func() *feegrant.MsgGrantAllowance {
				any, err := codectypes.NewAnyWithValue(&feegrant.PeriodicAllowance{
					Basic: feegrant.BasicAllowance{
						SpendLimit: suite.atom,
						Expiration: &oneYear,
					},
					Period:           time.Hour,
					PeriodSpendLimit: suite.atom,
					PeriodReset:      oneYear,
				})
				suite.Require().NoError(err)
				return &feegrant.MsgGrantAllowance{
					Granter:   suite.encodedAddrs[1],
					Grantee:   suite.encodedAddrs[2],
					Allowance: any,
				}
			},
			false,
			"",
		},
		{
			"error: fee allowance exists",
			func() *feegrant.MsgGrantAllowance {
				any, err := codectypes.NewAnyWithValue(&feegrant.PeriodicAllowance{
					Basic: feegrant.BasicAllowance{
						SpendLimit: suite.atom,
						Expiration: &oneYear,
					},
					PeriodSpendLimit: suite.atom,
				})
				suite.Require().NoError(err)
				return &feegrant.MsgGrantAllowance{
					Granter:   suite.encodedAddrs[1],
					Grantee:   suite.encodedAddrs[2],
					Allowance: any,
				}
			},
			true,
			"fee allowance already exists",
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			_, err := suite.msgSrvr.GrantAllowance(ctx, tc.req())
			if tc.expectErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.errMsg)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestRevokeAllowance() {
	suite.ctx = suite.ctx.WithHeaderInfo(header.Info{Time: time.Now()})
	oneYear := suite.ctx.HeaderInfo().Time.AddDate(1, 0, 0)

	testCases := []struct {
		name      string
		request   *feegrant.MsgRevokeAllowance
		preRun    func()
		expectErr bool
		errMsg    string
	}{
		{
			"error: invalid granter",
			&feegrant.MsgRevokeAllowance{
				Granter: invalidGranter,
				Grantee: suite.encodedAddrs[1],
			},
			func() {},
			true,
			"decoding bech32 failed",
		},
		{
			"error: invalid grantee",
			&feegrant.MsgRevokeAllowance{
				Granter: suite.encodedAddrs[0],
				Grantee: invalidGrantee,
			},
			func() {},
			true,
			"decoding bech32 failed",
		},
		{
			"error: fee allowance not found",
			&feegrant.MsgRevokeAllowance{
				Granter: suite.encodedAddrs[0],
				Grantee: suite.encodedAddrs[1],
			},
			func() {},
			true,
			"not found",
		},
		{
			"success: revoke fee allowance",
			&feegrant.MsgRevokeAllowance{
				Granter: suite.encodedAddrs[0],
				Grantee: suite.encodedAddrs[1],
			},
			func() {
				// removing fee allowance from previous tests if exists
				_, err := suite.msgSrvr.RevokeAllowance(suite.ctx, &feegrant.MsgRevokeAllowance{
					Granter: suite.encodedAddrs[0],
					Grantee: suite.encodedAddrs[1],
				})
				suite.Require().Error(err)
				any, err := codectypes.NewAnyWithValue(&feegrant.PeriodicAllowance{
					Basic: feegrant.BasicAllowance{
						SpendLimit: suite.atom,
						Expiration: &oneYear,
					},
					PeriodSpendLimit: suite.atom,
				})
				suite.Require().NoError(err)
				req := &feegrant.MsgGrantAllowance{
					Granter:   suite.encodedAddrs[0],
					Grantee:   suite.encodedAddrs[1],
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
			&feegrant.MsgRevokeAllowance{
				Granter: suite.encodedAddrs[0],
				Grantee: suite.encodedAddrs[1],
			},
			func() {},
			true,
			"not found",
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

func (suite *KeeperTestSuite) TestPruneAllowances() {
	ctx := suite.ctx.WithHeaderInfo(header.Info{Time: time.Now()})
	oneYear := ctx.HeaderInfo().Time.AddDate(1, 0, 0)

	// We create 76 allowances, all expiring in one year
	count := 0
	for i := 0; i < len(suite.encodedAddrs); i++ {
		for j := 0; j < len(suite.encodedAddrs); j++ {
			if count == 76 {
				break
			}
			if suite.encodedAddrs[i] == suite.encodedAddrs[j] {
				continue
			}

			any, err := codectypes.NewAnyWithValue(&feegrant.BasicAllowance{
				SpendLimit: suite.atom,
				Expiration: &oneYear,
			})
			suite.Require().NoError(err)
			req := &feegrant.MsgGrantAllowance{
				Granter:   suite.encodedAddrs[i],
				Grantee:   suite.encodedAddrs[j],
				Allowance: any,
			}

			_, err = suite.msgSrvr.GrantAllowance(ctx, req)
			if err != nil {
				// do not fail, just try with another pair
				continue
			}

			count++
		}
	}

	// we have 76 allowances
	count = 0
	err := suite.feegrantKeeper.FeeAllowance.Walk(ctx, nil, func(key collections.Pair[types.AccAddress, types.AccAddress], value feegrant.Grant) (stop bool, err error) {
		count++
		return false, nil
	})
	suite.Require().NoError(err)
	suite.Require().Equal(76, count)

	// after a year and one day passes, they are all expired
	oneYearAndADay := ctx.HeaderInfo().Time.AddDate(1, 0, 1)
	ctx = suite.ctx.WithHeaderInfo(header.Info{Time: oneYearAndADay})

	// we prune them, but currently only 75 will be pruned
	_, err = suite.msgSrvr.PruneAllowances(ctx, &feegrant.MsgPruneAllowances{})
	suite.Require().NoError(err)

	// we have 1 allowance left
	count = 0
	err = suite.feegrantKeeper.FeeAllowance.Walk(ctx, nil, func(key collections.Pair[types.AccAddress, types.AccAddress], value feegrant.Grant) (stop bool, err error) {
		count++

		return false, nil
	})
	suite.Require().NoError(err)
	suite.Require().Equal(1, count)
}
