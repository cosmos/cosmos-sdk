package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	"gotest.tools/v3/assert"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	stakingtestutil "github.com/cosmos/cosmos-sdk/x/staking/testutil"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
)

var (
	emptyDelAddr sdk.AccAddress
	emptyValAddr sdk.ValAddress
)

func TestMsgSetWithdrawAddress(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	tests := []struct {
		name          string
		delegatorAddr sdk.AccAddress
		withdrawAddr  sdk.AccAddress
		expErr        bool
		expErrMsg     string
	}{
		{
			name:          "valid case",
			delegatorAddr: f.addrs[0],
			withdrawAddr:  f.addrs[1],
			expErr:        false,
		},
		{
			name:          "valid case, same delegator and withdraw address",
			delegatorAddr: f.addrs[0],
			withdrawAddr:  f.addrs[0],
			expErr:        false,
		},
		{
			name:          "empty delegator address",
			delegatorAddr: emptyDelAddr,
			withdrawAddr:  f.addrs[0],
			expErr:        true,
			expErrMsg:     "invalid delegator address",
		},
		{
			name:          "empty withdraw address",
			delegatorAddr: f.addrs[0],
			withdrawAddr:  emptyDelAddr,
			expErr:        true,
			expErrMsg:     "invalid withdraw address",
		},
		{
			name:          "both empty addresses",
			delegatorAddr: emptyDelAddr,
			withdrawAddr:  emptyDelAddr,
			expErr:        true,
			expErrMsg:     "invalid delegator address",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			msg := types.NewMsgSetWithdrawAddress(tc.delegatorAddr, tc.withdrawAddr)
			_, err := f.msgServer.SetWithdrawAddress(f.ctx, msg)
			if tc.expErr {
				require.ErrorContains(t, err, tc.expErrMsg)
			} else {
				require.Nil(t, err)
			}
		})
	}
}

func TestMsgWithdrawDelegatorReward(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	tests := []struct {
		name          string
		delegatorAddr sdk.AccAddress
		validatorAddr sdk.ValAddress
		expErr        bool
		expErrMsg     string
	}{
		{
			name:          "empty delegator address",
			delegatorAddr: emptyDelAddr,
			validatorAddr: f.valAddrs[0],
			expErr:        true,
			expErrMsg:     "invalid delegator address",
		},
		{
			name:          "empty validator address",
			delegatorAddr: f.addrs[0],
			validatorAddr: emptyValAddr,
			expErr:        true,
			expErrMsg:     "invalid validator address",
		},
		{
			name:          "both empty addresses",
			delegatorAddr: emptyDelAddr,
			validatorAddr: emptyValAddr,
			expErr:        true,
			expErrMsg:     "invalid validator address",
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			msg := types.NewMsgWithdrawDelegatorReward(tc.delegatorAddr, tc.validatorAddr)
			_, err := f.msgServer.WithdrawDelegatorReward(f.ctx, msg)
			if tc.expErr {
				require.ErrorContains(t, err, tc.expErrMsg)
			} else {
				require.Nil(t, err)
			}
		})
	}
}

func TestMsgWithdrawValidatorCommission(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	tests := []struct {
		name          string
		validatorAddr sdk.ValAddress
		expErr        bool
		expErrMsg     string
	}{
		{
			name:          "valid withdraw (but validator has no commission)",
			validatorAddr: f.valAddrs[0],
			expErr:        true,
			expErrMsg:     "no validator commission to withdraw",
		},
		{
			name:          "empty validator address",
			validatorAddr: emptyValAddr,
			expErr:        true,
			expErrMsg:     "invalid validator address",
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			msg := types.NewMsgWithdrawValidatorCommission(tc.validatorAddr)
			_, err := f.msgServer.WithdrawValidatorCommission(f.ctx, msg)
			if tc.expErr {
				require.ErrorContains(t, err, tc.expErrMsg)
			} else {
				require.Nil(t, err)
			}
		})
	}
}

func TestMsgFundCommunityPool(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	tests := []struct {
		name      string
		amount    sdk.Coins
		depositor sdk.AccAddress
		expErr    bool
		expErrMsg string
	}{
		{
			name:      "no depositor",
			amount:    sdk.NewCoins(sdk.NewInt64Coin("stake", 10000)),
			depositor: sdk.AccAddress{},
			expErr:    true,
			expErrMsg: "invalid depositor address",
		},
		{
			name:      "invalid coin",
			amount:    sdk.Coins{sdk.NewInt64Coin("stake", 10), sdk.NewInt64Coin("stake", 10)},
			depositor: f.addrs[0],
			expErr:    true,
			expErrMsg: "10stake,10stake: invalid coins",
		},
		{
			name:      "valid deposit",
			amount:    sdk.NewCoins(sdk.NewInt64Coin("stake", 1000)),
			depositor: f.addrs[0],
			expErr:    false,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			msg := types.NewMsgFundCommunityPool(tc.amount, tc.depositor)
			_, err := f.msgServer.FundCommunityPool(f.ctx, msg)
			if tc.expErr {
				require.ErrorContains(t, err, tc.expErrMsg)
			} else {
				require.Nil(t, err)
			}
		})
	}
}

