package types

import (
	"testing"

	"github.com/stretchr/testify/assert"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestWithdrawRewards(t *testing.T) {

	// initialize
	height := int64(0)
	fp := InitialFeePool()
	vi := NewValidatorDistInfo(valAddr1, height)
	commissionRate := sdk.NewDecWithPrec(2, 2)
	validatorTokens := sdk.NewDec(10)
	validatorDelShares := sdk.NewDec(10)
	totalBondedTokens := validatorTokens.Add(sdk.NewDec(90)) // validator-1 is 10% of total power

	di1 := NewDelegationDistInfo(delAddr1, valAddr1, height)
	di1Shares := sdk.NewDec(5) // this delegator has half the shares in the validator

	di2 := NewDelegationDistInfo(delAddr2, valAddr1, height)
	di2Shares := sdk.NewDec(5)

	// simulate adding some stake for inflation
	height = 10
	fp.ValPool = DecCoins{NewDecCoin("stake", 1000)}

	// withdraw rewards
	wc := NewWithdrawContext(fp, height,
		totalBondedTokens, validatorTokens, commissionRate)
	di1, vi, fp, rewardRecv1 := di1.WithdrawRewards(wc, vi,
		validatorDelShares, di1Shares)

	assert.Equal(t, height, di1.DelPoolWithdrawalHeight)
	assert.True(sdk.DecEq(t, sdk.NewDec(900), fp.TotalValAccum.Accum))
	assert.True(sdk.DecEq(t, sdk.NewDec(900), fp.ValPool[0].Amount))
	assert.True(sdk.DecEq(t, sdk.NewDec(49), vi.DelPool[0].Amount))
	assert.True(sdk.DecEq(t, sdk.NewDec(2), vi.ValCommission[0].Amount))
	assert.True(sdk.DecEq(t, sdk.NewDec(49), rewardRecv1[0].Amount))

	// add more blocks and inflation
	height = 20
	fp.ValPool[0].Amount = fp.ValPool[0].Amount.Add(sdk.NewDec(1000))

	// withdraw rewards
	wc = NewWithdrawContext(fp, height,
		totalBondedTokens, validatorTokens, commissionRate)
	di2, vi, fp, rewardRecv2 := di2.WithdrawRewards(wc, vi,
		validatorDelShares, di2Shares)

	assert.Equal(t, height, di2.DelPoolWithdrawalHeight)
	assert.True(sdk.DecEq(t, sdk.NewDec(1800), fp.TotalValAccum.Accum))
	assert.True(sdk.DecEq(t, sdk.NewDec(1800), fp.ValPool[0].Amount))
	assert.True(sdk.DecEq(t, sdk.NewDec(49), vi.DelPool[0].Amount))
	assert.True(sdk.DecEq(t, sdk.NewDec(4), vi.ValCommission[0].Amount))
	assert.True(sdk.DecEq(t, sdk.NewDec(98), rewardRecv2[0].Amount))
}
