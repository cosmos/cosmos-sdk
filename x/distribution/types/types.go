package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// coins with decimal
type DecCoins []DecCoin

// Coins which can have additional decimal points
type DecCoin struct {
	Amount sdk.Dec
	Denom  string
}

// Global state for distribution
type Global struct {
	TotalValAccumUpdateHeight int64    // last height which the total validator accum was updated
	TotalValAccum             sdk.Dec  // total valdator accum held by validators
	Pool                      DecCoins // funds for all validators which have yet to be withdrawn
	CommunityPool             DecCoins // pool for community funds yet to be spent
}

// distribution info for a particular validator
type ValidatorDistInfo struct {
	GlobalWithdrawalHeight int64    // last height this validator withdrew from the global pool
	Pool                   DecCoins // rewards owed to delegators, commission has already been charged (includes proposer reward)
	PoolCommission         DecCoins // commission collected by this validator (pending withdrawal)

	TotalDelAccumUpdateHeight int64   // last height which the total delegator accum was updated
	TotalDelAccum             sdk.Dec // total proposer pool accumulation factor held by delegators
}

// distribution info for a delegation
type DelegatorDistInfo struct {
	WithdrawalHeight int64 // last time this delegation withdrew rewards
}
