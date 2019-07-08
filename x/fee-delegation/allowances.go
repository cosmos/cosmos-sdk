package fee_delegation

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"time"
)

type BasicFeeAllowance struct {
	// SpendLimit specifies the maximum amount of tokens that can be spent
	// by this capability and will be updated as tokens are spent. If it is
	// empty, there is no spend limit and any amount of coins can be spent.
	SpendLimit sdk.Coins
	// Expiration specifies an optional time when this allowance expires
	Expiration time.Time
}

var _ FeeAllowance = BasicFeeAllowance{}

func (allowance BasicFeeAllowance) Accept(fee sdk.Coins, block abci.Header) (allow bool, updated FeeAllowance, remove bool) {
	if !allowance.Expiration.IsZero() && allowance.Expiration.Before(block.Time) {
		return false, nil, true
	}
	// no spend limit case
	if allowance.SpendLimit == nil {
		return true, nil, false
	}
	left, invalid := allowance.SpendLimit.SafeSub(fee)
	if invalid {
		return false, nil, false
	}
	if left.IsZero() {
		return true, nil, true
	}
	return true, BasicFeeAllowance{SpendLimit: left}, false
}

type PeriodicFeeAllowance struct {
	BasicFeeAllowance
	// Period specifies the time duration in which PeriodSpendLimit coins can
	// be spent before that allowance is reset
	Period time.Duration
	// PeriodSpendLimit specifies the maximum number of coins that can be spent
	// in the Period
	PeriodSpendLimit sdk.Coins
	// PeriodCanSpend is the number of coins left to be spend before the PeriodReset time
	PeriodCanSpend sdk.Coins
	// PeriodReset is the time at which this period resets and a new one begins,
	// it is calculated from the start time of the first transaction after the
	// last period ended
	PeriodReset time.Time
}

func (allowance PeriodicFeeAllowance) Accept(fee sdk.Coins, block abci.Header) (allow bool, updated FeeAllowance, remove bool) {
	allow, basicUpdated, remove := allowance.BasicFeeAllowance.Accept(fee, block)
	if !allow {
		return allow, nil, remove
	}
	periodReset := allowance.PeriodReset
	periodCanSpend := allowance.PeriodCanSpend
	if periodReset.Before(block.Time) {
		periodReset = block.Time.Add(allowance.Period)
		periodCanSpend = allowance.PeriodSpendLimit
	}
	nextPeriodCanSpend, invalid := periodCanSpend.SafeSub(fee)
	if invalid {
		return false, nil, false
	}
	if basicUpdated == nil {
		basicUpdated = allowance.BasicFeeAllowance
	}
	updated = PeriodicFeeAllowance{
		basicUpdated.(BasicFeeAllowance),
		allowance.Period,
		allowance.PeriodSpendLimit,
		nextPeriodCanSpend,
		periodReset,
	}
	return true, updated, remove
}
