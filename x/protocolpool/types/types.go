package types

import (
	"errors"

	"cosmossdk.io/math"
)

func (cf *ContinuousFund) Validate() error {
	if cf.Recipient == "" {
		return errors.New("recipient cannot be empty")
	}

	// Validate percentage
	if cf.Percentage.IsNil() || cf.Percentage.IsZero() {
		return errors.New("percentage cannot be zero or empty")
	}
	if cf.Percentage.IsNegative() {
		return errors.New("percentage cannot be negative")
	}
	if cf.Percentage.GT(math.LegacyOneDec()) {
		return errors.New("percentage cannot be greater than one")
	}
	return nil
}
