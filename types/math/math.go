package math

import (
	"fmt"

	"github.com/cockroachdb/apd/v2"

	"github.com/cosmos/cosmos-sdk/types/errors"
)

// ParseNonNegativeDecimal parses a non-negative decimal or returns an error.
func ParseNonNegativeDecimal(x string) (*apd.Decimal, error) {
	res, _, err := apd.NewFromString(x)
	if err != nil || res.Sign() < 0 {
		return nil, errors.Wrap(errors.ErrInvalidRequest, fmt.Sprintf("expected a non-negative decimal, got %s", x))
	}

	return res, nil
}

// ParsePositiveDecimal parses a positive decimal or returns an error.
func ParsePositiveDecimal(x string) (*apd.Decimal, error) {
	res, _, err := apd.NewFromString(x)
	if err != nil || res.Sign() <= 0 {
		return nil, errors.Wrap(errors.ErrInvalidRequest, fmt.Sprintf("expected a positive decimal, got %s", x))
	}

	return res, nil
}

// DecimalString prints x as a floating point string.
func DecimalString(x *apd.Decimal) string {
	return x.Text('f')
}

var exactContext = apd.Context{
	Precision:   0,
	MaxExponent: apd.MaxExponent,
	MinExponent: apd.MinExponent,
	Traps:       apd.DefaultTraps | apd.Inexact | apd.Rounded,
}

// Add adds x and y and stores the result in res with arbitrary precision or returns an error.
func Add(res, x, y *apd.Decimal) error {
	_, err := exactContext.Add(res, x, y)
	if err != nil {
		return errors.Wrap(err, "decimal addition error")
	}
	return nil
}

// SafeSub subtracts the value of x from y and stores the result in res with arbitrary precision only
// if the result will be non-negative. An insufficient funds error is returned if the result would be negative.
func SafeSub(res, x, y *apd.Decimal) error {
	_, err := exactContext.Sub(res, x, y)
	if err != nil {
		return errors.Wrap(err, "decimal subtraction error")
	}

	if res.Sign() < 0 {
		return errors.ErrInsufficientFunds
	}

	return nil
}
