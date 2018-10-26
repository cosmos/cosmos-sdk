package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// distribution info for a particular validator
type ValidatorDistInfo struct {
	OperatorAddr sdk.ValAddress `json:"operator_addr"`

	FeePoolWithdrawalHeight int64 `json:"fee_pool_withdrawal_height"` // last height this validator withdrew from the global pool

	DelAccum      TotalAccum `json:"del_accum"`      // total accumulation factor held by delegators
	DelPool       DecCoins   `json:"del_pool"`       // rewards owed to delegators, commission has already been charged (includes proposer reward)
	ValCommission DecCoins   `json:"val_commission"` // commission collected by this validator (pending withdrawal)
}

func NewValidatorDistInfo(operatorAddr sdk.ValAddress, currentHeight int64) ValidatorDistInfo {
	return ValidatorDistInfo{
		OperatorAddr:            operatorAddr,
		FeePoolWithdrawalHeight: currentHeight,
		DelPool:                 DecCoins{},
		DelAccum:                NewTotalAccum(currentHeight),
		ValCommission:           DecCoins{},
	}
}

// update total delegator accumululation
func (vi ValidatorDistInfo) UpdateTotalDelAccum(height int64, totalDelShares sdk.Dec) ValidatorDistInfo {
	vi.DelAccum = vi.DelAccum.UpdateForNewHeight(height, totalDelShares)
	return vi
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
		vi.FeePoolWithdrawalHeight = height
		return vi, fp
	}

	// update the validators pool
	blocks := height - vi.FeePoolWithdrawalHeight
	vi.FeePoolWithdrawalHeight = height
	accum := vdTokens.MulInt(sdk.NewInt(blocks))

	if accum.GT(fp.TotalValAccum.Accum) {
		panic("individual accum should never be greater than the total")
	}
	withdrawalTokens := fp.ValPool.MulDec(accum).QuoDec(fp.TotalValAccum.Accum)
	remValPool := fp.ValPool.Minus(withdrawalTokens)

	commission := withdrawalTokens.MulDec(commissionRate)
	afterCommission := withdrawalTokens.Minus(commission)

	fp.TotalValAccum.Accum = fp.TotalValAccum.Accum.Sub(accum)
	fp.ValPool = remValPool
	vi.ValCommission = vi.ValCommission.Plus(commission)
	vi.DelPool = vi.DelPool.Plus(afterCommission)

	return vi, fp
}

// withdraw commission rewards
func (vi ValidatorDistInfo) WithdrawCommission(fp FeePool, height int64,
	totalBonded, vdTokens, commissionRate sdk.Dec) (vio ValidatorDistInfo, fpo FeePool, withdrawn DecCoins) {

	vi, fp = vi.TakeFeePoolRewards(fp, height, totalBonded, vdTokens, commissionRate)

	withdrawalTokens := vi.ValCommission
	vi.ValCommission = DecCoins{} // zero

	return vi, fp, withdrawalTokens
}
