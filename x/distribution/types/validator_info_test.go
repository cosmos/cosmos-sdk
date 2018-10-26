package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTakeFeePoolRewards(t *testing.T) {

	// initialize
	height := int64(0)
	fp := InitialFeePool()
	vi1 := NewValidatorDistInfo(valAddr1, height)
	vi2 := NewValidatorDistInfo(valAddr2, height)
	vi3 := NewValidatorDistInfo(valAddr3, height)
	commissionRate1 := sdk.NewDecWithPrec(2, 2)
	commissionRate2 := sdk.NewDecWithPrec(3, 2)
	commissionRate3 := sdk.NewDecWithPrec(4, 2)
	validatorTokens1 := sdk.NewDec(10)
	validatorTokens2 := sdk.NewDec(40)
	validatorTokens3 := sdk.NewDec(50)
	totalBondedTokens := validatorTokens1.Add(validatorTokens2).Add(validatorTokens3)

	// simulate adding some stake for inflation
	height = 10
	fp.ValPool = DecCoins{NewDecCoin("stake", 1000)}

	vi1, fp = vi1.TakeFeePoolRewards(fp, height, totalBondedTokens, validatorTokens1, commissionRate1)
	require.True(sdk.DecEq(t, sdk.NewDec(900), fp.TotalValAccum.Accum))
	assert.True(sdk.DecEq(t, sdk.NewDec(900), fp.ValPool[0].Amount))
	assert.True(sdk.DecEq(t, sdk.NewDec(100-2), vi1.DelPool[0].Amount))
	assert.True(sdk.DecEq(t, sdk.NewDec(2), vi1.ValCommission[0].Amount))

	vi2, fp = vi2.TakeFeePoolRewards(fp, height, totalBondedTokens, validatorTokens2, commissionRate2)
	require.True(sdk.DecEq(t, sdk.NewDec(500), fp.TotalValAccum.Accum))
	assert.True(sdk.DecEq(t, sdk.NewDec(500), fp.ValPool[0].Amount))
	assert.True(sdk.DecEq(t, sdk.NewDec(400-12), vi2.DelPool[0].Amount))
	assert.True(sdk.DecEq(t, vi2.ValCommission[0].Amount, sdk.NewDec(12)))

	// add more blocks and inflation
	height = 20
	fp.ValPool[0].Amount = fp.ValPool[0].Amount.Add(sdk.NewDec(1000))

	vi3, fp = vi3.TakeFeePoolRewards(fp, height, totalBondedTokens, validatorTokens3, commissionRate3)
	require.True(sdk.DecEq(t, sdk.NewDec(500), fp.TotalValAccum.Accum))
	assert.True(sdk.DecEq(t, sdk.NewDec(500), fp.ValPool[0].Amount))
	assert.True(sdk.DecEq(t, sdk.NewDec(1000-40), vi3.DelPool[0].Amount))
	assert.True(sdk.DecEq(t, vi3.ValCommission[0].Amount, sdk.NewDec(40)))
}

func TestWithdrawCommission(t *testing.T) {

	// initialize
	height := int64(0)
	fp := InitialFeePool()
	vi := NewValidatorDistInfo(valAddr1, height)
	commissionRate := sdk.NewDecWithPrec(2, 2)
	validatorTokens := sdk.NewDec(10)
	totalBondedTokens := validatorTokens.Add(sdk.NewDec(90)) // validator-1 is 10% of total power

	// simulate adding some stake for inflation
	height = 10
	fp.ValPool = DecCoins{NewDecCoin("stake", 1000)}

	// for a more fun staring condition, have an non-withdraw update
	vi, fp = vi.TakeFeePoolRewards(fp, height, totalBondedTokens, validatorTokens, commissionRate)
	require.True(sdk.DecEq(t, sdk.NewDec(900), fp.TotalValAccum.Accum))
	assert.True(sdk.DecEq(t, sdk.NewDec(900), fp.ValPool[0].Amount))
	assert.True(sdk.DecEq(t, sdk.NewDec(100-2), vi.DelPool[0].Amount))
	assert.True(sdk.DecEq(t, sdk.NewDec(2), vi.ValCommission[0].Amount))

	// add more blocks and inflation
	height = 20
	fp.ValPool[0].Amount = fp.ValPool[0].Amount.Add(sdk.NewDec(1000))

	vi, fp, commissionRecv := vi.WithdrawCommission(fp, height, totalBondedTokens, validatorTokens, commissionRate)
	require.True(sdk.DecEq(t, sdk.NewDec(1800), fp.TotalValAccum.Accum))
	assert.True(sdk.DecEq(t, sdk.NewDec(1800), fp.ValPool[0].Amount))
	assert.True(sdk.DecEq(t, sdk.NewDec(200-4), vi.DelPool[0].Amount))
	assert.Zero(t, len(vi.ValCommission))
	assert.True(sdk.DecEq(t, sdk.NewDec(4), commissionRecv[0].Amount))
}
