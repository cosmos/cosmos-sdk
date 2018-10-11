package keeper

import (
	"fmt"
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
	msgCreateValidator := stake.NewTestMsgCreateValidator(valOpAddr1, valConsPk1, 10)
	got := stakeHandler(ctx, msgCreateValidator)
	require.True(t, got.IsOK(), "expected msg to be ok, got %v", got)
	_ = sk.ApplyAndReturnValidatorSetUpdates(ctx)

	// verify everything has been set in staking correctly
	validator, found := sk.GetValidator(ctx, valOpAddr1)
	require.True(t, found)
	require.Equal(t, sdk.Bonded, validator.Status)
	assert.True(sdk.DecEq(t, sdk.NewDec(10), validator.Tokens))
	assert.True(sdk.DecEq(t, sdk.NewDec(10), validator.DelegatorShares))
	bondedTokens := sk.TotalPower(ctx)
	assert.True(sdk.DecEq(t, sdk.NewDec(10), bondedTokens))

	// initial fee pool should be empty
	feePool := keeper.GetFeePool(ctx)
	require.Nil(t, feePool.Pool)

	// allocate 10 denom of fees
	feeInputs := sdk.NewInt(100)
	fck.SetCollectedFees(sdk.Coins{sdk.NewCoin(denom, feeInputs)})
	require.Equal(t, feeInputs, fck.GetCollectedFees(ctx).AmountOf(denom))
	fmt.Printf("debug valConsAddr1: %v\n", valConsAddr1.String())
	keeper.SetProposerConsAddr(ctx, valConsAddr1)
	keeper.SetSumPrecommitPower(ctx, sdk.NewDec(10))
	keeper.AllocateFees(ctx)

	// verify that these fees have been received by the feePool
	feePool = keeper.GetFeePool(ctx)
	expRes := sdk.NewDecFromInt(feeInputs).Mul(sdk.NewDecWithPrec(95, 2)) // 5% goes to proposer
	require.Equal(t, 1, len(feePool.Pool))
	require.True(sdk.DecEq(t, expRes, feePool.Pool[0].Amount))
}
