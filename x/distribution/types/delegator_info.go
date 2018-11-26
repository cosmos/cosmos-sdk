package types

import (
	"fmt"

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
	accum := delegatorShares.MulInt(sdk.NewInt(blocks))

	// defensive check
	if accum.IsNegative() {
		panic(fmt.Sprintf("negative accum: %v\n"+
			"\theight: %v\n"+
			"\tdelegation_dist_info: %v\n"+
			"\tdelegator_shares: %v\n",
			accum.String(), height, di, delegatorShares))
	}
	return accum
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

	// Break out to prevent a divide by zero.
	if vi.DelAccum.Accum.IsZero() {
		di.DelPoolWithdrawalHeight = wc.Height
		return di, vi, fp, DecCoins{}
	}

	vi, fp = vi.TakeFeePoolRewards(wc)

	accum := di.GetDelAccum(wc.Height, delegatorShares)
	di.DelPoolWithdrawalHeight = wc.Height

	withdrawalTokens := vi.DelPool.MulDec(accum).QuoDec(vi.DelAccum.Accum)

	// Clip withdrawal tokens by pool, due to possible rounding errors.
	// This rounding error may be introduced upon multiplication since
	// we're clipping decimal digits, and then when we divide by a number ~1 or
	// < 1, the error doesn't get "buried", and if << 1 it'll get amplified.
	// more: https://github.com/cosmos/cosmos-sdk/issues/2888#issuecomment-441387987
	for i, decCoin := range withdrawalTokens {
		poolDenomAmount := vi.DelPool.AmountOf(decCoin.Denom)
		if decCoin.Amount.GT(poolDenomAmount) {
			withdrawalTokens[i] = NewDecCoinFromDec(decCoin.Denom, poolDenomAmount)
		}
	}

	// defensive check for impossible accum ratios
	if accum.GT(vi.DelAccum.Accum) {
		panic(fmt.Sprintf("accum > vi.DelAccum.Accum:\n"+
			"\taccum\t\t\t%v\n"+
			"\tvi.DelAccum.Accum\t%v\n",
			accum, vi.DelAccum.Accum))
	}

	remDelPool := vi.DelPool.Minus(withdrawalTokens)

	// defensive check
	if remDelPool.HasNegative() {
		panic(fmt.Sprintf("negative remDelPool: %v\n"+
			"\tvi.DelPool\t\t%v\n"+
			"\taccum\t\t\t%v\n"+
			"\tvi.DelAccum.Accum\t%v\n"+
			"\twithdrawalTokens\t%v\n",
			remDelPool, vi.DelPool, accum,
			vi.DelAccum.Accum, withdrawalTokens))
	}

	vi.DelPool = remDelPool
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
