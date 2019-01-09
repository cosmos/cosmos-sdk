package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake"
	"github.com/stretchr/testify/require"
)

func TestInitializeValidatorCalculateRewardsBasic(t *testing.T) {
	ctx, _, k, sk, _ := CreateTestInputDefault(t, false, 1000)
	sh := stake.NewHandler(sk)

	// create validator with 50% commission
	commission := stake.NewCommissionMsg(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), sdk.NewDec(0))
	msg := stake.NewMsgCreateValidator(valOpAddr1, valConsPk1,
		sdk.NewCoin(stake.DefaultBondDenom, sdk.NewInt(100)), stake.Description{}, commission)
	require.True(t, sh(ctx, msg).IsOK())

	// end block to bond validator
	stake.EndBlocker(ctx, sk)

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
	tokens := sdk.DecCoins{{stake.DefaultBondDenom, sdk.NewDec(10)}}
	k.AllocateTokensToValidator(ctx, val, tokens)

	// end period
	endingPeriod = k.incrementValidatorPeriod(ctx, val)

	// calculate delegation rewards
	rewards = k.calculateDelegationRewards(ctx, val, del, endingPeriod)

	// rewards should be half the tokens
	require.Equal(t, sdk.DecCoins{{stake.DefaultBondDenom, sdk.NewDec(5)}}, rewards)

	// commission should be the other half
	require.Equal(t, sdk.DecCoins{{stake.DefaultBondDenom, sdk.NewDec(5)}}, k.GetValidatorAccumulatedCommission(ctx, valOpAddr1))
}

func TestInitializeValidatorCalculateRewardsMultiDelegator(t *testing.T) {
	ctx, _, k, sk, _ := CreateTestInputDefault(t, false, 1000)
	sh := stake.NewHandler(sk)

	// create validator with 50% commission
	commission := stake.NewCommissionMsg(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), sdk.NewDec(0))
	msg := stake.NewMsgCreateValidator(valOpAddr1, valConsPk1,
		sdk.NewCoin(stake.DefaultBondDenom, sdk.NewInt(100)), stake.Description{}, commission)
	require.True(t, sh(ctx, msg).IsOK())

	// end block to bond validator
	stake.EndBlocker(ctx, sk)

	// fetch validator and delegation
	val := sk.Validator(ctx, valOpAddr1)
	del1 := sk.Delegation(ctx, sdk.AccAddress(valOpAddr1), valOpAddr1)

	// allocate some rewards
	tokens := sdk.DecCoins{{stake.DefaultBondDenom, sdk.NewDec(10)}}
	k.AllocateTokensToValidator(ctx, val, tokens)

	// second delegation
	msg2 := stake.NewMsgDelegate(sdk.AccAddress(valOpAddr2), valOpAddr1, sdk.NewCoin(stake.DefaultBondDenom, sdk.NewInt(100)))
	require.True(t, sh(ctx, msg2).IsOK())
	del2 := sk.Delegation(ctx, sdk.AccAddress(valOpAddr2), valOpAddr1)

	// fetch updated validator
	val = sk.Validator(ctx, valOpAddr1)

	// end block
	stake.EndBlocker(ctx, sk)

	// allocate some more rewards
	k.AllocateTokensToValidator(ctx, val, tokens)

	// end period
	endingPeriod := k.incrementValidatorPeriod(ctx, val)

	// calculate delegation rewards for del1
	rewards := k.calculateDelegationRewards(ctx, val, del1, endingPeriod)

	// rewards for del1 should be 7.5
	require.Equal(t, sdk.DecCoins{{stake.DefaultBondDenom, sdk.NewDecWithPrec(75, 1)}}, rewards)

	// calculate delegation rewards for del2
	rewards = k.calculateDelegationRewards(ctx, val, del2, endingPeriod)

	// rewards for del2 should be 2.5
	require.Equal(t, sdk.DecCoins{{stake.DefaultBondDenom, sdk.NewDecWithPrec(25, 1)}}, rewards)

	// commission should be 10
	require.Equal(t, sdk.DecCoins{{stake.DefaultBondDenom, sdk.NewDec(10)}}, k.GetValidatorAccumulatedCommission(ctx, valOpAddr1))
}

func TestWithdrawDelegationRewardsBasic(t *testing.T) {
	ctx, ak, k, sk, _ := CreateTestInputDefault(t, false, 1000)
	sh := stake.NewHandler(sk)

	// create validator with 50% commission
	commission := stake.NewCommissionMsg(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), sdk.NewDec(0))
	msg := stake.NewMsgCreateValidator(valOpAddr1, valConsPk1,
		sdk.NewCoin(stake.DefaultBondDenom, sdk.NewInt(100)), stake.Description{}, commission)
	require.True(t, sh(ctx, msg).IsOK())

	// assert correct initial balance
	require.Equal(t, sdk.Coins{{stake.DefaultBondDenom, sdk.NewInt(900)}}, ak.GetAccount(ctx, sdk.AccAddress(valOpAddr1)).GetCoins())

	// end block to bond validator
	stake.EndBlocker(ctx, sk)

	// set zero outstanding rewards
	k.SetOutstandingRewards(ctx, sdk.DecCoins{})

	// fetch validator and delegation
	val := sk.Validator(ctx, valOpAddr1)

	// allocate some rewards
	tokens := sdk.DecCoins{{stake.DefaultBondDenom, sdk.NewDec(10)}}
	k.AllocateTokensToValidator(ctx, val, tokens)

	// withdraw rewards
	require.Nil(t, k.WithdrawDelegationRewards(ctx, sdk.AccAddress(valOpAddr1), valOpAddr1))

	// assert correct balance
	require.Equal(t, sdk.Coins{{stake.DefaultBondDenom, sdk.NewInt(905)}}, ak.GetAccount(ctx, sdk.AccAddress(valOpAddr1)).GetCoins())

	// withdraw commission
	require.Nil(t, k.WithdrawValidatorCommission(ctx, valOpAddr1))

	// assert correct balance
	require.Equal(t, sdk.Coins{{stake.DefaultBondDenom, sdk.NewInt(910)}}, ak.GetAccount(ctx, sdk.AccAddress(valOpAddr1)).GetCoins())
}

func TestUpdateValidatorSlashFraction(t *testing.T) {
	require.True(t, true)
}
