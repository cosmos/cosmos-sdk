package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// distribution info for a delegation - used to determine entitled rewards
type DelegationDistInfo struct {
	DelegatorAddr           sdk.AccAddress `json:"delegator_addr"`
	ValOperatorAddr         sdk.ValAddress `json:"val_operator_addr"`
	DelPoolWithdrawalHeight int64          `json:"del_pool_withdrawal_height"` // last time this delegation withdrew rewards
}

func NewDelegationDistInfo(delegatorAddr sdk.AccAddress, valOperatorAddr sdk.ValAddress,
	currentHeight int64) DelegationDistInfo {

	return DelegationDistInfo{
		DelegatorAddr:           delegatorAddr,
		ValOperatorAddr:         valOperatorAddr,
		DelPoolWithdrawalHeight: currentHeight,
	}
}

// Get the calculated accum of this delegator at the provided height
func (di DelegationDistInfo) GetDelAccum(height int64, delegatorShares sdk.Dec) sdk.Dec {
	blocks := height - di.DelPoolWithdrawalHeight
	return delegatorShares.MulInt(sdk.NewInt(blocks))
}

// Withdraw rewards from delegator.
// Among many things, it does:
// * updates validator info's total del accum
// * calls vi.TakeFeePoolRewards, which:
//   * updates validator info's FeePoolWithdrawalHeight, thus setting accum to 0
//   * updates fee pool to latest height and total val accum w/ given totalBonded
//   (see comment on TakeFeePoolRewards for more info)
func (di DelegationDistInfo) WithdrawRewards(wc WithdrawContext, vi ValidatorDistInfo,
	totalDelShares, delegatorShares sdk.Dec) (
	DelegationDistInfo, ValidatorDistInfo, FeePool, DecCoins) {

	fp := wc.FeePool
	vi = vi.UpdateTotalDelAccum(wc.Height, totalDelShares)

	if vi.DelAccum.Accum.IsZero() {
		return di, vi, fp, DecCoins{}
	}

	vi, fp = vi.TakeFeePoolRewards(wc)

	accum := di.GetDelAccum(wc.Height, delegatorShares)
	di.DelPoolWithdrawalHeight = wc.Height
	withdrawalTokens := vi.DelPool.MulDec(accum).QuoDec(vi.DelAccum.Accum)

	vi.DelPool = vi.DelPool.Minus(withdrawalTokens)
	vi.DelAccum.Accum = vi.DelAccum.Accum.Sub(accum)

	return di, vi, fp, withdrawalTokens
}

// get the delegators rewards at this current state,
func (di DelegationDistInfo) CurrentRewards(wc WithdrawContext, vi ValidatorDistInfo,
	totalDelShares, delegatorShares sdk.Dec) DecCoins {

	totalDelAccum := vi.GetTotalDelAccum(wc.Height, totalDelShares)

	if vi.DelAccum.Accum.IsZero() {
		return DecCoins{}
	}

	rewards := vi.CurrentPoolRewards(wc)
	accum := di.GetDelAccum(wc.Height, delegatorShares)
	tokens := rewards.MulDec(accum).QuoDec(totalDelAccum)
	return tokens
}
