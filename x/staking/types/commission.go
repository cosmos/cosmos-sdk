// Package types defines the core data structures for the staking module.
// This file provides commission types and validation logic for validator commission rates,
// including initial commission setup and rate change validation with time-based restrictions.
package types

import (
	"time"

	"cosmossdk.io/math"
)

// NewCommissionRates creates a new CommissionRates instance with the given rate, maximum rate,
// and maximum change rate. These rates define the validator's commission structure and constraints.
func NewCommissionRates(rate, maxRate, maxChangeRate math.LegacyDec) CommissionRates {
	return CommissionRates{
		Rate:          rate,
		MaxRate:       maxRate,
		MaxChangeRate: maxChangeRate,
	}
}

// NewCommission creates a new Commission instance with the given rates and sets the update time
// to the zero time. This is typically used when initializing a new validator's commission.
func NewCommission(rate, maxRate, maxChangeRate math.LegacyDec) Commission {
	return Commission{
		CommissionRates: NewCommissionRates(rate, maxRate, maxChangeRate),
		UpdateTime:      time.Unix(0, 0).UTC(),
	}
}

// NewCommissionWithTime creates a new Commission instance with the given rates and update time.
// The update time should be set to the current block BFT time when updating an existing commission.
func NewCommissionWithTime(rate, maxRate, maxChangeRate math.LegacyDec, updatedAt time.Time) Commission {
	return Commission{
		CommissionRates: NewCommissionRates(rate, maxRate, maxChangeRate),
		UpdateTime:      updatedAt,
	}
}

// Validate performs validation checks on the commission rates to ensure they are within
// acceptable bounds. It checks that rates are non-negative, max rate is not greater than 1,
// rate does not exceed max rate, and max change rate constraints are satisfied.
// Returns an error if any validation check fails.
func (cr CommissionRates) Validate() error {
	switch {
	case cr.MaxRate.IsNegative():
		// max rate cannot be negative
		return ErrCommissionNegative

	case cr.MaxRate.GT(math.LegacyOneDec()):
		// max rate cannot be greater than 1
		return ErrCommissionHuge

	case cr.Rate.IsNegative():
		// rate cannot be negative
		return ErrCommissionNegative

	case cr.Rate.GT(cr.MaxRate):
		// rate cannot be greater than the max rate
		return ErrCommissionGTMaxRate

	case cr.MaxChangeRate.IsNegative():
		// change rate cannot be negative
		return ErrCommissionChangeRateNegative

	case cr.MaxChangeRate.GT(cr.MaxRate):
		// change rate cannot be greater than the max rate
		return ErrCommissionChangeRateGTMaxRate
	}

	return nil
}

// ValidateNewRate validates a new commission rate against the existing commission constraints.
// It checks that the rate change respects the 24-hour update interval, the new rate is within
// bounds, and the change does not exceed the maximum allowed change rate. Returns an error
// if any validation check fails.
func (c Commission) ValidateNewRate(newRate math.LegacyDec, blockTime time.Time) error {
	switch {
	case blockTime.Sub(c.UpdateTime).Hours() < 24:
		// new rate cannot be changed more than once within 24 hours
		return ErrCommissionUpdateTime

	case newRate.IsNegative():
		// new rate cannot be negative
		return ErrCommissionNegative

	case newRate.GT(c.MaxRate):
		// new rate cannot be greater than the max rate
		return ErrCommissionGTMaxRate

	case newRate.Sub(c.Rate).GT(c.MaxChangeRate):
		// new rate % points change cannot be greater than the max change rate
		return ErrCommissionGTMaxChangeRate
	}

	return nil
}
