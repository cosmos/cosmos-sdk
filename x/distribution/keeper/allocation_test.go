package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

func TestAllocateTokensToValidatorWithCommission(t *testing.T) {
	ctx, _, k, sk, _ := CreateTestInputDefault(t, false, 1000)
	sh := staking.NewHandler(sk)

	// create validator with 50% commission
	commission := staking.NewCommissionRates(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), sdk.NewDec(0))
	msg := staking.NewMsgCreateValidator(valOpAddr1, valConsPk1,
		sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100)), staking.Description{}, commission, sdk.OneInt())
	require.True(t, sh(ctx, msg).IsOK())
	val := sk.Validator(ctx, valOpAddr1)

	// allocate tokens
	tokens := sdk.DecCoins{
		{sdk.DefaultBondDenom, sdk.NewDec(10)},
	}
	k.AllocateTokensToValidator(ctx, val, tokens)

	// check commission
	expected := sdk.DecCoins{
		{sdk.DefaultBondDenom, sdk.NewDec(5)},
	}
	require.Equal(t, expected, k.GetValidatorAccumulatedCommission(ctx, val.GetOperator()))

	// check current rewards
	require.Equal(t, expected, k.GetValidatorCurrentRewards(ctx, val.GetOperator()).Rewards)
}

func TestAllocateTokensToManyValidators(t *testing.T) {
	ctx, ak, k, sk, supplyKeeper := CreateTestInputDefault(t, false, 1000)
	sh := staking.NewHandler(sk)

	// create validator with 50% commission
	commission := staking.NewCommissionRates(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), sdk.NewDec(0))
	msg := staking.NewMsgCreateValidator(valOpAddr1, valConsPk1,
		sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100)), staking.Description{}, commission, sdk.OneInt())
	require.True(t, sh(ctx, msg).IsOK())

	// create second validator with 0% commission
	commission = staking.NewCommissionRates(sdk.NewDec(0), sdk.NewDec(0), sdk.NewDec(0))
	msg = staking.NewMsgCreateValidator(valOpAddr2, valConsPk2,
		sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100)), staking.Description{}, commission, sdk.OneInt())
	require.True(t, sh(ctx, msg).IsOK())

	abciValA := abci.Validator{
		Address: valConsPk1.Address(),
		Power:   100,
	}
	abciValB := abci.Validator{
		Address: valConsPk2.Address(),
		Power:   100,
	}

	// assert initial state: zero outstanding rewards, zero community pool, zero commission, zero current rewards
	require.True(t, k.GetValidatorOutstandingRewards(ctx, valOpAddr1).IsZero())
	require.True(t, k.GetValidatorOutstandingRewards(ctx, valOpAddr2).IsZero())
	require.True(t, k.GetFeePool(ctx).CommunityPool.IsZero())
	require.True(t, k.GetValidatorAccumulatedCommission(ctx, valOpAddr1).IsZero())
	require.True(t, k.GetValidatorAccumulatedCommission(ctx, valOpAddr2).IsZero())
	require.True(t, k.GetValidatorCurrentRewards(ctx, valOpAddr1).Rewards.IsZero())
	require.True(t, k.GetValidatorCurrentRewards(ctx, valOpAddr2).Rewards.IsZero())

	// allocate tokens as if both had voted and second was proposer
	fees := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100)))
	feeCollector := supplyKeeper.GetModuleAccount(ctx, k.feeCollectorName)
	require.NotNil(t, feeCollector)

	err := feeCollector.SetCoins(fees)
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
	}
	k.AllocateTokens(ctx, 200, 200, valConsAddr2, votes)

	// 98 outstanding rewards (100 less 2 to community pool)
	require.Equal(t, sdk.DecCoins{{sdk.DefaultBondDenom, sdk.NewDecWithPrec(465, 1)}}, k.GetValidatorOutstandingRewards(ctx, valOpAddr1))
	require.Equal(t, sdk.DecCoins{{sdk.DefaultBondDenom, sdk.NewDecWithPrec(515, 1)}}, k.GetValidatorOutstandingRewards(ctx, valOpAddr2))
	// 2 community pool coins
	require.Equal(t, sdk.DecCoins{{sdk.DefaultBondDenom, sdk.NewDec(2)}}, k.GetFeePool(ctx).CommunityPool)
	// 50% commission for first proposer, (0.5 * 93%) * 100 / 2 = 23.25
	require.Equal(t, sdk.DecCoins{{sdk.DefaultBondDenom, sdk.NewDecWithPrec(2325, 2)}}, k.GetValidatorAccumulatedCommission(ctx, valOpAddr1))
	// zero commission for second proposer
	require.True(t, k.GetValidatorAccumulatedCommission(ctx, valOpAddr2).IsZero())
	// just staking.proportional for first proposer less commission = (0.5 * 93%) * 100 / 2 = 23.25
	require.Equal(t, sdk.DecCoins{{sdk.DefaultBondDenom, sdk.NewDecWithPrec(2325, 2)}}, k.GetValidatorCurrentRewards(ctx, valOpAddr1).Rewards)
	// proposer reward + staking.proportional for second proposer = (5 % + 0.5 * (93%)) * 100 = 51.5
	require.Equal(t, sdk.DecCoins{{sdk.DefaultBondDenom, sdk.NewDecWithPrec(515, 1)}}, k.GetValidatorCurrentRewards(ctx, valOpAddr2).Rewards)
}

