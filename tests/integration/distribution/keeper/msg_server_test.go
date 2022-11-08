package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	"github.com/cosmos/cosmos-sdk/x/distribution/testutil"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtestutil "github.com/cosmos/cosmos-sdk/x/staking/testutil"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

func (s *KeeperTestSuite) TestMsgUpdateParams() {
	// default params
	communityTax := sdk.NewDecWithPrec(2, 2)        // 2%
	baseProposerReward := sdk.NewDecWithPrec(1, 2)  // 1%
	bonusProposerReward := sdk.NewDecWithPrec(4, 2) // 4%
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
					BaseProposerReward:  baseProposerReward,
					BonusProposerReward: bonusProposerReward,
					WithdrawAddrEnabled: withdrawAddrEnabled,
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
					BaseProposerReward:  baseProposerReward,
					BonusProposerReward: bonusProposerReward,
					WithdrawAddrEnabled: withdrawAddrEnabled,
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
					BaseProposerReward:  baseProposerReward,
					BonusProposerReward: bonusProposerReward,
					WithdrawAddrEnabled: withdrawAddrEnabled,
				},
			},
			expErr:    true,
			expErrMsg: "community tax should be non-negative and less than one",
		},
		{
			name: "base proposer reward > 1",
			input: &types.MsgUpdateParams{
				Authority: s.distrKeeper.GetAuthority(),
				Params: types.Params{
					CommunityTax:        communityTax,
					BaseProposerReward:  sdk.NewDecWithPrec(2, 0),
					BonusProposerReward: bonusProposerReward,
					WithdrawAddrEnabled: withdrawAddrEnabled,
				},
			},
			expErr:    true,
			expErrMsg: "sum of base, bonus proposer rewards, and community tax cannot be greater than one",
		},
		{
			name: "negative base proposer reward",
			input: &types.MsgUpdateParams{
				Authority: s.distrKeeper.GetAuthority(),
				Params: types.Params{
					CommunityTax:        communityTax,
					BaseProposerReward:  sdk.NewDecWithPrec(-2, 0),
					BonusProposerReward: bonusProposerReward,
					WithdrawAddrEnabled: withdrawAddrEnabled,
				},
			},
			expErr:    true,
			expErrMsg: "base proposer reward should be positive",
		},
		{
			name: "bonus proposer reward > 1",
			input: &types.MsgUpdateParams{
				Authority: s.distrKeeper.GetAuthority(),
				Params: types.Params{
					CommunityTax:        communityTax,
					BaseProposerReward:  baseProposerReward,
					BonusProposerReward: sdk.NewDecWithPrec(2, 0),
					WithdrawAddrEnabled: withdrawAddrEnabled,
				},
			},
			expErr:    true,
			expErrMsg: "sum of base, bonus proposer rewards, and community tax cannot be greater than one",
		},
		{
			name: "negative bonus proposer reward",
			input: &types.MsgUpdateParams{
				Authority: s.distrKeeper.GetAuthority(),
				Params: types.Params{
					CommunityTax:        communityTax,
					BaseProposerReward:  baseProposerReward,
					BonusProposerReward: sdk.NewDecWithPrec(-2, 0),
					WithdrawAddrEnabled: withdrawAddrEnabled,
				},
			},
			expErr:    true,
			expErrMsg: "bonus proposer reward should be positive",
		},
		{
			name: "all good",
			input: &types.MsgUpdateParams{
				Authority: s.distrKeeper.GetAuthority(),
				Params: types.Params{
					CommunityTax:        communityTax,
					BaseProposerReward:  baseProposerReward,
					BonusProposerReward: bonusProposerReward,
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

func TestDepositValidatorRewardsPool(t *testing.T) {
	var (
		bankKeeper    bankkeeper.Keeper
		distrKeeper   keeper.Keeper
		stakingKeeper *stakingkeeper.Keeper
	)

	app, err := simtestutil.Setup(testutil.AppConfig,
		&bankKeeper,
		&distrKeeper,
		&stakingKeeper,
	)
	require.NoError(t, err)

	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	addrs := simtestutil.AddTestAddrs(bankKeeper, stakingKeeper, ctx, 2, sdk.NewInt(1000))
	valAddrs := simtestutil.ConvertAddrsToValAddrs(addrs)

	addrVal1 := valAddrs[0]

	tstaking := stakingtestutil.NewHelper(t, ctx, stakingKeeper)

	// Create Validators and Delegation
	tstaking.Commission = stakingtypes.NewCommissionRates(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), math.LegacyNewDec(0))
	tstaking.CreateValidator(valAddrs[0], valConsPk0, sdk.NewInt(100), true)

	coins := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000_000))
	err = bankKeeper.MintCoins(ctx, minttypes.ModuleName, coins)
	require.NoError(t, err)
	err = bankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, addrs[0], coins)
	require.NoError(t, err)

	msgServer := keeper.NewMsgServerImpl(distrKeeper)
	_, err = msgServer.DepositValidatorRewardsPool(sdk.WrapSDKContext(ctx), &types.MsgDepositValidatorRewardsPool{
		Authority:        addrs[0].String(),
		ValidatorAddress: addrVal1.String(),
		Amount:           coins,
	})
	require.NoError(t, err)

	// check validator outstanding rewards
	outstandingRewards := distrKeeper.GetValidatorOutstandingRewards(ctx, addrVal1)
	require.Equal(t, outstandingRewards.Rewards, sdk.NewDecCoinsFromCoins(coins...))
}
