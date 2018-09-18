package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// distribution info for a particular validator
type ValidatorDistInfo struct {
	OperatorAddr sdk.ValAddress `json:"operator_addr"`

	GlobalWithdrawalHeight int64    `json:"global_withdrawal_height"` // last height this validator withdrew from the global pool
	Pool                   DecCoins `json:"pool"`                     // rewards owed to delegators, commission has already been charged (includes proposer reward)
	PoolCommission         DecCoins `json:"pool_commission"`          // commission collected by this validator (pending withdrawal)

	DelAccum TotalAccum `json:"del_accum"` // total proposer pool accumulation factor held by delegators
}

// update total delegator accumululation
func (vi ValidatorDistInfo) UpdateTotalDelAccum(height int64, totalDelShares Dec) ValidatorDistInfo {
	vi.DelAccum = vi.DelAccum.Update(height, totalDelShares)
	return vi
}

// move any available accumulated fees in the Global to the validator's pool
func (vi ValidatorDistInfo) TakeGlobalRewards(g Global, height int64, totalBonded, vdTokens, commissionRate Dec) (
	vi ValidatorDistInfo, g Global) {

	g.UpdateTotalValAccum(height, totalBondedShares)

	// update the validators pool
	blocks = height - vi.GlobalWithdrawalHeight
	vi.GlobalWithdrawalHeight = height
	accum = blocks * vdTokens
	withdrawalTokens := g.Pool * accum / g.TotalValAccum
	commission := withdrawalTokens * commissionRate

	g.TotalValAccum -= accumm
	g.Pool -= withdrawalTokens
	vi.PoolCommission += commission
	vi.PoolCommissionFree += withdrawalTokens - commission

	return vi, g
}

// withdraw commission rewards
func (vi ValidatorDistInfo) WithdrawCommission(g Global, height int64,
	totalBonded, vdTokens, commissionRate Dec) (
	vi ValidatorDistInfo, g Global, withdrawn DecCoins) {

	g = vi.TakeGlobalRewards(g, height, totalBonded, vdTokens, commissionRate)

	withdrawalTokens := vi.PoolCommission
	vi.PoolCommission = 0

	return vi, g, withdrawalTokens
}
