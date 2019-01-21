package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

func TestCalculateRewardsBasic(t *testing.T) {
	ctx, _, k, sk, _ := CreateTestInputDefault(t, false, 1000)
	sh := staking.NewHandler(sk)

	// initialize state
	k.SetOutstandingRewards(ctx, sdk.DecCoins{})

	// create validator with 50% commission
	commission := staking.NewCommissionMsg(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), sdk.NewDec(0))
	msg := staking.NewMsgCreateValidator(valOpAddr1, valConsPk1,
		sdk.NewCoin(staking.DefaultBondDenom, sdk.NewInt(100)), staking.Description{}, commission)
	require.True(t, sh(ctx, msg).IsOK())

	// end block to bond validator
	staking.EndBlocker(ctx, sk)

	// fetch validator and delegation
	val := sk.Validator(ctx, valOpAddr1)
	del := sk.Delegation(ctx, sdk.AccAddress(valOpAddr1), valOpAddr1)

	// end period
	endingPeriod := k.incrementValidatorPeriod(ctx, val)

	// calculate delegation rewards
	rewards := k.calculateDelegationRewards(ctx, val, del, endingPeriod)

	// rewards should be zero
	require.True(t, rewards.IsZero())

	// allocate some rewards
	initial := int64(10)
	tokens := sdk.DecCoins{{staking.DefaultBondDenom, sdk.NewDec(initial)}}
	k.AllocateTokensToValidator(ctx, val, tokens)

	// end period
	endingPeriod = k.incrementValidatorPeriod(ctx, val)

	// calculate delegation rewards
	rewards = k.calculateDelegationRewards(ctx, val, del, endingPeriod)

	// rewards should be half the tokens
	require.Equal(t, sdk.DecCoins{{staking.DefaultBondDenom, sdk.NewDec(initial / 2)}}, rewards)

	// commission should be the other half
	require.Equal(t, sdk.DecCoins{{staking.DefaultBondDenom, sdk.NewDec(initial / 2)}}, k.GetValidatorAccumulatedCommission(ctx, valOpAddr1))
}

func TestCalculateRewardsAfterSlash(t *testing.T) {
	ctx, _, k, sk, _ := CreateTestInputDefault(t, false, 1000)
	sh := staking.NewHandler(sk)

	// initialize state
	k.SetOutstandingRewards(ctx, sdk.DecCoins{})

	// create validator with 50% commission
	commission := staking.NewCommissionMsg(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), sdk.NewDec(0))
	msg := staking.NewMsgCreateValidator(valOpAddr1, valConsPk1,
		sdk.NewCoin(staking.DefaultBondDenom, sdk.NewInt(100)), staking.Description{}, commission)
	require.True(t, sh(ctx, msg).IsOK())

	// end block to bond validator
	staking.EndBlocker(ctx, sk)

	// fetch validator and delegation
	val := sk.Validator(ctx, valOpAddr1)
	del := sk.Delegation(ctx, sdk.AccAddress(valOpAddr1), valOpAddr1)

	// end period
	endingPeriod := k.incrementValidatorPeriod(ctx, val)

	// calculate delegation rewards
	rewards := k.calculateDelegationRewards(ctx, val, del, endingPeriod)

	// rewards should be zero
	require.True(t, rewards.IsZero())

	// start out block height
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 3)

	// slash the validator by 50%
	sk.Slash(ctx, valConsAddr1, ctx.BlockHeight(), 100, sdk.NewDecWithPrec(5, 1))

	// retrieve validator
	val = sk.Validator(ctx, valOpAddr1)

	// increase block height
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 3)

	// allocate some rewards
	initial := int64(10)
	tokens := sdk.DecCoins{{staking.DefaultBondDenom, sdk.NewDec(initial)}}
	k.AllocateTokensToValidator(ctx, val, tokens)

	// end period
	endingPeriod = k.incrementValidatorPeriod(ctx, val)

	// calculate delegation rewards
	rewards = k.calculateDelegationRewards(ctx, val, del, endingPeriod)

	// rewards should be half the tokens
	require.Equal(t, sdk.DecCoins{{staking.DefaultBondDenom, sdk.NewDec(initial / 2)}}, rewards)

	// commission should be the other half
	require.Equal(t, sdk.DecCoins{{staking.DefaultBondDenom, sdk.NewDec(initial / 2)}}, k.GetValidatorAccumulatedCommission(ctx, valOpAddr1))
}

