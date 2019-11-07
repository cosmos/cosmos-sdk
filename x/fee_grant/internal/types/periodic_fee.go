package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/fee_grant/exported"
)

// PeriodicFeeAllowance extends FeeAllowance to allow for both a maximum cap,
// as well as a limit per time period.
type PeriodicFeeAllowance struct {
	Basic BasicFeeAllowance

	// Period is the duration of one period
	Period Duration
	// PeriodSpendLimit is the maximum amount of tokens to be spent in this period
	PeriodSpendLimit sdk.Coins

	// PeriodCanSpend is how much is available until PeriodReset
	PeriodCanSpend sdk.Coins

	// PeriodRest is when the PeriodCanSpend is updated
	PeriodReset ExpiresAt
}

var _ exported.FeeAllowance = (*PeriodicFeeAllowance)(nil)

// Accept can use fee payment requested as well as timestamp/height of the current block
// to determine whether or not to process this. This is checked in
// Keeper.UseGrantedFees and the return values should match how it is handled there.
//
// If it returns an error, the fee payment is rejected, otherwise it is accepted.
// The FeeAllowance implementation is expected to update it's internal state
// and will be saved again after an acceptance.
//
// If remove is true (regardless of the error), the FeeAllowance will be deleted from storage
// (eg. when it is used up). (See call to RevokeFeeAllowance in Keeper.UseGrantedFees)
func (a *PeriodicFeeAllowance) Accept(fee sdk.Coins, blockTime time.Time, blockHeight int64) (bool, error) {
	if a.Basic.Expiration.IsExpired(blockTime, blockHeight) {
		return true, ErrFeeLimitExpired()
	}

	a.TryResetPeriod(blockTime, blockHeight)

	// deduct from both the current period and the max amount
	var isNeg bool
	a.PeriodCanSpend, isNeg = a.PeriodCanSpend.SafeSub(fee)
	if isNeg {
		return false, ErrFeeLimitExceeded()
	}
	a.Basic.SpendLimit, isNeg = a.Basic.SpendLimit.SafeSub(fee)
	if isNeg {
		return false, ErrFeeLimitExceeded()
	}

	return a.Basic.SpendLimit.IsZero(), nil
}

// TryResetPeriod will check if the PeriodReset has been hit. If not, it is a no-op.
// If we hit the reset period, it will top up the PeriodCanSpend amount to
// min(PeriodicSpendLimit, a.Basic.SpendLimit) so it is never more than the maximum allowed.
// It will also update the PeriodReset. If we are within one Period, it will update from the
// last PeriodReset (eg. if you always do one tx per day, it will always reset the same time)
// If we are more then one period out (eg. no activity in a week), reset is one Period from the execution of this method
func (a *PeriodicFeeAllowance) TryResetPeriod(blockTime time.Time, blockHeight int64) {
	if !a.PeriodReset.IsZero() && !a.PeriodReset.IsExpired(blockTime, blockHeight) {
		return
	}
	// set CanSpend to the lesser of PeriodSpendLimit and the TotalLimit
	if _, isNeg := a.Basic.SpendLimit.SafeSub(a.PeriodSpendLimit); isNeg {
		a.PeriodCanSpend = a.Basic.SpendLimit
	} else {
		a.PeriodCanSpend = a.PeriodSpendLimit
	}

	// If we are within the period, step from expiration (eg. if you always do one tx per day, it will always reset the same time)
	// If we are more then one period out (eg. no activity in a week), reset is one period from this time
	a.PeriodReset = a.PeriodReset.MustStep(a.Period)
	if a.PeriodReset.IsExpired(blockTime, blockHeight) {
		a.PeriodReset = a.PeriodReset.FastForward(blockTime, blockHeight).MustStep(a.Period)
	}
}

// PrepareForExport will adjust the expiration based on export time. In particular,
// it will subtract the dumpHeight from any height-based expiration to ensure that
// the elapsed number of blocks this allowance is valid for is fixed.
// (For PeriodReset and Basic.Expiration)
func (a *PeriodicFeeAllowance) PrepareForExport(dumpTime time.Time, dumpHeight int64) exported.FeeAllowance {
	return &PeriodicFeeAllowance{
		Basic: BasicFeeAllowance{
			SpendLimit: a.Basic.SpendLimit,
			Expiration: a.Basic.Expiration.PrepareForExport(dumpTime, dumpHeight),
		},
		PeriodSpendLimit: a.PeriodSpendLimit,
		PeriodCanSpend:   a.PeriodCanSpend,
		Period:           a.Period,
		PeriodReset:      a.PeriodReset.PrepareForExport(dumpTime, dumpHeight),
	}
}

// ValidateBasic implements FeeAllowance and enforces basic sanity checks
func (a PeriodicFeeAllowance) ValidateBasic() error {
	if err := a.Basic.ValidateBasic(); err != nil {
		return err
	}

	if !a.PeriodSpendLimit.IsValid() {
		return sdk.ErrInvalidCoins("spend amount is invalid: " + a.PeriodSpendLimit.String())
	}
	if !a.PeriodSpendLimit.IsAllPositive() {
		return sdk.ErrInvalidCoins("spend limit must be positive")
	}
	if !a.PeriodCanSpend.IsValid() {
		return sdk.ErrInvalidCoins("can spend amount is invalid: " + a.PeriodCanSpend.String())
	}
	// We allow 0 for CanSpend
	if a.PeriodCanSpend.IsAnyNegative() {
		return sdk.ErrInvalidCoins("can spend must not be negative")
	}

	// TODO: ensure PeriodSpendLimit can be subtracted from total (same coin types)

	// check times
	if err := a.Period.ValidateBasic(); err != nil {
		return err
	}
	return a.PeriodReset.ValidateBasic()
}
