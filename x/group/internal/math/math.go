// Package math provides helper functions for doing mathematical calculations and parsing for the ecocredit module.
package math

import (
	"fmt"

	"github.com/cockroachdb/apd/v2"

	"github.com/cosmos/cosmos-sdk/types/errors"
)

var exactContext = apd.Context{
	Precision:   0,
	MaxExponent: apd.MaxExponent,
	MinExponent: apd.MinExponent,
	Traps:       apd.DefaultTraps | apd.Inexact | apd.Rounded,
}

// Add adds x and y
func Add(x Dec, y Dec) (Dec, error) {
	return x.Add(y)
}

// SubNonNegative subtracts the value of y from x and returns the result with
// arbitrary precision. Returns an error if the result is negative.
func SubNonNegative(x Dec, y Dec) (Dec, error) {
	z, err := x.Sub(y)
	if err != nil {
		return Dec{}, err
	}

	if z.IsNegative() {
		return z, fmt.Errorf("result negative during non-negative subtraction")
	}

	return z, nil
}

// SafeSubBalance subtracts the value of y from x and returns the result with arbitrary precision.
// Returns with ErrInsufficientFunds error if the result is negative.
func SafeSubBalance(x Dec, y Dec) (Dec, error) {
	var z Dec
	_, err := exactContext.Sub(&z.dec, &x.dec, &y.dec)
	if err != nil {
		return z, errors.Wrap(err, "decimal subtraction error")
	}

	if z.IsNegative() {
		return z, errors.ErrInsufficientFunds
	}

	return z, nil
}

// SafeAddBalance adds the value of x+y and returns the result with arbitrary precision.
// Returns with ErrInvalidRequest error if either x or y is negative.
func SafeAddBalance(x Dec, y Dec) (Dec, error) {
	var z Dec

	if x.IsNegative() || y.IsNegative() {
		return z, errors.Wrap(
			errors.ErrInvalidRequest,
			fmt.Sprintf("AddBalance() requires two non-negative Dec parameters, but received %s and %s", x, y))
	}

	_, err := exactContext.Add(&z.dec, &x.dec, &y.dec)
	if err != nil {
		return z, errors.Wrap(err, "decimal subtraction error")
	}

	return z, nil
}