func TestCalculateRewardsAfterManySlashes(t *testing.T) {
	ctx, _, k, sk, _ := CreateTestInputDefault(t, false, 1000)
	sh := staking.NewHandler(sk)

	// initialize state
	k.SetOutstandingRewards(ctx, sdk.DecCoins{})

	// create validator with 50% commission
	commission := staking.NewCommissionMsg(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), sdk.NewDec(0))
	msg := staking.NewMsgCreateValidator(valOpAddr1, valConsPk1,
		sdk.NewCoin(staking.DefaultBondDenom, sdk.NewInt(100)), staking.Description{}, commission)
	require.True(t, sh(ctx, msg).IsOK())

	// end block to bond validator
	staking.EndBlocker(ctx, sk)

	// fetch validator and delegation
	val := sk.Validator(ctx, valOpAddr1)
	del := sk.Delegation(ctx, sdk.AccAddress(valOpAddr1), valOpAddr1)

	// end period
	endingPeriod := k.incrementValidatorPeriod(ctx, val)

	// calculate delegation rewards
	rewards := k.calculateDelegationRewards(ctx, val, del, endingPeriod)

	// rewards should be zero
	require.True(t, rewards.IsZero())

	// start out block height
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 3)

	// slash the validator by 50%
	sk.Slash(ctx, valConsAddr1, ctx.BlockHeight(), 100, sdk.NewDecWithPrec(5, 1))

	// fetch the validator again
	val = sk.Validator(ctx, valOpAddr1)

	// increase block height
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 3)

	// allocate some rewards
	initial := int64(10)
	tokens := sdk.DecCoins{{staking.DefaultBondDenom, sdk.NewDec(initial)}}
	k.AllocateTokensToValidator(ctx, val, tokens)

	// slash the validator by 50% again
	sk.Slash(ctx, valConsAddr1, ctx.BlockHeight(), 50, sdk.NewDecWithPrec(5, 1))

	// fetch the validator again
	val = sk.Validator(ctx, valOpAddr1)

	// increase block height
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 3)

	// allocate some more rewards
	k.AllocateTokensToValidator(ctx, val, tokens)

	// end period
	endingPeriod = k.incrementValidatorPeriod(ctx, val)

	// calculate delegation rewards
	rewards = k.calculateDelegationRewards(ctx, val, del, endingPeriod)

	// rewards should be half the tokens
	require.Equal(t, sdk.DecCoins{{staking.DefaultBondDenom, sdk.NewDec(initial)}}, rewards)

	// commission should be the other half
	require.Equal(t, sdk.DecCoins{{staking.DefaultBondDenom, sdk.NewDec(initial)}}, k.GetValidatorAccumulatedCommission(ctx, valOpAddr1))
}

func TestCalculateRewardsMultiDelegator(t *testing.T) {
	ctx, _, k, sk, _ := CreateTestInputDefault(t, false, 1000)
	sh := staking.NewHandler(sk)

	// initialize state
	k.SetOutstandingRewards(ctx, sdk.DecCoins{})

	// create validator with 50% commission
	commission := staking.NewCommissionMsg(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), sdk.NewDec(0))
	msg := staking.NewMsgCreateValidator(valOpAddr1, valConsPk1,
		sdk.NewCoin(staking.DefaultBondDenom, sdk.NewInt(100)), staking.Description{}, commission)
	require.True(t, sh(ctx, msg).IsOK())

	// end block to bond validator
	staking.EndBlocker(ctx, sk)

	// fetch validator and delegation
	val := sk.Validator(ctx, valOpAddr1)
	del1 := sk.Delegation(ctx, sdk.AccAddress(valOpAddr1), valOpAddr1)

	// allocate some rewards
	initial := int64(20)
	tokens := sdk.DecCoins{{staking.DefaultBondDenom, sdk.NewDec(initial)}}
	k.AllocateTokensToValidator(ctx, val, tokens)

	// second delegation
	msg2 := staking.NewMsgDelegate(sdk.AccAddress(valOpAddr2), valOpAddr1, sdk.NewCoin(staking.DefaultBondDenom, sdk.NewInt(100)))
	require.True(t, sh(ctx, msg2).IsOK())
	del2 := sk.Delegation(ctx, sdk.AccAddress(valOpAddr2), valOpAddr1)

	// fetch updated validator
	val = sk.Validator(ctx, valOpAddr1)

	// end block
	staking.EndBlocker(ctx, sk)

	// allocate some more rewards
	k.AllocateTokensToValidator(ctx, val, tokens)

	// end period
	endingPeriod := k.incrementValidatorPeriod(ctx, val)

	// calculate delegation rewards for del1
	rewards := k.calculateDelegationRewards(ctx, val, del1, endingPeriod)

	// rewards for del1 should be 3/4 initial
	require.Equal(t, sdk.DecCoins{{staking.DefaultBondDenom, sdk.NewDec(initial * 3 / 4)}}, rewards)

	// calculate delegation rewards for del2
	rewards = k.calculateDelegationRewards(ctx, val, del2, endingPeriod)

	// rewards for del2 should be 1/4 initial
	require.Equal(t, sdk.DecCoins{{staking.DefaultBondDenom, sdk.NewDec(initial * 1 / 4)}}, rewards)

	// commission should be equal to initial (50% twice)
	require.Equal(t, sdk.DecCoins{{staking.DefaultBondDenom, sdk.NewDec(initial)}}, k.GetValidatorAccumulatedCommission(ctx, valOpAddr1))
}

