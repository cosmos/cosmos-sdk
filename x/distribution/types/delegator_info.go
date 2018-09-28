package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// distribution info for a delegation
type DelegatorDistInfo struct {
	DelegatorAddr    sdk.AccAddress `json:"delegator_addr"`
	ValOperatorAddr  sdk.ValAddress `json:"val_operator_addr"`
	WithdrawalHeight int64          `json:"withdrawal_height"` // last time this delegation withdrew rewards
}

// withdraw rewards from delegator
func (di DelegatorDistInfo) WithdrawRewards(fp FeePool, vi ValidatorDistInfo,
	height int64, totalBonded, vdTokens, totalDelShares, delegatorShares,
	commissionRate sdk.Dec) (DelegatorDistInfo, FeePool, DecCoins) {

	if vi.DelAccum.Accum.IsZero() {
		return di, fp, DecCoins{}
	}

	vi.UpdateTotalDelAccum(height, totalDelShares)
	vi, fp = vi.TakeFeePoolRewards(fp, height, totalBonded, vdTokens, commissionRate)

	blocks := height - di.WithdrawalHeight
	di.WithdrawalHeight = height
	accum := delegatorShares.Mul(sdk.NewDec(blocks))
	withdrawalTokens := vi.Pool.Mul(accum.Quo(vi.DelAccum.Accum))
	remainingTokens := vi.Pool.Mul(sdk.OneDec().Sub(accum.Quo(vi.DelAccum.Accum)))

	vi.Pool = remainingTokens
	vi.DelAccum.Accum = vi.DelAccum.Accum.Sub(accum)

	return di, fp, withdrawalTokens
}

//_____________________________________________________________________

// withdraw address for the delegation rewards
type DelegatorWithdrawInfo struct {
	DelegatorAddr sdk.AccAddress `json:"delegator_addr"`
	WithdrawAddr  sdk.AccAddress `json:"withdraw_addr"`
}
