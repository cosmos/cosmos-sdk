package feegrant

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ FeeAllowanceI = (*PeriodicAllowance)(nil)

// Accept can use fee payment requested as well as timestamp of the current block
// to determine whether or not to process this. This is checked in
// Keeper.UseGrantedFees and the return values should match how it is handled there.
//
// If it returns an error, the fee payment is rejected, otherwise it is accepted.
// The FeeAllowance implementation is expected to update it's internal state
// and will be saved again after an acceptance.
//
// If remove is true (regardless of the error), the FeeAllowance will be deleted from storage
// (eg. when it is used up). (See call to RevokeAllowance in Keeper.UseGrantedFees)
func (a *PeriodicAllowance) Accept(ctx sdk.Context, fee sdk.Coins, _ []sdk.Msg) (bool, error) {
	blockTime := ctx.BlockTime()

	if a.Basic.Expiration != nil && blockTime.After(*a.Basic.Expiration) {
		return true, sdkerrors.Wrap(ErrFeeLimitExpired, "absolute limit")
	}

	a.tryResetPeriod(blockTime)

	// deduct from both the current period and the max amount
	var isNeg bool
	a.PeriodCanSpend, isNeg = a.PeriodCanSpend.SafeSub(fee)
	if isNeg {
		return false, sdkerrors.Wrap(ErrFeeLimitExceeded, "period limit")
	}

	if a.Basic.SpendLimit != nil {
		a.Basic.SpendLimit, isNeg = a.Basic.SpendLimit.SafeSub(fee)
		if isNeg {
			return false, sdkerrors.Wrap(ErrFeeLimitExceeded, "absolute limit")
		}

		return a.Basic.SpendLimit.IsZero(), nil
	}

	return false, nil
}

// tryResetPeriod will check if the PeriodReset has been hit. If not, it is a no-op.
// If we hit the reset period, it will top up the PeriodCanSpend amount to
// min(PeriodSpendLimit, Basic.SpendLimit) so it is never more than the maximum allowed.
// It will also update the PeriodReset. If we are within one Period, it will update from the
// last PeriodReset (eg. if you always do one tx per day, it will always reset the same time)
// If we are more then one period out (eg. no activity in a week), reset is one Period from the execution of this method
func (a *PeriodicAllowance) tryResetPeriod(blockTime time.Time) {
	if blockTime.Before(a.PeriodReset) {
		return
	}

	// set PeriodCanSpend to the lesser of Basic.SpendLimit and PeriodSpendLimit
	if _, isNeg := a.Basic.SpendLimit.SafeSub(a.PeriodSpendLimit); isNeg && !a.Basic.SpendLimit.Empty() {
		a.PeriodCanSpend = a.Basic.SpendLimit
	} else {
		a.PeriodCanSpend = a.PeriodSpendLimit
	}

	// If we are within the period, step from expiration (eg. if you always do one tx per day, it will always reset the same time)
	// If we are more then one period out (eg. no activity in a week), reset is one period from this time
	a.PeriodReset = a.PeriodReset.Add(a.Period)
	if blockTime.After(a.PeriodReset) {
		a.PeriodReset = blockTime.Add(a.Period)
	}
}

// ValidateBasic implements FeeAllowance and enforces basic sanity checks
func (a PeriodicAllowance) ValidateBasic() error {
	if err := a.Basic.ValidateBasic(); err != nil {
		return err
	}

	if !a.PeriodSpendLimit.IsValid() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "spend amount is invalid: %s", a.PeriodSpendLimit)
	}
	if !a.PeriodSpendLimit.IsAllPositive() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, "spend limit must be positive")
	}
	if !a.PeriodCanSpend.IsValid() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "can spend amount is invalid: %s", a.PeriodCanSpend)
	}
	// We allow 0 for CanSpend
	if a.PeriodCanSpend.IsAnyNegative() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, "can spend must not be negative")
	}

	// ensure PeriodSpendLimit can be subtracted from total (same coin types)
	if a.Basic.SpendLimit != nil && !a.PeriodSpendLimit.DenomsSubsetOf(a.Basic.SpendLimit) {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, "period spend limit has different currency than basic spend limit")
	}

	// check times
	if a.Period.Seconds() < 0 {
		return sdkerrors.Wrap(ErrInvalidDuration, "negative clock step")
	}

	return nil
}