func TestWithdrawDelegationRewardsBasic(t *testing.T) {
	balance := int64(1000)
	ctx, ak, k, sk, _ := CreateTestInputDefault(t, false, balance)
	sh := staking.NewHandler(sk)

	// initialize state
	k.SetOutstandingRewards(ctx, sdk.DecCoins{})

	// create validator with 50% commission
	bond := int64(100)
	commission := staking.NewCommissionMsg(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), sdk.NewDec(0))
	msg := staking.NewMsgCreateValidator(valOpAddr1, valConsPk1,
		sdk.NewCoin(staking.DefaultBondDenom, sdk.NewInt(bond)), staking.Description{}, commission)
	require.True(t, sh(ctx, msg).IsOK())

	// assert correct initial balance
	require.Equal(t, sdk.Coins{{staking.DefaultBondDenom, sdk.NewInt(balance - bond)}}, ak.GetAccount(ctx, sdk.AccAddress(valOpAddr1)).GetCoins())

	// end block to bond validator
	staking.EndBlocker(ctx, sk)

	// set zero outstanding rewards
	k.SetOutstandingRewards(ctx, sdk.DecCoins{})

	// fetch validator and delegation
	val := sk.Validator(ctx, valOpAddr1)

	// allocate some rewards
	initial := int64(10)
	tokens := sdk.DecCoins{{staking.DefaultBondDenom, sdk.NewDec(initial)}}
	k.AllocateTokensToValidator(ctx, val, tokens)

	// withdraw rewards
	require.Nil(t, k.WithdrawDelegationRewards(ctx, sdk.AccAddress(valOpAddr1), valOpAddr1))

	// assert correct balance
	require.Equal(t, sdk.Coins{{staking.DefaultBondDenom, sdk.NewInt(balance - bond + (initial / 2))}}, ak.GetAccount(ctx, sdk.AccAddress(valOpAddr1)).GetCoins())

	// withdraw commission
	require.Nil(t, k.WithdrawValidatorCommission(ctx, valOpAddr1))

	// assert correct balance
	require.Equal(t, sdk.Coins{{staking.DefaultBondDenom, sdk.NewInt(balance - bond + initial)}}, ak.GetAccount(ctx, sdk.AccAddress(valOpAddr1)).GetCoins())
}

func TestCalculateRewardsAfterManySlashesInSameBlock(t *testing.T) {
	ctx, _, k, sk, _ := CreateTestInputDefault(t, false, 1000)
	sh := staking.NewHandler(sk)

	// initialize state
	k.SetOutstandingRewards(ctx, sdk.DecCoins{})

	// create validator with 50% commission
	commission := staking.NewCommissionMsg(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), sdk.NewDec(0))
	msg := staking.NewMsgCreateValidator(valOpAddr1, valConsPk1,
		sdk.NewCoin(staking.DefaultBondDenom, sdk.NewInt(100)), staking.Description{}, commission)
	require.True(t, sh(ctx, msg).IsOK())

	// end block to bond validator
	staking.EndBlocker(ctx, sk)

	// fetch validator and delegation
	val := sk.Validator(ctx, valOpAddr1)
	del := sk.Delegation(ctx, sdk.AccAddress(valOpAddr1), valOpAddr1)

	// end period
	endingPeriod := k.incrementValidatorPeriod(ctx, val)

	// calculate delegation rewards
	rewards := k.calculateDelegationRewards(ctx, val, del, endingPeriod)

	// rewards should be zero
	require.True(t, rewards.IsZero())

	// start out block height
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 3)

	// allocate some rewards
	initial := int64(10)
	tokens := sdk.DecCoins{{staking.DefaultBondDenom, sdk.NewDec(initial)}}
	k.AllocateTokensToValidator(ctx, val, tokens)

	// slash the validator by 50%
	sk.Slash(ctx, valConsAddr1, ctx.BlockHeight(), 100, sdk.NewDecWithPrec(5, 1))

	// slash the validator by 50% again
	sk.Slash(ctx, valConsAddr1, ctx.BlockHeight(), 50, sdk.NewDecWithPrec(5, 1))

	// fetch the validator again
	val = sk.Validator(ctx, valOpAddr1)

	// increase block height
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 3)

	// allocate some more rewards
	k.AllocateTokensToValidator(ctx, val, tokens)

	// end period
	endingPeriod = k.incrementValidatorPeriod(ctx, val)

	// calculate delegation rewards
	rewards = k.calculateDelegationRewards(ctx, val, del, endingPeriod)

	// rewards should be half the tokens
	require.Equal(t, sdk.DecCoins{{staking.DefaultBondDenom, sdk.NewDec(initial)}}, rewards)

	// commission should be the other half
	require.Equal(t, sdk.DecCoins{{staking.DefaultBondDenom, sdk.NewDec(initial)}}, k.GetValidatorAccumulatedCommission(ctx, valOpAddr1))
}

