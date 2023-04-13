package keeper_test

import (
	"errors"
	"time"

	"cosmossdk.io/x/feegrant"

	codecaddress "github.com/cosmos/cosmos-sdk/codec/address"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/golang/mock/gomock"
)

func (suite *KeeperTestSuite) TestGrantAllowance() {
	ctx := suite.ctx.WithBlockTime(time.Now())
	oneYear := ctx.BlockTime().AddDate(1, 0, 0)
	yesterday := ctx.BlockTime().AddDate(0, 0, -1)

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
				suite.accountKeeper.EXPECT().StringToBytes(invalid).Return(nil, errors.New("decoding bech32 failed")).AnyTimes()

				return &feegrant.MsgGrantAllowance{
					Granter:   invalid,
					Grantee:   suite.addrs[1].String(),
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
				suite.accountKeeper.EXPECT().StringToBytes(invalid).Return(nil, errors.New("decoding bech32 failed")).AnyTimes()

				return &feegrant.MsgGrantAllowance{
					Granter:   suite.addrs[0].String(),
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
				suite.accountKeeper.EXPECT().StringToBytes(grantee).Return(granteeAccAddr, nil).AnyTimes()
				suite.accountKeeper.EXPECT().BytesToString(granteeAccAddr).Return(grantee, nil).AnyTimes()

				acc := authtypes.NewBaseAccountWithAddress(granteeAccAddr)
				add, err := addressCodec.StringToBytes(grantee)
				suite.Require().NoError(err)

				suite.accountKeeper.EXPECT().NewAccountWithAddress(gomock.Any(), add).Return(acc).AnyTimes()
				suite.accountKeeper.EXPECT().SetAccount(gomock.Any(), acc).Return()
				suite.accountKeeper.EXPECT().StringToBytes(grantee).Return(granteeAccAddr, nil).AnyTimes()
				suite.accountKeeper.EXPECT().BytesToString(granteeAccAddr).Return(grantee, nil).AnyTimes()

				suite.Require().NoError(err)
				return &feegrant.MsgGrantAllowance{
					Granter:   suite.addrs[0].String(),
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
					Granter:   suite.addrs[0].String(),
					Grantee:   suite.addrs[1].String(),
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
			func() *feegrant.MsgGrantAllowance {
				any, err := codectypes.NewAnyWithValue(&feegrant.BasicAllowance{
					SpendLimit: suite.atom,
					Expiration: &oneYear,
				})
				suite.Require().NoError(err)
				return &feegrant.MsgGrantAllowance{
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
			_, err := suite.msgSrvr.GrantAllowance(ctx, tc.req())
			if tc.expectErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.errMsg)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestRevokeAllowance() {
	oneYear := suite.ctx.BlockTime().AddDate(1, 0, 0)

	invalidGranter := "invalid-granter"
	suite.accountKeeper.EXPECT().StringToBytes(invalidGranter).Return(nil, errors.New("decoding bech32 failed")).AnyTimes()

	invalidGrantee := "invalid-grantee"
	suite.accountKeeper.EXPECT().StringToBytes(invalidGrantee).Return(nil, errors.New("decoding bech32 failed")).AnyTimes()

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
				Grantee: suite.addrs[1].String(),
			},
			func() {},
			true,
			"decoding bech32 failed",
		},
		{
			"error: invalid grantee",
			&feegrant.MsgRevokeAllowance{
				Granter: suite.addrs[0].String(),
				Grantee: invalidGrantee,
			},
			func() {},
			true,
			"decoding bech32 failed",
		},
		{
			"error: fee allowance not found",
			&feegrant.MsgRevokeAllowance{
				Granter: suite.addrs[0].String(),
				Grantee: suite.addrs[1].String(),
			},
			func() {},
			true,
			"fee-grant not found",
		},
		{
			"success: revoke fee allowance",
			&feegrant.MsgRevokeAllowance{
				Granter: suite.addrs[0].String(),
				Grantee: suite.addrs[1].String(),
			},
			func() {
				// removing fee allowance from previous tests if exists
				suite.msgSrvr.RevokeAllowance(suite.ctx, &feegrant.MsgRevokeAllowance{
					Granter: suite.addrs[0].String(),
					Grantee: suite.addrs[1].String(),
				})

				any, err := codectypes.NewAnyWithValue(&feegrant.PeriodicAllowance{
					Basic: feegrant.BasicAllowance{
						SpendLimit: suite.atom,
						Expiration: &oneYear,
					},
					PeriodSpendLimit: suite.atom,
				})
				suite.Require().NoError(err)
				req := &feegrant.MsgGrantAllowance{
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
			&feegrant.MsgRevokeAllowance{
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
