package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

func TestAllocateTokensTruncation(t *testing.T) {
	communityTax := sdk.NewDec(0)
	ctx, ak, bk, k, sk, _, supplyKeeper := CreateTestInputAdvanced(t, false, 1000000, communityTax)
	sh := staking.NewHandler(sk)

	// create validator with 10% commission
	commission := staking.NewCommissionRates(sdk.NewDecWithPrec(1, 1), sdk.NewDecWithPrec(1, 1), sdk.NewDec(0))
	msg := staking.NewMsgCreateValidator(valOpAddr1, valConsPk1,
		sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(110)), staking.Description{}, commission, sdk.OneInt())
	res, err := sh(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, res)

	// create second validator with 10% commission
	commission = staking.NewCommissionRates(sdk.NewDecWithPrec(1, 1), sdk.NewDecWithPrec(1, 1), sdk.NewDec(0))
	msg = staking.NewMsgCreateValidator(valOpAddr2, valConsPk2,
		sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100)), staking.Description{}, commission, sdk.OneInt())
	res, err = sh(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, res)

	// create third validator with 10% commission
	commission = staking.NewCommissionRates(sdk.NewDecWithPrec(1, 1), sdk.NewDecWithPrec(1, 1), sdk.NewDec(0))
	msg = staking.NewMsgCreateValidator(valOpAddr3, valConsPk3,
		sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100)), staking.Description{}, commission, sdk.OneInt())
	res, err = sh(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, res)

	abciValA := abci.Validator{
		Address: valConsPk1.Address(),
		Power:   11,
	}
	abciValB := abci.Validator{
		Address: valConsPk2.Address(),
		Power:   10,
	}
	abciValС := abci.Validator{
		Address: valConsPk3.Address(),
		Power:   10,
	}

	// assert initial state: zero outstanding rewards, zero community pool, zero commission, zero current rewards
	require.True(t, k.GetValidatorOutstandingRewards(ctx, valOpAddr1).Rewards.IsZero())
	require.True(t, k.GetValidatorOutstandingRewards(ctx, valOpAddr2).Rewards.IsZero())
	require.True(t, k.GetValidatorOutstandingRewards(ctx, valOpAddr3).Rewards.IsZero())
	require.True(t, k.GetFeePool(ctx).CommunityPool.IsZero())
	require.True(t, k.GetValidatorAccumulatedCommission(ctx, valOpAddr1).Commission.IsZero())
	require.True(t, k.GetValidatorAccumulatedCommission(ctx, valOpAddr2).Commission.IsZero())
	require.True(t, k.GetValidatorCurrentRewards(ctx, valOpAddr1).Rewards.IsZero())
	require.True(t, k.GetValidatorCurrentRewards(ctx, valOpAddr2).Rewards.IsZero())

	// allocate tokens as if both had voted and second was proposer
	fees := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(634195840)))

	feeCollector := supplyKeeper.GetModuleAccount(ctx, k.feeCollectorName)
	require.NotNil(t, feeCollector)

	err = bk.SetBalances(ctx, feeCollector.GetAddress(), fees)
	require.NoError(t, err)

	ak.SetAccount(ctx, feeCollector)

	votes := []abci.VoteInfo{
		{
			Validator:       abciValA,
			SignedLastBlock: true,
		},
		{
			Validator:       abciValB,
			SignedLastBlock: true,
		},
		{
			Validator:       abciValС,
			SignedLastBlock: true,
		},
	}
	k.AllocateTokens(ctx, 31, 31, valConsAddr2, votes)

	require.True(t, k.GetValidatorOutstandingRewards(ctx, valOpAddr1).Rewards.IsValid())
	require.True(t, k.GetValidatorOutstandingRewards(ctx, valOpAddr2).Rewards.IsValid())
	require.True(t, k.GetValidatorOutstandingRewards(ctx, valOpAddr3).Rewards.IsValid())
}
