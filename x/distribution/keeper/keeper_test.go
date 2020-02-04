package keeper

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestSetWithdrawAddr(t *testing.T) {
	ctx, _, _, keeper, _, _ := CreateTestInputDefault(t, false, 1000) // nolint: dogseld

	params := keeper.GetParams(ctx)
	params.WithdrawAddrEnabled = false
	keeper.SetParams(ctx, params)

	err := keeper.SetWithdrawAddr(ctx, delAddr1, delAddr2)
	require.NotNil(t, err)

	params.WithdrawAddrEnabled = true
	keeper.SetParams(ctx, params)

	err = keeper.SetWithdrawAddr(ctx, delAddr1, delAddr2)
	require.Nil(t, err)

	keeper.blacklistedAddrs[distrAcc.GetAddress().String()] = true
	require.Error(t, keeper.SetWithdrawAddr(ctx, delAddr1, distrAcc.GetAddress()))
}

func TestWithdrawValidatorCommission(t *testing.T) {
	ctx, _, bk, keeper, _, _ := CreateTestInputDefault(t, false, 1000)

	valCommission := sdk.DecCoins{
		sdk.NewDecCoinFromDec("mytoken", sdk.NewDec(5).Quo(sdk.NewDec(4))),
		sdk.NewDecCoinFromDec("stake", sdk.NewDec(3).Quo(sdk.NewDec(2))),
	}

	// set module account coins
	distrAcc := keeper.GetDistributionAccount(ctx)
	bk.SetBalances(ctx, distrAcc.GetAddress(), sdk.NewCoins(
		sdk.NewCoin("mytoken", sdk.NewInt(2)),
		sdk.NewCoin("stake", sdk.NewInt(2)),
	))
	keeper.supplyKeeper.SetModuleAccount(ctx, distrAcc)

	// check initial balance
	balance := bk.GetAllBalances(ctx, sdk.AccAddress(valOpAddr3))
	expTokens := sdk.TokensFromConsensusPower(1000)
	expCoins := sdk.NewCoins(sdk.NewCoin("stake", expTokens))
	require.Equal(t, expCoins, balance)

	// set outstanding rewards
	keeper.SetValidatorOutstandingRewards(ctx, valOpAddr3, valCommission)

	// set commission
	keeper.SetValidatorAccumulatedCommission(ctx, valOpAddr3, valCommission)

	// withdraw commission
	keeper.WithdrawValidatorCommission(ctx, valOpAddr3)

	// check balance increase
	balance = bk.GetAllBalances(ctx, sdk.AccAddress(valOpAddr3))
	require.Equal(t, sdk.NewCoins(
		sdk.NewCoin("mytoken", sdk.NewInt(1)),
		sdk.NewCoin("stake", expTokens.AddRaw(1)),
	), balance)

	// check remainder
	remainder := keeper.GetValidatorAccumulatedCommission(ctx, valOpAddr3)
	require.Equal(t, sdk.DecCoins{
		sdk.NewDecCoinFromDec("mytoken", sdk.NewDec(1).Quo(sdk.NewDec(4))),
		sdk.NewDecCoinFromDec("stake", sdk.NewDec(1).Quo(sdk.NewDec(2))),
	}, remainder)

	require.True(t, true)
}

func TestGetTotalRewards(t *testing.T) {
	ctx, _, _, keeper, _, _ := CreateTestInputDefault(t, false, 1000) // nolint: dogseld

	valCommission := sdk.DecCoins{
		sdk.NewDecCoinFromDec("mytoken", sdk.NewDec(5).Quo(sdk.NewDec(4))),
		sdk.NewDecCoinFromDec("stake", sdk.NewDec(3).Quo(sdk.NewDec(2))),
	}

	keeper.SetValidatorOutstandingRewards(ctx, valOpAddr1, valCommission)
	keeper.SetValidatorOutstandingRewards(ctx, valOpAddr2, valCommission)

	expectedRewards := valCommission.MulDec(sdk.NewDec(2))
	totalRewards := keeper.GetTotalRewards(ctx)

	require.Equal(t, expectedRewards, totalRewards)
}

func TestFundCommunityPool(t *testing.T) {
	// nolint dogsled
	ctx, _, bk, keeper, _, _, _ := CreateTestInputAdvanced(t, false, 1000, sdk.NewDecWithPrec(2, 2))

	amount := sdk.NewCoins(sdk.NewInt64Coin("stake", 100))
	require.NoError(t, bk.SetBalances(ctx, delAddr1, amount))

	initPool := keeper.GetFeePool(ctx)
	assert.Empty(t, initPool.CommunityPool)

	err := keeper.FundCommunityPool(ctx, amount, delAddr1)
	assert.Nil(t, err)

	assert.Equal(t, initPool.CommunityPool.Add(sdk.NewDecCoinsFromCoins(amount...)...), keeper.GetFeePool(ctx).CommunityPool)
	assert.Empty(t, bk.GetAllBalances(ctx, delAddr1))
}
