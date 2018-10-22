package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	cmn "github.com/tendermint/tendermint/libs/common"
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

// Move any available accumulated fees in the FeePool to the validator's pool.
// * updates validator info's FeePoolWithdrawalHeight, thus setting accum to 0.
// * updates fee pool to latest height and total val accum w/ given totalBonded.
// This is the only way to update the FeePool's validator TotalAccum.
// NOTE: This algorithm works as long as TakeFeePoolRewards is called after every power change.
// - called in ValidationDistInfo.WithdrawCommission.
// - called in DelegationDistInfo.WithdrawRewards.
// NOTE: When a delegator unbonds, say, onDelegationSharesModified ->
//       WithdrawDelegationReward -> WithdrawRewards.
func (vi ValidatorDistInfo) TakeFeePoolRewards(fp FeePool, height int64, totalBonded, vdTokens,
	commissionRate sdk.Dec) (ValidatorDistInfo, FeePool) {

	fp = fp.UpdateTotalValAccum(height, totalBonded)

	if fp.TotalValAccum.Accum.IsZero() {
		return vi, fp
	}

	// update the validators pool
	blocks := height - vi.FeePoolWithdrawalHeight
	vi.FeePoolWithdrawalHeight = height
	accum := vdTokens.MulInt(sdk.NewInt(blocks))
	if accum.GT(fp.TotalValAccum.Accum) {
		panic("individual accum should never be greater than the total")
	}
	withdrawalTokens := fp.Pool.MulDec(accum).QuoDec(fp.TotalValAccum.Accum)
	remainingTokens := fp.Pool.Minus(withdrawalTokens)

	commission := withdrawalTokens.MulDec(commissionRate)
	afterCommission := withdrawalTokens.Minus(commission)

	fmt.Println(
		cmn.Red(
			fmt.Sprintf("FP Sub %v * %v = %v -=> %v",
				vdTokens, sdk.NewInt(blocks),
				accum,
				fp.TotalValAccum.Accum.Sub(accum),
			),
		),
	)

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
