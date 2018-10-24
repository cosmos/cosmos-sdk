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
	vi.DelAccum = vi.DelAccum.UpdateForNewHeight(height, totalDelShares)
	return vi
}

// Get the total delegator accum within this validator at the provided height
func (vi ValidatorDistInfo) GetTotalDelAccum(height int64, totalDelShares sdk.Dec) sdk.Dec {
	return vi.DelAccum.GetAccum(height, totalDelShares)
}

// Get the validator accum at the provided height
func (vi ValidatorDistInfo) GetValAccum(height int64, vdTokens sdk.Dec) sdk.Dec {
	blocks := height - vi.FeePoolWithdrawalHeight
	return vdTokens.MulInt(sdk.NewInt(blocks))
}

// Move any available accumulated fees in the FeePool to the validator's pool
// - updates validator info's FeePoolWithdrawalHeight, thus setting accum to 0
// - updates fee pool to latest height and total val accum w/ given totalBonded
// This is the only way to update the FeePool's validator TotalAccum.
// NOTE: This algorithm works as long as TakeFeePoolRewards is called after every power change.
// - called in ValidationDistInfo.WithdrawCommission
// - called in DelegationDistInfo.WithdrawRewards
// NOTE: When a delegator unbonds, say, onDelegationSharesModified ->
//       WithdrawDelegationReward -> WithdrawRewards
func (vi ValidatorDistInfo) TakeFeePoolRewards(fp FeePool, height int64, totalBonded, vdTokens,
	commissionRate sdk.Dec) (ValidatorDistInfo, FeePool) {

	fp = fp.UpdateTotalValAccum(height, totalBonded)

	if fp.TotalValAccum.Accum.IsZero() {
		return vi, fp
	}

	// update the validators pool
	accum := vi.GetValAccum(height, vdTokens)
	vi.FeePoolWithdrawalHeight = height

	if accum.GT(fp.TotalValAccum.Accum) {
		panic("individual accum should never be greater than the total")
	}
	withdrawalTokens := fp.Pool.MulDec(accum).QuoDec(fp.TotalValAccum.Accum)
	remainingTokens := fp.Pool.Minus(withdrawalTokens)

	commission := withdrawalTokens.MulDec(commissionRate)
	afterCommission := withdrawalTokens.Minus(commission)

	fp.TotalValAccum.Accum = fp.TotalValAccum.Accum.Sub(accum)
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

// Estimate the validator's pool rewards at this current state,
// the estimated rewards are subject to fluctuation
func (vi ValidatorDistInfo) EstimatePoolRewards(fp FeePool, height int64,
	totalBonded, vdTokens, commissionRate sdk.Dec) sdk.Coins {

	totalValAccum = fp.GetTotalValAccum(height, totalBonded)
	valAccum := vi.GetValAccum(height, vdTokens)

	if accum.GT(fp.TotalValAccum.Accum) {
		panic("individual accum should never be greater than the total")
	}
	withdrawalTokens := fp.Pool.MulDec(valAccum).QuoDec(totalValAccum)
	commission := withdrawalTokens.MulDec(commissionRate)
	afterCommission := withdrawalTokens.Minus(commission)
	pool := vi.Pool.Plus(afterCommission)
	return pool
}

// Estimate the validator's commission pool rewards at this current state,
// the estimated rewards are subject to fluctuation
func (vi ValidatorDistInfo) EstimateCommissionRewards(fp FeePool, height int64,
	totalBonded, vdTokens, commissionRate sdk.Dec) sdk.Coins {

	totalValAccum = fp.GetTotalValAccum(height, totalBonded)
	valAccum := vi.GetValAccum(height, vdTokens)

	if accum.GT(fp.TotalValAccum.Accum) {
		panic("individual accum should never be greater than the total")
	}
	withdrawalTokens := fp.Pool.MulDec(valAccum).QuoDec(totalValAccum)
	commission := withdrawalTokens.MulDec(commissionRate)
	commissionPool := vi.PoolCommission.Plus(commission)
	return commissionPool
}
