package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAllocateFeesBasic(t *testing.T) {

	// no community tax on inputs
	ctx, _, keeper, sk, fck := CreateTestInputAdvanced(t, false, 100, sdk.ZeroDec())
	stakeHandler := stake.NewHandler(sk)
	denom := sk.GetParams(ctx).BondDenom

	//first make a validator
	totalPower := int64(10)
	totalPowerDec := sdk.NewDec(totalPower)
	msgCreateValidator := stake.NewTestMsgCreateValidator(valOpAddr1, valConsPk1, totalPower)
	got := stakeHandler(ctx, msgCreateValidator)
	require.True(t, got.IsOK(), "expected msg to be ok, got %v", got)
	_ = sk.ApplyAndReturnValidatorSetUpdates(ctx)

	// verify everything has been set in staking correctly
	validator, found := sk.GetValidator(ctx, valOpAddr1)
	require.True(t, found)
	require.Equal(t, sdk.Bonded, validator.Status)
	assert.True(sdk.DecEq(t, totalPowerDec, validator.Tokens))
	assert.True(sdk.DecEq(t, totalPowerDec, validator.DelegatorShares))
	bondedTokens := sk.TotalPower(ctx)
	assert.True(sdk.DecEq(t, totalPowerDec, bondedTokens))

	// initial fee pool should be empty
	feePool := keeper.GetFeePool(ctx)
	require.Nil(t, feePool.Pool)

	// allocate 100 denom of fees
	feeInputs := sdk.NewInt(100)
	fck.SetCollectedFees(sdk.Coins{sdk.NewCoin(denom, feeInputs)})
	require.Equal(t, feeInputs, fck.GetCollectedFees(ctx).AmountOf(denom))
	keeper.AllocateFees(ctx, sdk.OneDec(), valConsAddr1)

	// verify that these fees have been received by the feePool
	percentProposer := sdk.NewDecWithPrec(5, 2)
	percentRemaining := sdk.OneDec().Sub(percentProposer)
	feePool = keeper.GetFeePool(ctx)
	expRes := sdk.NewDecFromInt(feeInputs).Mul(percentRemaining)
	require.Equal(t, 1, len(feePool.Pool))
	require.True(sdk.DecEq(t, expRes, feePool.Pool[0].Amount))
}

func TestAllocateFeesWithCommunityTax(t *testing.T) {
	communityTax := sdk.NewDecWithPrec(1, 2) //1%
	ctx, _, keeper, sk, fck := CreateTestInputAdvanced(t, false, 100, communityTax)
	stakeHandler := stake.NewHandler(sk)
	denom := sk.GetParams(ctx).BondDenom

	//first make a validator
	totalPower := int64(10)
	msgCreateValidator := stake.NewTestMsgCreateValidator(valOpAddr1, valConsPk1, totalPower)
	got := stakeHandler(ctx, msgCreateValidator)
	require.True(t, got.IsOK(), "expected msg to be ok, got %v", got)
	_ = sk.ApplyAndReturnValidatorSetUpdates(ctx)

	// allocate 100 denom of fees
	feeInputs := sdk.NewInt(100)
	fck.SetCollectedFees(sdk.Coins{sdk.NewCoin(denom, feeInputs)})
	keeper.AllocateFees(ctx, sdk.OneDec(), valConsAddr1)

	// verify that these fees have been received by the feePool
	feePool := keeper.GetFeePool(ctx)
	// 5% goes to proposer, 1% community tax
	percentProposer := sdk.NewDecWithPrec(5, 2)
	percentRemaining := sdk.OneDec().Sub(communityTax.Add(percentProposer))
	expRes := sdk.NewDecFromInt(feeInputs).Mul(percentRemaining)
	require.Equal(t, 1, len(feePool.Pool))
	require.True(sdk.DecEq(t, expRes, feePool.Pool[0].Amount))
}

func TestAllocateFeesWithPartialPrecommitPower(t *testing.T) {
	communityTax := sdk.NewDecWithPrec(1, 2)
	ctx, _, keeper, sk, fck := CreateTestInputAdvanced(t, false, 100, communityTax)
	stakeHandler := stake.NewHandler(sk)
	denom := sk.GetParams(ctx).BondDenom

	//first make a validator
	totalPower := int64(100)
	msgCreateValidator := stake.NewTestMsgCreateValidator(valOpAddr1, valConsPk1, totalPower)
	got := stakeHandler(ctx, msgCreateValidator)
	require.True(t, got.IsOK(), "expected msg to be ok, got %v", got)
	_ = sk.ApplyAndReturnValidatorSetUpdates(ctx)

	// allocate 100 denom of fees
	feeInputs := sdk.NewInt(100)
	fck.SetCollectedFees(sdk.Coins{sdk.NewCoin(denom, feeInputs)})
	percentPrecommitVotes := sdk.NewDecWithPrec(25, 2)
	keeper.AllocateFees(ctx, percentPrecommitVotes, valConsAddr1)

	// verify that these fees have been received by the feePool
	feePool := keeper.GetFeePool(ctx)
	// 1% + 4%*0.25 to proposer + 1% community tax = 97%
	percentProposer := sdk.NewDecWithPrec(1, 2).Add(sdk.NewDecWithPrec(4, 2).Mul(percentPrecommitVotes))
	percentRemaining := sdk.OneDec().Sub(communityTax.Add(percentProposer))
	expRes := sdk.NewDecFromInt(feeInputs).Mul(percentRemaining)
	require.Equal(t, 1, len(feePool.Pool))
	require.True(sdk.DecEq(t, expRes, feePool.Pool[0].Amount))
}
