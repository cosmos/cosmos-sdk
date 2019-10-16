package types

import "time"

// ExpiresAt is a point in time where something expires.
// It may be *either* block time or block height
type ExpiresAt struct {
	Time   time.Time `json:"time" yaml:"time"`
	Height int64     `json:"height" yaml:"height"`
}

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
		return ErrInvalidPeriod("both time and height are set")
	}
	if e.Height < 0 {
		return ErrInvalidPeriod("negative height")
	}
	return nil
}

// IsZero returns true for an uninitialized struct
func (e ExpiresAt) IsZero() bool {
	return e.Time.IsZero() && e.Height == 0
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
func (e ExpiresAt) IsCompatible(p Period) bool {
	if !e.Time.IsZero() {
		return p.Clock > 0
	}
	return p.Block > 0
}

// Step will increase the expiration point by one period
// It returns an error if the period is incompatible
func (e *ExpiresAt) Step(p Period) error {
	if !e.IsCompatible(p) {
		return ErrInvalidPeriod("expires_at and period have different units")
	}
	if !e.Time.IsZero() {
		e.Time = e.Time.Add(p.Clock)
	} else {
		e.Height = e.Height + p.Block
	}
	return nil
}

// Period is a repeating unit of either clock time or number of blocks.
// This is designed to be added to an ExpiresAt struct.
type Period struct {
	Clock time.Duration `json:"clock" yaml:"clock"`
	Block int64         `json:"block" yaml:"block"`
}

// ClockPeriod creates an period by clock time
func ClockPeriod(d time.Duration) Period {
	// assert nothing negative
	if d < 0 {
		panic("Cannot use a negative duration")
	}
	return Period{Clock: d}
}

// BlockPeriod creates an period by block height
func BlockPeriod(h int64) Period {
	// assert nothing negative
	if h < 0 {
		panic("Cannot use a negative block step")
	}
	return Period{Block: h}
}

// ValidateBasic performs basic sanity checks
// Note that exactly one must be set and it must be positive
func (p Period) ValidateBasic() error {
	if p.Block == 0 && p.Clock == 0 {
		return ErrInvalidPeriod("neither time and height are set")
	}
	if p.Block != 0 && p.Clock != 0 {
		return ErrInvalidPeriod("both time and height are set")
	}
	if p.Block < 0 {
		return ErrInvalidPeriod("negative block step")
	}
	if p.Clock < 0 {
		return ErrInvalidPeriod("negative clock step")
	}
	return nil
}