func TestCalculateRewardsMultiDelegatorMultiSlash(t *testing.T) {
	ctx, _, k, sk, _ := CreateTestInputDefault(t, false, 1000)
	sh := staking.NewHandler(sk)

	// initialize state
	k.SetOutstandingRewards(ctx, sdk.DecCoins{})

	// create validator with 50% commission
	commission := staking.NewCommissionMsg(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), sdk.NewDec(0))
	msg := staking.NewMsgCreateValidator(valOpAddr1, valConsPk1,
		sdk.NewCoin(staking.DefaultBondDenom, sdk.NewInt(100)), staking.Description{}, commission)
	require.True(t, sh(ctx, msg).IsOK())

	// end block to bond validator
	staking.EndBlocker(ctx, sk)

	// fetch validator and delegation
	val := sk.Validator(ctx, valOpAddr1)
	del1 := sk.Delegation(ctx, sdk.AccAddress(valOpAddr1), valOpAddr1)

	// allocate some rewards
	initial := int64(30)
	tokens := sdk.DecCoins{{staking.DefaultBondDenom, sdk.NewDec(initial)}}
	k.AllocateTokensToValidator(ctx, val, tokens)

	// slash the validator
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 3)
	sk.Slash(ctx, valConsAddr1, ctx.BlockHeight(), 100, sdk.NewDecWithPrec(5, 1))
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 3)

	// second delegation
	msg2 := staking.NewMsgDelegate(sdk.AccAddress(valOpAddr2), valOpAddr1, sdk.NewCoin(staking.DefaultBondDenom, sdk.NewInt(100)))
	require.True(t, sh(ctx, msg2).IsOK())
	del2 := sk.Delegation(ctx, sdk.AccAddress(valOpAddr2), valOpAddr1)

	// end block
	staking.EndBlocker(ctx, sk)

	// allocate some more rewards
	k.AllocateTokensToValidator(ctx, val, tokens)

	// slash the validator again
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 3)
	sk.Slash(ctx, valConsAddr1, ctx.BlockHeight(), 100, sdk.NewDecWithPrec(5, 1))
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 3)

	// fetch updated validator
	val = sk.Validator(ctx, valOpAddr1)

	// end period
	endingPeriod := k.incrementValidatorPeriod(ctx, val)

	// calculate delegation rewards for del1
	rewards := k.calculateDelegationRewards(ctx, val, del1, endingPeriod)

	// rewards for del1 should be 2/3 initial (half initial first period, 1/6 initial second period)
	require.Equal(t, sdk.DecCoins{{staking.DefaultBondDenom, sdk.NewDec((initial / 2) + (initial / 6))}}, rewards)

	// calculate delegation rewards for del2
	rewards = k.calculateDelegationRewards(ctx, val, del2, endingPeriod)

	// rewards for del2 should be initial / 3
	require.Equal(t, sdk.DecCoins{{staking.DefaultBondDenom, sdk.NewDec(initial / 3)}}, rewards)

	// commission should be equal to initial (twice 50% commission, unaffected by slashing)
	require.Equal(t, sdk.DecCoins{{staking.DefaultBondDenom, sdk.NewDec(initial)}}, k.GetValidatorAccumulatedCommission(ctx, valOpAddr1))
}

