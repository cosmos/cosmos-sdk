package keeper_test

import (
	"time"

	"github.com/golang/mock/gomock"

	"cosmossdk.io/core/header"
	sdkmath "cosmossdk.io/math"
	authtypes "cosmossdk.io/x/auth/types"
	"cosmossdk.io/x/authz"
	banktypes "cosmossdk.io/x/bank/types"

	"github.com/cosmos/cosmos-sdk/codec/address"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (suite *TestSuite) createAccounts() []sdk.AccAddress {
	addrs := simtestutil.CreateIncrementalAccounts(2)
	suite.accountKeeper.EXPECT().GetAccount(gomock.Any(), suite.addrs[0]).Return(authtypes.NewBaseAccountWithAddress(suite.addrs[0])).AnyTimes()
	suite.accountKeeper.EXPECT().GetAccount(gomock.Any(), suite.addrs[1]).Return(authtypes.NewBaseAccountWithAddress(suite.addrs[1])).AnyTimes()
	return addrs
}

func (suite *TestSuite) TestGrant() {
	ctx := suite.ctx.WithHeaderInfo(header.Info{Time: time.Now()})
	addrs := suite.createAccounts()
	curBlockTime := ctx.HeaderInfo().Time

	suite.accountKeeper.EXPECT().AddressCodec().Return(address.NewBech32Codec("cosmos")).AnyTimes()

	oneHour := curBlockTime.Add(time.Hour)
	oneYear := curBlockTime.AddDate(1, 0, 0)

	coins := sdk.NewCoins(sdk.NewCoin("steak", sdkmath.NewInt(10)))

	grantee, granter := addrs[0], addrs[1]

	testCases := []struct {
		name     string
		malleate func() *authz.MsgGrant
		expErr   bool
		errMsg   string
	}{
		{
			name: "identical grantee and granter",
			malleate: func() *authz.MsgGrant {
				grant, err := authz.NewGrant(curBlockTime, banktypes.NewSendAuthorization(coins, nil), &oneYear)
				suite.Require().NoError(err)
				return &authz.MsgGrant{
					Granter: grantee.String(),
					Grantee: grantee.String(),
					Grant:   grant,
				}
			},
			expErr: true,
			errMsg: "grantee and granter should be different",
		},
		{
			name: "invalid granter",
			malleate: func() *authz.MsgGrant {
				grant, err := authz.NewGrant(curBlockTime, banktypes.NewSendAuthorization(coins, nil), &oneYear)
				suite.Require().NoError(err)
				return &authz.MsgGrant{
					Granter: "invalid",
					Grantee: grantee.String(),
					Grant:   grant,
				}
			},
			expErr: true,
			errMsg: "invalid bech32 string",
		},
		{
			name: "invalid grantee",
			malleate: func() *authz.MsgGrant {
				grant, err := authz.NewGrant(curBlockTime, banktypes.NewSendAuthorization(coins, nil), &oneYear)
				suite.Require().NoError(err)
				return &authz.MsgGrant{
					Granter: granter.String(),
					Grantee: "invalid",
					Grant:   grant,
				}
			},
			expErr: true,
			errMsg: "invalid bech32 string",
		},
		{
			name: "invalid grant",
			malleate: func() *authz.MsgGrant {
				return &authz.MsgGrant{
					Granter: granter.String(),
					Grantee: grantee.String(),
					Grant: authz.Grant{
						Expiration: &oneYear,
					},
				}
			},
			expErr: true,
			errMsg: "authorization is nil: invalid type",
		},
		{
			name: "invalid grant, past time",
			malleate: func() *authz.MsgGrant {
				pTime := curBlockTime.Add(-time.Hour)
				grant, err := authz.NewGrant(curBlockTime, banktypes.NewSendAuthorization(coins, nil), &oneHour) // we only need the authorization
				suite.Require().NoError(err)
				return &authz.MsgGrant{
					Granter: granter.String(),
					Grantee: grantee.String(),
					Grant: authz.Grant{
						Authorization: grant.Authorization,
						Expiration:    &pTime,
					},
				}
			},
			expErr: true,
			errMsg: "expiration must be after the current block time",
		},
		{
			name: "grantee account does not exist on chain: valid grant",
			malleate: func() *authz.MsgGrant {
				newAcc := sdk.AccAddress("valid")
				suite.accountKeeper.EXPECT().GetAccount(gomock.Any(), newAcc).Return(nil).AnyTimes()
				acc := authtypes.NewBaseAccountWithAddress(newAcc)
				suite.accountKeeper.EXPECT().NewAccountWithAddress(gomock.Any(), newAcc).Return(acc).AnyTimes()
				suite.accountKeeper.EXPECT().SetAccount(gomock.Any(), acc).Return()

				grant, err := authz.NewGrant(curBlockTime, banktypes.NewSendAuthorization(coins, nil), &oneYear)
				suite.Require().NoError(err)
				return &authz.MsgGrant{
					Granter: granter.String(),
					Grantee: newAcc.String(),
					Grant:   grant,
				}
			},
		},
		{
			name: "valid grant",
			malleate: func() *authz.MsgGrant {
				grant, err := authz.NewGrant(curBlockTime, banktypes.NewSendAuthorization(coins, nil), &oneYear)
				suite.Require().NoError(err)
				return &authz.MsgGrant{
					Granter: granter.String(),
					Grantee: grantee.String(),
					Grant:   grant,
				}
			},
		},
		{
			name: "valid grant, same grantee, granter pair but different msgType",
			malleate: func() *authz.MsgGrant {
				g, err := authz.NewGrant(curBlockTime, banktypes.NewSendAuthorization(coins, nil), &oneHour)
				suite.Require().NoError(err)
				_, err = suite.msgSrvr.Grant(suite.ctx, &authz.MsgGrant{
					Granter: granter.String(),
					Grantee: grantee.String(),
					Grant:   g,
				})
				suite.Require().NoError(err)

				grant, err := authz.NewGrant(curBlockTime, authz.NewGenericAuthorization("/cosmos.bank.v1beta1.MsgUpdateParams"), &oneHour)
				suite.Require().NoError(err)
				return &authz.MsgGrant{
					Granter: granter.String(),
					Grantee: grantee.String(),
					Grant:   grant,
				}
			},
		},
		{
			name: "valid grant with allow list",
			malleate: func() *authz.MsgGrant {
				grant, err := authz.NewGrant(curBlockTime, banktypes.NewSendAuthorization(coins, []sdk.AccAddress{granter}), &oneYear)
				suite.Require().NoError(err)
				return &authz.MsgGrant{
					Granter: granter.String(),
					Grantee: grantee.String(),
					Grant:   grant,
				}
			},
		},
		{
			name: "valid grant with nil expiration time",
			malleate: func() *authz.MsgGrant {
				grant, err := authz.NewGrant(curBlockTime, banktypes.NewSendAuthorization(coins, nil), nil)
				suite.Require().NoError(err)
				return &authz.MsgGrant{
					Granter: granter.String(),
					Grantee: grantee.String(),
					Grant:   grant,
				}
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			_, err := suite.msgSrvr.Grant(suite.ctx, tc.malleate())
			if tc.expErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.errMsg)
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}

func (suite *TestSuite) TestRevoke() {
	addrs := suite.createAccounts()

	grantee, granter := addrs[0], addrs[1]

	testCases := []struct {
		name     string
		malleate func() *authz.MsgRevoke
		expErr   bool
		errMsg   string
	}{
		{
			name: "identical grantee and granter",
			malleate: func() *authz.MsgRevoke {
				return &authz.MsgRevoke{
					Granter:    grantee.String(),
					Grantee:    grantee.String(),
					MsgTypeUrl: bankSendAuthMsgType,
				}
			},
			expErr: true,
			errMsg: "grantee and granter should be different",
		},
		{
			name: "invalid granter",
			malleate: func() *authz.MsgRevoke {
				return &authz.MsgRevoke{
					Granter:    "invalid",
					Grantee:    grantee.String(),
					MsgTypeUrl: bankSendAuthMsgType,
				}
			},
			expErr: true,
			errMsg: "invalid bech32 string",
		},
		{
			name: "invalid grantee",
			malleate: func() *authz.MsgRevoke {
				return &authz.MsgRevoke{
					Granter:    granter.String(),
					Grantee:    "invalid",
					MsgTypeUrl: bankSendAuthMsgType,
				}
			},
			expErr: true,
			errMsg: "invalid bech32 string",
		},
		{
			name: "no msg given",
			malleate: func() *authz.MsgRevoke {
				return &authz.MsgRevoke{
					Granter:    granter.String(),
					Grantee:    grantee.String(),
					MsgTypeUrl: "",
				}
			},
			expErr: true,
			errMsg: "missing msg method name",
		},
		{
			name: "valid grant",
			malleate: func() *authz.MsgRevoke {
				suite.createSendAuthorization(grantee, granter)

				return &authz.MsgRevoke{
					Granter:    granter.String(),
					Grantee:    grantee.String(),
					MsgTypeUrl: bankSendAuthMsgType,
				}
			},
		},
		{
			name: "no existing grant to revoke",
			malleate: func() *authz.MsgRevoke {
				return &authz.MsgRevoke{
					Granter:    granter.String(),
					Grantee:    grantee.String(),
					MsgTypeUrl: bankSendAuthMsgType,
				}
			},
			expErr: true,
			errMsg: "authorization not found",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			_, err := suite.msgSrvr.Revoke(suite.ctx, tc.malleate())
			if tc.expErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.errMsg)
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}

func (suite *TestSuite) TestExec() {
	addrs := suite.createAccounts()

	grantee, granter := addrs[0], addrs[1]
	coins := sdk.NewCoins(sdk.NewCoin("steak", sdkmath.NewInt(10)))

	msg := &banktypes.MsgSend{
		FromAddress: granter.String(),
		ToAddress:   grantee.String(),
		Amount:      coins,
	}

	testCases := []struct {
		name     string
		malleate func() authz.MsgExec
		expErr   bool
		errMsg   string
	}{
		{
			name: "invalid grantee (empty)",
			malleate: func() authz.MsgExec {
				return authz.NewMsgExec(sdk.AccAddress{}, []sdk.Msg{msg})
			},
			expErr: true,
			errMsg: "empty address string is not allowed",
		},
		{
			name: "non existing grant",
			malleate: func() authz.MsgExec {
				return authz.NewMsgExec(grantee, []sdk.Msg{msg})
			},
			expErr: true,
			errMsg: "authorization not found",
		},
		{
			name: "no message case",
			malleate: func() authz.MsgExec {
				return authz.NewMsgExec(grantee, []sdk.Msg{})
			},
			expErr: true,
			errMsg: "messages cannot be empty",
		},
		{
			name: "valid case",
			malleate: func() authz.MsgExec {
				suite.createSendAuthorization(grantee, granter)
				return authz.NewMsgExec(grantee, []sdk.Msg{msg})
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			req := tc.malleate()
			_, err := suite.msgSrvr.Exec(suite.ctx, &req)
			if tc.expErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.errMsg)
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}

func (suite *TestSuite) TestPruneExpiredGrants() {
	addrs := suite.createAccounts()

	timeNow := suite.ctx.BlockTime()
	expiration := timeNow.Add(time.Hour)
	coins := sdk.NewCoins(sdk.NewCoin("steak", sdkmath.NewInt(10)))
	grant, err := authz.NewGrant(timeNow, banktypes.NewSendAuthorization(coins, nil), &expiration)
	suite.Require().NoError(err)

	_, err = suite.msgSrvr.Grant(suite.ctx, &authz.MsgGrant{
		Granter: addrs[0].String(),
		Grantee: addrs[1].String(),
		Grant:   grant,
	})
	suite.Require().NoError(err)

	_, err = suite.msgSrvr.Grant(suite.ctx, &authz.MsgGrant{
		Granter: addrs[1].String(),
		Grantee: addrs[0].String(),
		Grant:   grant,
	})
	suite.Require().NoError(err)

	totalGrants := 0
	suite.authzKeeper.IterateGrants(suite.ctx, func(sdk.AccAddress, sdk.AccAddress, authz.Grant) bool {
		totalGrants++
		return false
	})
	suite.Require().Equal(len(addrs), totalGrants)

	// prune expired grants
	headerInfo := suite.ctx.HeaderInfo()
	headerInfo.Time = headerInfo.Time.Add(2 * time.Hour)
	suite.ctx = suite.ctx.WithHeaderInfo(headerInfo)

	_, err = suite.authzKeeper.PruneExpiredGrants(suite.ctx, &authz.MsgPruneExpiredGrants{Pruner: addrs[0].String()})
	suite.Require().NoError(err)

	totalGrants = 0
	suite.authzKeeper.IterateGrants(suite.ctx, func(sdk.AccAddress, sdk.AccAddress, authz.Grant) bool {
		totalGrants++
		return false
	})
	suite.Require().Equal(0, totalGrants)
}
