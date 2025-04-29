package keeper_test

import (
	"time"

	"go.uber.org/mock/gomock"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec/address"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/protocolpool/types"
)

func (suite *KeeperTestSuite) TestFundCommunityPool() {
	validDepositor := recipientAddr
	validAmount := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(1000)))

	testCases := []struct {
		name      string
		preRun    func()
		msg       *types.MsgFundCommunityPool
		expErr    bool
		expErrMsg string
	}{
		{
			name: "invalid depositor address",
			msg: &types.MsgFundCommunityPool{
				Depositor: "invalid",
				Amount:    validAmount,
			},
			preRun: func() {
				suite.authKeeper.EXPECT().AddressCodec().
					Return(address.NewBech32Codec("cosmos")).AnyTimes()
			},
			expErr:    true,
			expErrMsg: "invalid depositor address:",
		},
		{
			name: "invalid amount",
			msg: &types.MsgFundCommunityPool{
				Depositor: validDepositor.String(),
				Amount:    sdk.Coins{sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: math.NewInt(-1)}},
			},
			preRun: func() {
				suite.authKeeper.EXPECT().AddressCodec().
					Return(address.NewBech32Codec("cosmos")).AnyTimes()
			},
			expErr:    true,
			expErrMsg: "-1stake: invalid coins",
		},
		{
			name: "valid fund community pool",
			msg: &types.MsgFundCommunityPool{
				Depositor: validDepositor.String(),
				Amount:    validAmount,
			},
			preRun: func() {
				suite.authKeeper.EXPECT().AddressCodec().
					Return(address.NewBech32Codec("cosmos")).AnyTimes()
				suite.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), validDepositor, types.ModuleName, validAmount).Return(nil).Times(1)
			},
			expErr: false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			if tc.preRun != nil {
				tc.preRun()
			}
			resp, err := suite.msgServer.FundCommunityPool(suite.ctx, tc.msg)
			if tc.expErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				suite.Require().NoError(err)
				suite.Require().NotNil(resp)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestCommunityPoolSpend() {
	validAuthority := suite.poolKeeper.GetAuthority()
	validRecipient := recipientAddr
	validAmount := sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(500)))

	testCases := []struct {
		name      string
		preRun    func()
		msg       *types.MsgCommunityPoolSpend
		expErr    bool
		expErrMsg string
	}{
		{
			name: "invalid authority",
			msg: &types.MsgCommunityPoolSpend{
				Authority: "invalid_auth",
				Recipient: validRecipient.String(),
				Amount:    validAmount,
			},
			preRun:    nil,
			expErr:    true,
			expErrMsg: "invalid authority address",
		},
		{
			name: "invalid amount",
			msg: &types.MsgCommunityPoolSpend{
				Authority: validAuthority,
				Recipient: validRecipient.String(),
				Amount:    sdk.Coins{sdk.Coin{Denom: "stake", Amount: math.NewInt(-1)}},
			},
			preRun: func() {
				suite.authKeeper.EXPECT().AddressCodec().
					Return(address.NewBech32Codec("cosmos")).AnyTimes()
			},
			expErr:    true,
			expErrMsg: "-1stake: invalid coins",
		},
		{
			name: "invalid recipient address",
			msg: &types.MsgCommunityPoolSpend{
				Authority: validAuthority,
				Recipient: "invalid",
				Amount:    validAmount,
			},
			preRun: func() {
				suite.authKeeper.EXPECT().AddressCodec().
					Return(address.NewBech32Codec("cosmos")).AnyTimes()
			},
			expErr:    true,
			expErrMsg: "decoding bech32 failed",
		},
		{
			name: "valid community pool spend",
			msg: &types.MsgCommunityPoolSpend{
				Authority: validAuthority,
				Recipient: validRecipient.String(),
				Amount:    validAmount,
			},
			preRun: func() {
				suite.authKeeper.EXPECT().AddressCodec().
					Return(address.NewBech32Codec("cosmos")).AnyTimes()
				suite.bankKeeper.EXPECT().SendCoinsFromModuleToAccount(gomock.Any(), types.ModuleName, validRecipient, validAmount).Return(nil).Times(1)
			},
			expErr: false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			if tc.preRun != nil {
				tc.preRun()
			}
			resp, err := suite.msgServer.CommunityPoolSpend(suite.ctx, tc.msg)
			if tc.expErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				suite.Require().NoError(err)
				suite.Require().NotNil(resp)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestCreateContinuousFund() {
	validAuthority := suite.poolKeeper.GetAuthority()
	validRecipient := recipientAddr
	validPercentage := math.LegacyMustNewDecFromStr("0.2")
	validExpiry := suite.ctx.BlockTime().Add(24 * time.Hour)

	testCases := []struct {
		name      string
		preRun    func()
		msg       *types.MsgCreateContinuousFund
		expErr    bool
		expErrMsg string
		verify    func(msg *types.MsgCreateContinuousFund)
	}{
		{
			name: "invalid authority",
			msg: &types.MsgCreateContinuousFund{
				Authority:  "invalid_auth",
				Recipient:  validRecipient.String(),
				Percentage: validPercentage,
				Expiry:     &validExpiry,
			},
			preRun:    func() {},
			expErr:    true,
			expErrMsg: "invalid authority address",
		},
		{
			name: "invalid recipient address",
			msg: &types.MsgCreateContinuousFund{
				Authority:  validAuthority,
				Recipient:  "invalid",
				Percentage: validPercentage,
				Expiry:     &validExpiry,
			},
			preRun: func() {
				suite.authKeeper.EXPECT().AddressCodec().
					Return(address.NewBech32Codec("cosmos")).AnyTimes()
			},
			expErr:    true,
			expErrMsg: "decoding bech32 failed",
		},
		{
			name: "continuous fund already exists",
			msg: &types.MsgCreateContinuousFund{
				Authority:  validAuthority,
				Recipient:  validRecipient.String(),
				Percentage: validPercentage,
				Expiry:     &validExpiry,
			},
			preRun: func() {
				suite.authKeeper.EXPECT().AddressCodec().
					Return(address.NewBech32Codec("cosmos")).AnyTimes()
				suite.bankKeeper.EXPECT().BlockedAddr(validRecipient).Return(false).Times(1)
				// Pre-create a continuous fund.
				err := suite.poolKeeper.ContinuousFunds.Set(suite.ctx, validRecipient, types.ContinuousFund{
					Recipient:  validRecipient.String(),
					Percentage: validPercentage,
					Expiry:     &validExpiry,
				})
				suite.Require().NoError(err)
			},
			expErr:    true,
			expErrMsg: "continuous fund already exists",
		},
		{
			name: "invalid continuous fund fields",
			msg: &types.MsgCreateContinuousFund{
				Authority:  validAuthority,
				Recipient:  validRecipient.String(),
				Percentage: math.LegacyZeroDec(), // zero percent is invalid
				Expiry:     &validExpiry,
			},
			preRun: func() {
				suite.authKeeper.EXPECT().AddressCodec().
					Return(address.NewBech32Codec("cosmos")).AnyTimes()
				suite.bankKeeper.EXPECT().BlockedAddr(validRecipient).Return(false).Times(1)
			},
			expErr:    true,
			expErrMsg: "invalid continuous fund",
		},
		{
			name: "total percentage exceeds 100%",
			msg: &types.MsgCreateContinuousFund{
				Authority: validAuthority,
				Recipient: validRecipient.String(),
				// Set a high percentage so that total exceeds 1 when added to an existing fund.
				Percentage: math.LegacyMustNewDecFromStr("0.9"),
				Expiry:     &validExpiry,
			},
			preRun: func() {
				suite.authKeeper.EXPECT().AddressCodec().
					Return(address.NewBech32Codec("cosmos")).AnyTimes()
				suite.bankKeeper.EXPECT().BlockedAddr(validRecipient).Return(false).Times(1)

				existingRecipient := recipientAddr2
				err := suite.poolKeeper.ContinuousFunds.Set(suite.ctx, existingRecipient, types.ContinuousFund{
					Recipient:  existingRecipient.String(),
					Percentage: math.LegacyMustNewDecFromStr("0.2"), // total will become 1.1
					Expiry:     nil,
				})
				suite.Require().NoError(err)
			},
			expErr:    true,
			expErrMsg: "total funds percentage exceeds 100",
		},
		{
			name: "address is bocked",
			msg: &types.MsgCreateContinuousFund{
				Authority:  validAuthority,
				Recipient:  validRecipient.String(),
				Percentage: validPercentage,
				Expiry:     &validExpiry,
			},
			preRun: func() {
				suite.authKeeper.EXPECT().AddressCodec().
					Return(address.NewBech32Codec("cosmos")).AnyTimes()
				suite.bankKeeper.EXPECT().BlockedAddr(validRecipient).Return(true).Times(1)

				// Ensure any existing fund for validRecipient is removed.
				_ = suite.poolKeeper.ContinuousFunds.Remove(suite.ctx, validRecipient)
			},
			expErr:    true,
			expErrMsg: "recipient is blocked in the bank keeper",
		},
		{
			name: "valid create continuous fund",
			msg: &types.MsgCreateContinuousFund{
				Authority:  validAuthority,
				Recipient:  validRecipient.String(),
				Percentage: validPercentage,
				Expiry:     &validExpiry,
			},
			preRun: func() {
				suite.authKeeper.EXPECT().AddressCodec().
					Return(address.NewBech32Codec("cosmos")).AnyTimes()
				suite.bankKeeper.EXPECT().BlockedAddr(validRecipient).Return(false).Times(1)
				// Ensure any existing fund for validRecipient is removed.
				_ = suite.poolKeeper.ContinuousFunds.Remove(suite.ctx, validRecipient)
			},
			expErr: false,
		},
	}

	for _, tc := range testCases {
		suite.SetupTest()
		suite.Run(tc.name, func() {
			if tc.preRun != nil {
				tc.preRun()
			}
			resp, err := suite.msgServer.CreateContinuousFund(suite.ctx, tc.msg)
			if tc.expErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				suite.Require().NoError(err)
				suite.Require().NotNil(resp)
				// Verify that the fund was stored.
				fund, err := suite.poolKeeper.ContinuousFunds.Get(suite.ctx, sdk.MustAccAddressFromBech32(tc.msg.Recipient))
				suite.Require().NoError(err)
				suite.Require().Equal(tc.msg.Recipient, fund.Recipient)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestCancelContinuousFund() {
	validAuthority := suite.poolKeeper.GetAuthority()
	validRecipient := recipientAddr

	testCases := []struct {
		name      string
		preRun    func()
		msg       *types.MsgCancelContinuousFund
		expErr    bool
		expErrMsg string
		verify    func(msg *types.MsgCancelContinuousFund)
	}{
		{
			name: "invalid authority",
			msg: &types.MsgCancelContinuousFund{
				Authority: "invalid_auth",
				Recipient: validRecipient.String(),
			},
			preRun:    func() {},
			expErr:    true,
			expErrMsg: "invalid authority address",
		},
		{
			name: "invalid recipient address",
			msg: &types.MsgCancelContinuousFund{
				Authority: validAuthority,
				Recipient: "invalid",
			},
			preRun: func() {
				suite.authKeeper.EXPECT().AddressCodec().
					Return(address.NewBech32Codec("cosmos")).AnyTimes()
			},
			expErr:    true,
			expErrMsg: "decoding bech32 failed:",
		},
		{
			name: "remove a continuous fund that already was removed - error does not exist",
			msg: &types.MsgCancelContinuousFund{
				Authority: validAuthority,
				Recipient: validRecipient.String(),
			},
			preRun: func() {
				suite.authKeeper.EXPECT().AddressCodec().
					Return(address.NewBech32Codec("cosmos")).AnyTimes()
				// Ensure the continuous fund is not set so that Remove fails.
				_ = suite.poolKeeper.ContinuousFunds.Remove(suite.ctx, validRecipient)
			},
			expErr: true,
		},
		{
			name: "valid cancel continuous fund",
			msg: &types.MsgCancelContinuousFund{
				Authority: validAuthority,
				Recipient: validRecipient.String(),
			},
			preRun: func() {
				suite.authKeeper.EXPECT().AddressCodec().
					Return(address.NewBech32Codec("cosmos")).AnyTimes()
				fund := types.ContinuousFund{
					Recipient:  validRecipient.String(),
					Percentage: math.LegacyMustNewDecFromStr("0.3"),
					Expiry:     nil,
				}
				err := suite.poolKeeper.ContinuousFunds.Set(suite.ctx, validRecipient, fund)
				suite.Require().NoError(err)
			},
			expErr: false,
			verify: func(msg *types.MsgCancelContinuousFund) {
				// Verify that the fund has been removed.
				_, err := suite.poolKeeper.ContinuousFunds.Get(suite.ctx, validRecipient)
				suite.Require().Error(err, "expected error when retrieving removed fund")
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			if tc.preRun != nil {
				tc.preRun()
			}
			resp, err := suite.msgServer.CancelContinuousFund(suite.ctx, tc.msg)
			if tc.expErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				suite.Require().NoError(err)
				suite.Require().NotNil(resp)
				suite.Require().Equal(uint64(suite.ctx.BlockHeight()), resp.CanceledHeight)
				suite.Require().Equal(tc.msg.Recipient, resp.Recipient)
				if tc.verify != nil {
					tc.verify(tc.msg)
				}
			}
		})
	}
}

func (suite *KeeperTestSuite) TestUpdateParams() {
	validAuthority := suite.poolKeeper.GetAuthority()

	testCases := []struct {
		name      string
		msg       *types.MsgUpdateParams
		expErr    bool
		expErrMsg string
	}{
		{
			name: "invalid authority",
			msg: &types.MsgUpdateParams{
				Authority: "invalid_auth",
				Params:    types.Params{EnabledDistributionDenoms: []string{sdk.DefaultBondDenom}},
			},
			expErr:    true,
			expErrMsg: "invalid authority address",
		},
		{
			name: "error setting params (invalid params)",
			msg: &types.MsgUpdateParams{
				Authority: validAuthority,
				Params:    types.Params{EnabledDistributionDenoms: []string{sdk.DefaultBondDenom}, DistributionFrequency: 0},
			},
			expErr:    true,
			expErrMsg: "invalid params",
		},
		{
			name: "valid update params",
			msg: &types.MsgUpdateParams{
				Authority: validAuthority,
				Params:    types.DefaultParams(),
			},
			expErr: false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			resp, err := suite.msgServer.UpdateParams(suite.ctx, tc.msg)
			if tc.expErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				suite.Require().NoError(err)
				suite.Require().NotNil(resp)
			}
		})
	}
}
