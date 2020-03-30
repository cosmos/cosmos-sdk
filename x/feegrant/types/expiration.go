package types

import (
	"time"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// ExpiresAtTime creates an expiration at the given time
func ExpiresAtTime(t time.Time) ExpiresAt {
	return ExpiresAt{Time: t}
}

// ExpiresAtHeight creates an expiration at the given height
func ExpiresAtHeight(h int64) ExpiresAt {
	return ExpiresAt{Height: h}
}

// ValidateBasic performs basic sanity checks.
// Note that empty expiration is allowed
func (e ExpiresAt) ValidateBasic() error {
	if !e.Time.IsZero() && e.Height != 0 {
		return sdkerrors.Wrap(ErrInvalidDuration, "both time and height are set")
	}
	if e.Height < 0 {
		return sdkerrors.Wrap(ErrInvalidDuration, "negative height")
	}
	return nil
}

// IsZero returns true for an uninitialized struct
func (e ExpiresAt) IsZero() bool {
	return e.Time.IsZero() && e.Height == 0
}

// FastForward produces a new Expiration with the time or height set to the
// new value, depending on what was set on the original expiration
func (e ExpiresAt) FastForward(t time.Time, h int64) ExpiresAt {
	if !e.Time.IsZero() {
		return ExpiresAtTime(t)
	}
	return ExpiresAtHeight(h)
}

// IsExpired returns if the time or height is *equal to* or greater
// than the defined expiration point. Note that it is expired upon
// an exact match.
//
// Note a "zero" ExpiresAt is never expired
func (e ExpiresAt) IsExpired(t time.Time, h int64) bool {
	if !e.Time.IsZero() && !t.Before(e.Time) {
		return true
	}
	return e.Height != 0 && h >= e.Height
}

// IsCompatible returns true iff the two use the same units.
// If false, they cannot be added.
func (e ExpiresAt) IsCompatible(d Duration) bool {
	if !e.Time.IsZero() {
		return d.Clock > 0
	}
	return d.Block > 0
}

// Step will increase the expiration point by one Duration
// It returns an error if the Duration is incompatible
func (e ExpiresAt) Step(d Duration) (ExpiresAt, error) {
	if !e.IsCompatible(d) {
		return ExpiresAt{}, sdkerrors.Wrap(ErrInvalidDuration, "expiration time and provided duration have different units")
	}
	if !e.Time.IsZero() {
		e.Time = e.Time.Add(d.Clock)
	} else {
		e.Height += d.Block
	}
	return e, nil
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
	if e.Height != 0 {
		e.Height -= dumpHeight
	}
	return e
}

// ClockDuration creates an Duration by clock time
func ClockDuration(d time.Duration) Duration {
	return Duration{Clock: d}
}

// BlockDuration creates an Duration by block height
func BlockDuration(h int64) Duration {
	return Duration{Block: h}
}

// ValidateBasic performs basic sanity checks
// Note that exactly one must be set and it must be positive
func (d Duration) ValidateBasic() error {
	if d.Block == 0 && d.Clock == 0 {
		return sdkerrors.Wrap(ErrInvalidDuration, "neither time and height are set")
	}
	if d.Block != 0 && d.Clock != 0 {
		return sdkerrors.Wrap(ErrInvalidDuration, "both time and height are set")
	}
	if d.Block < 0 {
		return sdkerrors.Wrap(ErrInvalidDuration, "negative block step")
	}
	if d.Clock < 0 {
		return sdkerrors.Wrap(ErrInvalidDuration, "negative clock step")
	}
	return nil
}
