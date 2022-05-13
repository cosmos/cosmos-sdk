package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/authz"
	"time"
)

var (
	_ authz.Authorization = &PeriodicSendAuthorization{}
)

// NewPeriodicSendAuthorization creates a new PeriodicSendAuthorization object.
func NewPeriodicSendAuthorization(SpendLimit sdk.Coins, Expiration *time.Time, Period time.Duration, PeriodSpendLimit sdk.Coins, PeriodReset time.Time) *PeriodicSendAuthorization {
	return &PeriodicSendAuthorization{
		SpendLimit:       SpendLimit,
		Expiration:       Expiration,
		Period:           Period,
		PeriodSpendLimit: PeriodSpendLimit,
		PeriodCanSpend:   PeriodSpendLimit,
		PeriodReset:      PeriodReset,
	}
}

// MsgTypeURL implements Authorization.MsgTypeURL.
func (a PeriodicSendAuthorization) MsgTypeURL() string {
	return sdk.MsgTypeURL(&MsgSend{})
}

// The following implementations are heavily based on x/feegrant/periodic_fee.go. Two separate implementations
// are provided so that they can easily diverge in the future.

// Accept implements Authorization.Accept.
func (a PeriodicSendAuthorization) Accept(ctx sdk.Context, msg sdk.Msg) (authz.AcceptResponse, error) {
	mSend, ok := msg.(*MsgSend)
	if !ok {
		return authz.AcceptResponse{}, sdkerrors.Wrap(sdkerrors.ErrInvalidType, "type mismatch")
	}

	blockTime := ctx.BlockTime()

	if a.Expiration != nil && blockTime.After(*a.Expiration) {
		return authz.AcceptResponse{Delete: true}, sdkerrors.Wrap(ErrSpendLimitExceeded, "absolute limit")
	}

	a.tryResetPeriod(blockTime)

	// deduct from both the current period and the max amount
	var isNeg bool
	a.PeriodCanSpend, isNeg = a.PeriodCanSpend.SafeSub(mSend.Amount...)
	if isNeg {
		return authz.AcceptResponse{}, sdkerrors.Wrap(ErrSpendLimitExceeded, "period limit")
	}

	if a.SpendLimit != nil {
		a.SpendLimit, isNeg = a.SpendLimit.SafeSub(mSend.Amount...)
		if isNeg {
			return authz.AcceptResponse{}, sdkerrors.Wrap(ErrSpendLimitExceeded, "absolute limit")
		}

		return authz.AcceptResponse{Delete: a.SpendLimit.IsZero(), Updated: &a}, nil
	}

	return authz.AcceptResponse{Accept: true, Updated: &a}, nil
}

// tryResetPeriod will check if the PeriodReset has been hit. If not, it is a no-op.
// If we hit the reset period, it will top up the PeriodCanSpend amount to
// min(PeriodSpendLimit, Basic.SpendLimit) so it is never more than the maximum allowed.
// It will also update the PeriodReset. If we are within one Period, it will update from the
// last PeriodReset (eg. if you always do one tx per day, it will always reset the same time)
// If we are more then one period out (eg. no activity in a week), reset is one Period from the execution of this method
func (a *PeriodicSendAuthorization) tryResetPeriod(blockTime time.Time) {
	if blockTime.Before(a.PeriodReset) {
		return
	}

	// set PeriodCanSpend to the lesser of Basic.SpendLimit and PeriodSpendLimit
	if _, isNeg := a.SpendLimit.SafeSub(a.PeriodSpendLimit...); isNeg && !a.SpendLimit.Empty() {
		a.PeriodCanSpend = a.SpendLimit
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

// ValidateBasic implements Authorization.ValidateBasic.
func (a PeriodicSendAuthorization) ValidateBasic() error {
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
	if a.SpendLimit != nil && !a.PeriodSpendLimit.DenomsSubsetOf(a.SpendLimit) {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, "period spend limit has different currency than basic spend limit")
	}

	// check times
	if a.Period.Seconds() < 0 {
		return sdkerrors.Wrap(ErrInvalidDuration, "negative clock step")
	}

	return nil
}
