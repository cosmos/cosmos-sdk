package types

import (
	"time"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// ExpiresAtTime creates an expiration at the given time
func ExpiresAtTime(t time.Time) ExpiresAt {
	return ExpiresAt{
		Sum: &ExpiresAt_Time{
			Time: &t,
		},
	}
}

// ExpiresAtHeight creates an expiration at the given height
func ExpiresAtHeight(h int64) ExpiresAt {
	return ExpiresAt{
		&ExpiresAt_Height{
			Height: h,
		},
	}
}

// ValidateBasic performs basic sanity checks.
// Note that empty expiration is allowed
func (e ExpiresAt) ValidateBasic() error {
	if e.HasDefinedTime() && e.GetHeight() != 0 {
		return sdkerrors.Wrap(ErrInvalidDuration, "both time and height are set")
	}
	if e.GetHeight() < 0 {
		return sdkerrors.Wrap(ErrInvalidDuration, "negative height")
	}
	return nil
}

// Undefined returns true for an uninitialized struct
func (e ExpiresAt) Undefined() bool {
	return (e.GetTime() == nil || e.GetTime().Unix() <= 0) && e.GetHeight() == 0
}

// HasDefinedTime returns true if `ExpiresAt` has valid time
func (e ExpiresAt) HasDefinedTime() bool {
	t := e.GetTime()
	return t != nil && t.Unix() > 0
}

// FastForward produces a new Expiration with the time or height set to the
// new value, depending on what was set on the original expiration
func (e ExpiresAt) FastForward(t time.Time, h int64) ExpiresAt {
	if e.HasDefinedTime() {
		return ExpiresAtTime(t)
	}
	return ExpiresAtHeight(h)
}

// IsExpired returns if the time or height is *equal to* or greater
// than the defined expiration point. Note that it is expired upon
// an exact match.
//
// Note a "zero" ExpiresAt is never expired
func (e ExpiresAt) IsExpired(t *time.Time, h int64) bool {
	if e.HasDefinedTime() && t.After(*e.GetTime()) {
		return true
	}

	return e.GetHeight() != 0 && h >= e.GetHeight()
}

// IsCompatible returns true iff the two use the same units.
// If false, they cannot be added.
func (e ExpiresAt) IsCompatible(d Duration) bool {
	if e.HasDefinedTime() {
		return d.GetDuration() != nil && d.GetDuration().Seconds() > float64(0)
	}
	return d.GetBlocks() > 0
}

// Step will increase the expiration point by one Duration
// It returns an error if the Duration is incompatible
func (e ExpiresAt) Step(d Duration) (ExpiresAt, error) {
	if !e.IsCompatible(d) {
		return ExpiresAt{}, sdkerrors.Wrap(ErrInvalidDuration, "expiration time and provided duration have different units")
	}
	if e.HasDefinedTime() {
		return ExpiresAtTime(e.GetTime().Add(*d.GetDuration())), nil
	}
	return ExpiresAtHeight(e.GetHeight() + int64(d.GetBlocks())), nil
}

// MustStep is like Step, but panics on error
func (e ExpiresAt) MustStep(d Duration) ExpiresAt {
	res, err := e.Step(d)
	if err != nil {
		panic(err)
	}
	return res
}

// PrepareForExport will deduct the dumpHeight from the expiration, so when this is
// reloaded after a hard fork, the actual number of allowed blocks is constant
func (e ExpiresAt) PrepareForExport(dumpTime time.Time, dumpHeight int64) ExpiresAt {
	if e.GetHeight() != 0 {
		return ExpiresAtHeight(e.GetHeight() - dumpHeight)
	}
	return ExpiresAt{}
}

// ClockDuration creates an Duration by clock time
func ClockDuration(d time.Duration) Duration {
	return Duration{Sum: &Duration_Duration{
		Duration: &d,
	}}
}

// BlockDuration creates an Duration by block height
func BlockDuration(h uint64) Duration {
	return Duration{Sum: &Duration_Blocks{
		Blocks: h,
	}}
}

// ValidateBasic performs basic sanity checks
// Note that exactly one must be set and it must be positive
func (d Duration) ValidateBasic() error {
	if d.GetBlocks() == 0 && d.GetDuration() == nil {
		return sdkerrors.Wrap(ErrInvalidDuration, "neither time and height are set")
	}
	if d.GetBlocks() != 0 && d.GetDuration() != nil && d.GetDuration().Seconds() != float64(0) {
		return sdkerrors.Wrap(ErrInvalidDuration, "both time and height are set")
	}
	if d.GetDuration() != nil && d.GetDuration().Seconds() < 0 {
		return sdkerrors.Wrap(ErrInvalidDuration, "negative clock step")
	}
	return nil
}