func TestMsgUpdateParams(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

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
				Authority: f.distrKeeper.GetAuthority(),
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
				Authority: f.distrKeeper.GetAuthority(),
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
				Authority: f.distrKeeper.GetAuthority(),
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
				Authority: f.distrKeeper.GetAuthority(),
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
				Authority: f.distrKeeper.GetAuthority(),
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
		t.Run(tc.name, func(t *testing.T) {
			_, err := f.msgServer.UpdateParams(f.ctx, tc.input)
			if tc.expErr {
				assert.ErrorContains(t, err, tc.expErrMsg)
			} else {
				assert.NilError(t, err)
			}
		})
	}
}

func TestCommunityPoolSpend(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

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
				Authority: f.distrKeeper.GetAuthority(),
				Recipient: "invalid",
				Amount:    sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(100))),
			},
			expErr:    true,
			expErrMsg: "decoding bech32 failed",
		},
		{
			name: "valid message",
			input: &types.MsgCommunityPoolSpend{
				Authority: f.distrKeeper.GetAuthority(),
				Recipient: recipient.String(),
				Amount:    sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(100))),
			},
			expErr: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			_, err := f.msgServer.CommunityPoolSpend(f.ctx, tc.input)

			if tc.expErr {
				assert.ErrorContains(t, err, tc.expErrMsg)
			} else {
				assert.NilError(t, err)
				r, err := sdk.AccAddressFromBech32(tc.input.Recipient)
				assert.NilError(t, err)

				b := f.bankKeeper.GetAllBalances(f.ctx, r)

				assert.Assert(t, b.IsZero() == false)
			}
		})
	}
}

func TestMsgDepositValidatorRewardsPool(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	tstaking := stakingtestutil.NewHelper(t, f.ctx, f.stakingKeeper)
	tstaking.Commission = stakingtypes.NewCommissionRates(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), math.LegacyNewDec(0))
	tstaking.CreateValidator(f.valAddrs[1], valConsPk0, sdk.NewInt(100), true)

	// mint a non-staking token and send to an account
	amt := sdk.NewCoins(sdk.NewInt64Coin("foo", 500))
	f.bankKeeper.MintCoins(f.ctx, minttypes.ModuleName, amt)
	f.bankKeeper.SendCoinsFromModuleToAccount(f.ctx, minttypes.ModuleName, f.addrs[0], amt)

	testCases := []struct {
		name      string
		input     *types.MsgDepositValidatorRewardsPool
		expErr    bool
		expErrMsg string
	}{
		{
			name: "happy path (staking token)",
			input: &types.MsgDepositValidatorRewardsPool{
				Depositor:        f.addrs[0].String(),
				ValidatorAddress: f.valAddrs[1].String(),
				Amount:           sdk.NewCoins(sdk.NewCoin(f.stakingKeeper.BondDenom(f.ctx), sdk.NewInt(100))),
			},
		},
		{
			name: "happy path (non-staking token)",
			input: &types.MsgDepositValidatorRewardsPool{
				Depositor:        f.addrs[0].String(),
				ValidatorAddress: f.valAddrs[1].String(),
				Amount:           amt,
			},
		},
		{
			name: "invalid validator",
			input: &types.MsgDepositValidatorRewardsPool{
				Depositor:        f.addrs[0].String(),
				ValidatorAddress: sdk.ValAddress([]byte("addr1_______________")).String(),
				Amount:           sdk.NewCoins(sdk.NewCoin(f.stakingKeeper.BondDenom(f.ctx), sdk.NewInt(100))),
			},
			expErr:    true,
			expErrMsg: "validator does not exist",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			_, err := f.msgServer.DepositValidatorRewardsPool(f.ctx, tc.input)

			if tc.expErr {
				assert.ErrorContains(t, err, tc.expErrMsg)
			} else {
				assert.NilError(t, err)

				valAddr, err := sdk.ValAddressFromBech32(tc.input.ValidatorAddress)
				assert.NilError(t, err)

				// check validator outstanding rewards
				outstandingRewards := f.distrKeeper.GetValidatorOutstandingRewards(f.ctx, valAddr)
				for _, c := range tc.input.Amount {
					x := outstandingRewards.Rewards.AmountOf(c.Denom)
					assert.DeepEqual(t, x, sdk.NewDecFromInt(c.Amount))
				}
			}
		})
	}
}
