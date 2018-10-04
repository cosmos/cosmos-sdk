package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// distribution info for a particular validator
type ValidatorDistInfo struct {
	OperatorAddr sdk.ValAddress `json:"operator_addr"`

	FeePoolWithdrawalHeight int64    `json:"global_withdrawal_height"` // last height this validator withdrew from the global pool
	Pool                    DecCoins `json:"pool"`                     // rewards owed to delegators, commission has already been charged (includes proposer reward)
	PoolCommission          DecCoins `json:"pool_commission"`          // commission collected by this validator (pending withdrawal)

	DelAccum TotalAccum `json:"del_accum"` // total proposer pool accumulation factor held by delegators
}

func NewValidatorDistInfo(operatorAddr sdk.ValAddress, currentHeight int64) ValidatorDistInfo {
	return ValidatorDistInfo{
		OperatorAddr:            operatorAddr,
		FeePoolWithdrawalHeight: currentHeight,
		Pool:           DecCoins{},
		PoolCommission: DecCoins{},
		DelAccum:       NewTotalAccum(currentHeight),
	}
}

// update total delegator accumululation
func (vi ValidatorDistInfo) UpdateTotalDelAccum(height int64, totalDelShares sdk.Dec) ValidatorDistInfo {
	vi.DelAccum = vi.DelAccum.Update(height, totalDelShares)
	return vi
}

// move any available accumulated fees in the FeePool to the validator's pool
func (vi ValidatorDistInfo) TakeFeePoolRewards(fp FeePool, height int64, totalBonded, vdTokens,
	commissionRate sdk.Dec) (ValidatorDistInfo, FeePool) {

	fp = fp.UpdateTotalValAccum(height, totalBonded)

	if fp.ValAccum.Accum.IsZero() {
		return vi, fp
	}

	// update the validators pool
	blocks := height - vi.FeePoolWithdrawalHeight
	vi.FeePoolWithdrawalHeight = height
	accum := sdk.NewDec(blocks).Mul(vdTokens)
	withdrawalTokens := fp.Pool.MulDec(accum).QuoDec(fp.ValAccum.Accum)
	remainingTokens := fp.Pool.Minus(withdrawalTokens)

	commission := withdrawalTokens.MulDec(commissionRate)
	afterCommission := withdrawalTokens.MulDec(sdk.OneDec().Sub(commissionRate))

	fp.ValAccum.Accum = fp.ValAccum.Accum.Sub(accum)
	fp.Pool = remainingTokens
	vi.PoolCommission = vi.PoolCommission.Plus(commission)
	vi.Pool = vi.Pool.Plus(afterCommission)

	return vi, fp
}

// withdraw commission rewards
func (vi ValidatorDistInfo) WithdrawCommission(fp FeePool, height int64,
	totalBonded, vdTokens, commissionRate sdk.Dec) (vio ValidatorDistInfo, fpo FeePool, withdrawn DecCoins) {

	vi, fp = vi.TakeFeePoolRewards(fp, height, totalBonded, vdTokens, commissionRate)

	withdrawalTokens := vi.PoolCommission
	vi.PoolCommission = DecCoins{} // zero

	return vi, fp, withdrawalTokens
}
