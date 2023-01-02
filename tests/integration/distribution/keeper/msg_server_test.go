package keeper_test

import (
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	stakingtestutil "github.com/cosmos/cosmos-sdk/x/staking/testutil"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func (s *KeeperTestSuite) TestMsgUpdateParams() {
	// default params
	communityTax := sdk.NewDecWithPrec(2, 2) // 2%
	withdrawAddrEnabled := true

	testCases := []struct {
		name      string
		input     *types.MsgUpdateParams
		expErr    bool
		expErrMsg string
	}{
		{
			name: "invalid authority",
			input: &types.MsgUpdateParams{
				Authority: "invalid",
				Params: types.Params{
					CommunityTax:        sdk.NewDecWithPrec(2, 0),
					WithdrawAddrEnabled: withdrawAddrEnabled,
					BaseProposerReward:  sdk.ZeroDec(),
					BonusProposerReward: sdk.ZeroDec(),
				},
			},
			expErr:    true,
			expErrMsg: "invalid authority",
		},
		{
			name: "community tax > 1",
			input: &types.MsgUpdateParams{
				Authority: s.distrKeeper.GetAuthority(),
				Params: types.Params{
					CommunityTax:        sdk.NewDecWithPrec(2, 0),
					WithdrawAddrEnabled: withdrawAddrEnabled,
					BaseProposerReward:  sdk.ZeroDec(),
					BonusProposerReward: sdk.ZeroDec(),
				},
			},
			expErr:    true,
			expErrMsg: "community tax should be non-negative and less than one",
		},
		{
			name: "negative community tax",
			input: &types.MsgUpdateParams{
				Authority: s.distrKeeper.GetAuthority(),
				Params: types.Params{
					CommunityTax:        sdk.NewDecWithPrec(-2, 1),
					WithdrawAddrEnabled: withdrawAddrEnabled,
					BaseProposerReward:  sdk.ZeroDec(),
					BonusProposerReward: sdk.ZeroDec(),
				},
			},
			expErr:    true,
			expErrMsg: "community tax should be non-negative and less than one",
		},
		{
			name: "base proposer reward set",
			input: &types.MsgUpdateParams{
				Authority: s.distrKeeper.GetAuthority(),
				Params: types.Params{
					CommunityTax:        communityTax,
					BaseProposerReward:  sdk.NewDecWithPrec(1, 2),
					BonusProposerReward: sdk.ZeroDec(),
					WithdrawAddrEnabled: withdrawAddrEnabled,
				},
			},
			expErr:    true,
			expErrMsg: "cannot update base or bonus proposer reward because these are deprecated fields: invalid request",
		},
		{
			name: "bonus proposer reward set",
			input: &types.MsgUpdateParams{
				Authority: s.distrKeeper.GetAuthority(),
				Params: types.Params{
					CommunityTax:        communityTax,
					BaseProposerReward:  sdk.ZeroDec(),
					BonusProposerReward: sdk.NewDecWithPrec(1, 2),
					WithdrawAddrEnabled: withdrawAddrEnabled,
				},
			},
			expErr:    true,
			expErrMsg: "cannot update base or bonus proposer reward because these are deprecated fields: invalid request",
		},
		{
			name: "all good",
			input: &types.MsgUpdateParams{
				Authority: s.distrKeeper.GetAuthority(),
				Params: types.Params{
					CommunityTax:        communityTax,
					BaseProposerReward:  sdk.ZeroDec(),
					BonusProposerReward: sdk.ZeroDec(),
					WithdrawAddrEnabled: withdrawAddrEnabled,
				},
			},
			expErr: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			_, err := s.msgServer.UpdateParams(s.ctx, tc.input)

			if tc.expErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

func (s *KeeperTestSuite) TestCommunityPoolSpend() {
	recipient := sdk.AccAddress([]byte("addr1_______________"))

	testCases := []struct {
		name      string
		input     *types.MsgCommunityPoolSpend
		expErr    bool
		expErrMsg string
	}{
		{
			name: "invalid authority",
			input: &types.MsgCommunityPoolSpend{
				Authority: "invalid",
				Recipient: recipient.String(),
				Amount:    sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(100))),
			},
			expErr:    true,
			expErrMsg: "invalid authority",
		},
		{
			name: "invalid recipient",
			input: &types.MsgCommunityPoolSpend{
				Authority: s.distrKeeper.GetAuthority(),
				Recipient: "invalid",
				Amount:    sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(100))),
			},
			expErr:    true,
			expErrMsg: "decoding bech32 failed",
		},
		{
			name: "valid message",
			input: &types.MsgCommunityPoolSpend{
				Authority: s.distrKeeper.GetAuthority(),
				Recipient: recipient.String(),
				Amount:    sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(100))),
			},
			expErr: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			_, err := s.msgServer.CommunityPoolSpend(s.ctx, tc.input)

			if tc.expErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				s.Require().NoError(err)

				r, err := sdk.AccAddressFromBech32(tc.input.Recipient)
				s.Require().NoError(err)

				b := s.bankKeeper.GetAllBalances(s.ctx, r)
				s.Require().False(b.IsZero())
			}
		})
	}
}

