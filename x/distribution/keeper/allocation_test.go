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

	// initialize state
	k.SetOutstandingRewards(ctx, sdk.DecCoins{})

	// create validator with 50% commission
	commission := staking.NewCommissionMsg(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), sdk.NewDec(0))
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
	ctx, _, k, sk, fck := CreateTestInputDefault(t, false, 1000)
	sh := staking.NewHandler(sk)

	// initialize state
	k.SetOutstandingRewards(ctx, sdk.DecCoins{})

	// create validator with 50% commission
	commission := staking.NewCommissionMsg(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), sdk.NewDec(0))
	msg := staking.NewMsgCreateValidator(valOpAddr1, valConsPk1,
		sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100)), staking.Description{}, commission, sdk.OneInt())
	require.True(t, sh(ctx, msg).IsOK())

	// create second validator with 0% commission
	commission = staking.NewCommissionMsg(sdk.NewDec(0), sdk.NewDec(0), sdk.NewDec(0))
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
	require.True(t, k.GetOutstandingRewards(ctx).IsZero())
	require.True(t, k.GetFeePool(ctx).CommunityPool.IsZero())
	require.True(t, k.GetValidatorAccumulatedCommission(ctx, valOpAddr1).IsZero())
	require.True(t, k.GetValidatorAccumulatedCommission(ctx, valOpAddr2).IsZero())
	require.True(t, k.GetValidatorCurrentRewards(ctx, valOpAddr1).Rewards.IsZero())
	require.True(t, k.GetValidatorCurrentRewards(ctx, valOpAddr2).Rewards.IsZero())

	// allocate tokens as if both had voted and second was proposer
	fees := sdk.Coins{
		{sdk.DefaultBondDenom, sdk.NewInt(100)},
	}
	fck.SetCollectedFees(fees)
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
	require.Equal(t, sdk.DecCoins{{sdk.DefaultBondDenom, sdk.NewDec(98)}}, k.GetOutstandingRewards(ctx))
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
