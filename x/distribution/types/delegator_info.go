package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// distribution info for a delegation
type DelegatorDistInfo struct {
	DelegatorAddr    sdk.AccAddress
	ValOperatorAddr  sdk.ValAddress
	WithdrawalHeight int64 // last time this delegation withdrew rewards
}

// withdraw rewards from delegator
func (di DelegatorDistInfo) WithdrawRewards(g Global, vi ValidatorDistInfo,
	height int64, totalBonded, vdTokens, totalDelShares, commissionRate Dec) (
	di DelegatorDistInfo, g Global, withdrawn DecCoins) {

	vi.UpdateTotalDelAccum(height, totalDelShares)
	g = vi.TakeGlobalRewards(g, height, totalBonded, vdTokens, commissionRate)

	blocks = height - di.WithdrawalHeight
	di.WithdrawalHeight = height
	accum = delegatorShares * blocks

	withdrawalTokens := vi.Pool * accum / vi.TotalDelAccum
	vi.TotalDelAccum -= accum

	vi.Pool -= withdrawalTokens
	vi.TotalDelAccum -= accum
	return di, g, withdrawalTokens
}
