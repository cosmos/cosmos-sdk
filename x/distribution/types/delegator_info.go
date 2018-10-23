package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// distribution info for a delegation - used to determine entitled rewards
type DelegationDistInfo struct {
	DelegatorAddr    sdk.AccAddress `json:"delegator_addr"`
	ValOperatorAddr  sdk.ValAddress `json:"val_operator_addr"`
	WithdrawalHeight int64          `json:"withdrawal_height"` // last time this delegation withdrew rewards
}

func NewDelegationDistInfo(delegatorAddr sdk.AccAddress, valOperatorAddr sdk.ValAddress,
	currentHeight int64) DelegationDistInfo {

	return DelegationDistInfo{
		DelegatorAddr:    delegatorAddr,
		ValOperatorAddr:  valOperatorAddr,
		WithdrawalHeight: currentHeight,
	}
}

// Withdraw rewards from delegator.
// Among many things, it does:
// * updates validator info's total del accum.
// * calls vi.TakeFeePoolRewards, which:
//   * updates validator info's FeePoolWithdrawalHeight, thus setting accum to 0.
//   * updates fee pool to latest height and total val accum w/ given totalBonded.
//   (see comment on TakeFeePoolRewards for more info).
func (di DelegationDistInfo) WithdrawRewards(fp FeePool, vi ValidatorDistInfo,
	height int64, totalBonded, vdTokens, totalDelShares, delegatorShares,
	commissionRate sdk.Dec) (DelegationDistInfo, ValidatorDistInfo, FeePool, DecCoins) {

	vi = vi.UpdateTotalDelAccum(height, totalDelShares)

	if vi.DelAccum.Accum.IsZero() {
		return di, vi, fp, DecCoins{}
	}

	vi, fp = vi.TakeFeePoolRewards(fp, height, totalBonded, vdTokens, commissionRate)

	blocks := height - di.WithdrawalHeight
	di.WithdrawalHeight = height
	accum := delegatorShares.MulInt(sdk.NewInt(blocks))
	withdrawalTokens := vi.Pool.MulDec(accum).QuoDec(vi.DelAccum.Accum)
	remainingTokens := vi.Pool.Minus(withdrawalTokens)

	vi.Pool = remainingTokens
	vi.DelAccum.Accum = vi.DelAccum.Accum.Sub(accum)

	return di, vi, fp, withdrawalTokens
}
