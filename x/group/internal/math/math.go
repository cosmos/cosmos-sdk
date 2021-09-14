// Package math provides helper functions for doing mathematical calculations and parsing for the ecocredit module.
package math

import (
	"fmt"

	"github.com/cockroachdb/apd/v2"
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