func (s *KeeperTestSuite) TestMsgDepositValidatorRewardsPool() {
	tstaking := stakingtestutil.NewHelper(s.T(), s.ctx, s.stakingKeeper)
	tstaking.Commission = stakingtypes.NewCommissionRates(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), math.LegacyNewDec(0))
	tstaking.CreateValidator(s.valAddrs[1], valConsPk0, sdk.NewInt(100), true)

	// mint a non-staking token and send to an account
	amt := sdk.NewCoins(sdk.NewInt64Coin("foo", 500))
	s.bankKeeper.MintCoins(s.ctx, minttypes.ModuleName, amt)
	s.bankKeeper.SendCoinsFromModuleToAccount(s.ctx, minttypes.ModuleName, s.addrs[0], amt)

	testCases := []struct {
		name      string
		input     *types.MsgDepositValidatorRewardsPool
		expErr    bool
		expErrMsg string
	}{
		{
			name: "happy path (staking token)",
			input: &types.MsgDepositValidatorRewardsPool{
				Authority:        s.addrs[0].String(),
				ValidatorAddress: s.valAddrs[1].String(),
				Amount:           sdk.NewCoins(sdk.NewCoin(s.stakingKeeper.BondDenom(s.ctx), sdk.NewInt(100))),
			},
		},
		{
			name: "happy path (non-staking token)",
			input: &types.MsgDepositValidatorRewardsPool{
				Authority:        s.addrs[0].String(),
				ValidatorAddress: s.valAddrs[1].String(),
				Amount:           amt,
			},
		},
		{
			name: "invalid validator",
			input: &types.MsgDepositValidatorRewardsPool{
				Authority:        s.addrs[0].String(),
				ValidatorAddress: sdk.ValAddress([]byte("addr1_______________")).String(),
				Amount:           sdk.NewCoins(sdk.NewCoin(s.stakingKeeper.BondDenom(s.ctx), sdk.NewInt(100))),
			},
			expErr:    true,
			expErrMsg: "validator does not exist",
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			_, err := s.msgServer.DepositValidatorRewardsPool(s.ctx, tc.input)

			if tc.expErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				s.Require().NoError(err)

				valAddr, err := sdk.ValAddressFromBech32(tc.input.ValidatorAddress)
				s.Require().NoError(err)

				// check validator outstanding rewards
				outstandingRewards := s.distrKeeper.GetValidatorOutstandingRewards(s.ctx, valAddr)
				for _, c := range tc.input.Amount {
					x := outstandingRewards.Rewards.AmountOf(c.Denom)
					s.Require().Equal(x, sdk.NewDecFromInt(c.Amount))
				}
			}
		})
	}
}