func TestCalculateRewardsMultiDelegatorMultWithdraw(t *testing.T) {
	ctx, _, k, sk, _ := CreateTestInputDefault(t, false, 1000)
	sh := staking.NewHandler(sk)

	// initialize state
	k.SetOutstandingRewards(ctx, sdk.DecCoins{})

	// create validator with 50% commission
	commission := staking.NewCommissionMsg(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), sdk.NewDec(0))
	msg := staking.NewMsgCreateValidator(valOpAddr1, valConsPk1,
		sdk.NewCoin(staking.DefaultBondDenom, sdk.NewInt(100)), staking.Description{}, commission)
	require.True(t, sh(ctx, msg).IsOK())

	// end block to bond validator
	staking.EndBlocker(ctx, sk)

	// fetch validator and delegation
	val := sk.Validator(ctx, valOpAddr1)
	del1 := sk.Delegation(ctx, sdk.AccAddress(valOpAddr1), valOpAddr1)

	// allocate some rewards
	initial := int64(20)
	tokens := sdk.DecCoins{{staking.DefaultBondDenom, sdk.NewDec(initial)}}
	k.AllocateTokensToValidator(ctx, val, tokens)

	// second delegation
	msg2 := staking.NewMsgDelegate(sdk.AccAddress(valOpAddr2), valOpAddr1, sdk.NewCoin(staking.DefaultBondDenom, sdk.NewInt(100)))
	require.True(t, sh(ctx, msg2).IsOK())

	// fetch updated validator
	val = sk.Validator(ctx, valOpAddr1)
	del2 := sk.Delegation(ctx, sdk.AccAddress(valOpAddr2), valOpAddr1)

	// end block
	staking.EndBlocker(ctx, sk)

	// allocate some more rewards
	k.AllocateTokensToValidator(ctx, val, tokens)

	// first delegator withdraws
	k.WithdrawDelegationRewards(ctx, sdk.AccAddress(valOpAddr1), valOpAddr1)

	// second delegator withdraws
	k.WithdrawDelegationRewards(ctx, sdk.AccAddress(valOpAddr2), valOpAddr1)

	// validator withdraws commission
	k.WithdrawValidatorCommission(ctx, valOpAddr1)

	// end period
	endingPeriod := k.incrementValidatorPeriod(ctx, val)

	// calculate delegation rewards for del1
	rewards := k.calculateDelegationRewards(ctx, val, del1, endingPeriod)

	// rewards for del1 should be zero
	require.True(t, rewards.IsZero())

	// calculate delegation rewards for del2
	rewards = k.calculateDelegationRewards(ctx, val, del2, endingPeriod)

	// rewards for del2 should be zero
	require.True(t, rewards.IsZero())

	// commission should be zero
	require.True(t, k.GetValidatorAccumulatedCommission(ctx, valOpAddr1).IsZero())

	// allocate some more rewards
	k.AllocateTokensToValidator(ctx, val, tokens)

	// first delegator withdraws again
	k.WithdrawDelegationRewards(ctx, sdk.AccAddress(valOpAddr1), valOpAddr1)

	// end period
	endingPeriod = k.incrementValidatorPeriod(ctx, val)

	// calculate delegation rewards for del1
	rewards = k.calculateDelegationRewards(ctx, val, del1, endingPeriod)

	// rewards for del1 should be zero
	require.True(t, rewards.IsZero())

	// calculate delegation rewards for del2
	rewards = k.calculateDelegationRewards(ctx, val, del2, endingPeriod)

	// rewards for del2 should be 1/4 initial
	require.Equal(t, sdk.DecCoins{{staking.DefaultBondDenom, sdk.NewDec(initial / 4)}}, rewards)

	// commission should be half initial
	require.Equal(t, sdk.DecCoins{{staking.DefaultBondDenom, sdk.NewDec(initial / 2)}}, k.GetValidatorAccumulatedCommission(ctx, valOpAddr1))

	// allocate some more rewards
	k.AllocateTokensToValidator(ctx, val, tokens)

	// withdraw commission
	k.WithdrawValidatorCommission(ctx, valOpAddr1)

	// end period
	endingPeriod = k.incrementValidatorPeriod(ctx, val)

	// calculate delegation rewards for del1
	rewards = k.calculateDelegationRewards(ctx, val, del1, endingPeriod)

	// rewards for del1 should be 1/4 initial
	require.Equal(t, sdk.DecCoins{{staking.DefaultBondDenom, sdk.NewDec(initial / 4)}}, rewards)

	// calculate delegation rewards for del2
	rewards = k.calculateDelegationRewards(ctx, val, del2, endingPeriod)

	// rewards for del2 should be 1/2 initial
	require.Equal(t, sdk.DecCoins{{staking.DefaultBondDenom, sdk.NewDec(initial / 2)}}, rewards)

	// commission should be zero
	require.True(t, k.GetValidatorAccumulatedCommission(ctx, valOpAddr1).IsZero())
}
