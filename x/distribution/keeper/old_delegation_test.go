package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

func TestCalculateRewardsMultiDelegatorMultWithdraw(t *testing.T) {
	ctx, _, bk, k, sk, _ := CreateTestInputDefault(t, false, 1000)
	sh := staking.NewHandler(sk)
	initial := int64(20)

	// set module account coins
	distrAcc := k.GetDistributionAccount(ctx)
	err := bk.SetBalances(ctx, distrAcc.GetAddress(), sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1000))))
	require.NoError(t, err)
	k.supplyKeeper.SetModuleAccount(ctx, distrAcc)

	tokens := sdk.DecCoins{sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, sdk.NewDec(initial))}

	// create validator with 50% commission
	commission := staking.NewCommissionRates(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), sdk.NewDec(0))
	msg := staking.NewMsgCreateValidator(valOpAddr1, valConsPk1,
		sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100)), staking.Description{}, commission, sdk.OneInt())

	res, err := sh(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, res)

	// end block to bond validator
	staking.EndBlocker(ctx, sk)

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// fetch validator and delegation
	val := sk.Validator(ctx, valOpAddr1)
	del1 := sk.Delegation(ctx, sdk.AccAddress(valOpAddr1), valOpAddr1)

	// allocate some rewards
	k.AllocateTokensToValidator(ctx, val, tokens)

	// historical count should be 2 (validator init, delegation init)
	require.Equal(t, uint64(2), k.GetValidatorHistoricalReferenceCount(ctx))

	// second delegation
	msg2 := staking.NewMsgDelegate(sdk.AccAddress(valOpAddr2), valOpAddr1, sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100)))
	res, err = sh(ctx, msg2)
	require.NoError(t, err)
	require.NotNil(t, res)

	// historical count should be 3 (second delegation init)
	require.Equal(t, uint64(3), k.GetValidatorHistoricalReferenceCount(ctx))

	// fetch updated validator
	val = sk.Validator(ctx, valOpAddr1)
	del2 := sk.Delegation(ctx, sdk.AccAddress(valOpAddr2), valOpAddr1)

	// end block
	staking.EndBlocker(ctx, sk)

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// allocate some more rewards
	k.AllocateTokensToValidator(ctx, val, tokens)

	// first delegator withdraws
	k.WithdrawDelegationRewards(ctx, sdk.AccAddress(valOpAddr1), valOpAddr1)

	// second delegator withdraws
	k.WithdrawDelegationRewards(ctx, sdk.AccAddress(valOpAddr2), valOpAddr1)

	// historical count should be 3 (validator init + two delegations)
	require.Equal(t, uint64(3), k.GetValidatorHistoricalReferenceCount(ctx))

	// validator withdraws commission
	k.WithdrawValidatorCommission(ctx, valOpAddr1)

	// end period
	endingPeriod := k.IncrementValidatorPeriod(ctx, val)

	// calculate delegation rewards for del1
	rewards := k.CalculateDelegationRewards(ctx, val, del1, endingPeriod)

	// rewards for del1 should be zero
	require.True(t, rewards.IsZero())

	// calculate delegation rewards for del2
	rewards = k.CalculateDelegationRewards(ctx, val, del2, endingPeriod)

	// rewards for del2 should be zero
	require.True(t, rewards.IsZero())

	// commission should be zero
	require.True(t, k.GetValidatorAccumulatedCommission(ctx, valOpAddr1).Commission.IsZero())

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// allocate some more rewards
	k.AllocateTokensToValidator(ctx, val, tokens)

	// first delegator withdraws again
	k.WithdrawDelegationRewards(ctx, sdk.AccAddress(valOpAddr1), valOpAddr1)

	// end period
	endingPeriod = k.IncrementValidatorPeriod(ctx, val)

	// calculate delegation rewards for del1
	rewards = k.CalculateDelegationRewards(ctx, val, del1, endingPeriod)

	// rewards for del1 should be zero
	require.True(t, rewards.IsZero())

	// calculate delegation rewards for del2
	rewards = k.CalculateDelegationRewards(ctx, val, del2, endingPeriod)

	// rewards for del2 should be 1/4 initial
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDec(initial / 4)}}, rewards)

	// commission should be half initial
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDec(initial / 2)}}, k.GetValidatorAccumulatedCommission(ctx, valOpAddr1).Commission)

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// allocate some more rewards
	k.AllocateTokensToValidator(ctx, val, tokens)

	// withdraw commission
	k.WithdrawValidatorCommission(ctx, valOpAddr1)

	// end period
	endingPeriod = k.IncrementValidatorPeriod(ctx, val)

	// calculate delegation rewards for del1
	rewards = k.CalculateDelegationRewards(ctx, val, del1, endingPeriod)

	// rewards for del1 should be 1/4 initial
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDec(initial / 4)}}, rewards)

	// calculate delegation rewards for del2
	rewards = k.CalculateDelegationRewards(ctx, val, del2, endingPeriod)

	// rewards for del2 should be 1/2 initial
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDec(initial / 2)}}, rewards)

	// commission should be zero
	require.True(t, k.GetValidatorAccumulatedCommission(ctx, valOpAddr1).Commission.IsZero())
}