func TestAllocateTokensTruncation(t *testing.T) {
	communityTax := sdk.NewDec(0)
	ctx, ak, _, k, sk, _, supplyKeeper := CreateTestInputAdvanced(t, false, 1000000, communityTax)
	sh := staking.NewHandler(sk)

	// create validator with 10% commission
	commission := staking.NewCommissionRates(sdk.NewDecWithPrec(1, 1), sdk.NewDecWithPrec(1, 1), sdk.NewDec(0))
	msg := staking.NewMsgCreateValidator(valOpAddr1, valConsPk1,
		sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(110)), staking.Description{}, commission, sdk.OneInt())
	require.True(t, sh(ctx, msg).IsOK())

	// create second validator with 10% commission
	commission = staking.NewCommissionRates(sdk.NewDecWithPrec(1, 1), sdk.NewDecWithPrec(1, 1), sdk.NewDec(0))
	msg = staking.NewMsgCreateValidator(valOpAddr2, valConsPk2,
		sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100)), staking.Description{}, commission, sdk.OneInt())
	require.True(t, sh(ctx, msg).IsOK())

	// create third validator with 10% commission
	commission = staking.NewCommissionRates(sdk.NewDecWithPrec(1, 1), sdk.NewDecWithPrec(1, 1), sdk.NewDec(0))
	msg = staking.NewMsgCreateValidator(valOpAddr3, valConsPk3,
		sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100)), staking.Description{}, commission, sdk.OneInt())
	require.True(t, sh(ctx, msg).IsOK())

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
	require.True(t, k.GetValidatorOutstandingRewards(ctx, valOpAddr1).IsZero())
	require.True(t, k.GetValidatorOutstandingRewards(ctx, valOpAddr2).IsZero())
	require.True(t, k.GetValidatorOutstandingRewards(ctx, valOpAddr3).IsZero())
	require.True(t, k.GetFeePool(ctx).CommunityPool.IsZero())
	require.True(t, k.GetValidatorAccumulatedCommission(ctx, valOpAddr1).IsZero())
	require.True(t, k.GetValidatorAccumulatedCommission(ctx, valOpAddr2).IsZero())
	require.True(t, k.GetValidatorCurrentRewards(ctx, valOpAddr1).Rewards.IsZero())
	require.True(t, k.GetValidatorCurrentRewards(ctx, valOpAddr2).Rewards.IsZero())

	// allocate tokens as if both had voted and second was proposer
	fees := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(634195840)))

	feeCollector := supplyKeeper.GetModuleAccount(ctx, k.feeCollectorName)
	require.NotNil(t, feeCollector)

	err := feeCollector.SetCoins(fees)
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

	require.True(t, k.GetValidatorOutstandingRewards(ctx, valOpAddr1).IsValid())
	require.True(t, k.GetValidatorOutstandingRewards(ctx, valOpAddr2).IsValid())
	require.True(t, k.GetValidatorOutstandingRewards(ctx, valOpAddr3).IsValid())
}
