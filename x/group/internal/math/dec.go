// Package math provides helper functions for doing mathematical calculations and parsing for the ecocredit module.
package math

import (
	"fmt"

	"github.com/cockroachdb/apd/v2"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/group/errors"
)

// Dec is a wrapper struct around apd.Decimal that does no mutation of apd.Decimal's when performing
// arithmetic, instead creating a new apd.Decimal for every operation ensuring usage is safe.
//
// Using apd.Decimal directly can be unsafe because apd operations mutate the underlying Decimal,
// but when copying the big.Int structure can be shared between Decimal instances causing corruption.
// This was originally discovered in regen0-network/mainnet#15.
type Dec struct {
	dec apd.Decimal
}

func NewDecFromString(s string) (Dec, error) {
	d, _, err := apd.NewFromString(s)
	if err != nil {
		return Dec{}, errors.ErrInvalidDecString.Wrap(err.Error())
	}
	return Dec{*d}, nil
}

func NewDecFromInt64(x int64) Dec {
	var res Dec
	res.dec.SetInt64(x)
	return res
}

// Add returns a new Dec with value `x+y` without mutating any argument and error if
// there is an overflow.
func (x Dec) Add(y Dec) (Dec, error) {
	var z Dec
	_, err := apd.BaseContext.Add(&z.dec, &x.dec, &y.dec)
	return z, sdkerrors.Wrap(err, "decimal addition error")
}

// Sub returns a new Dec with value `x-y` without mutating any argument and error if
// there is an overflow.
func (x Dec) Sub(y Dec) (Dec, error) {
	var z Dec
	_, err := apd.BaseContext.Sub(&z.dec, &x.dec, &y.dec)
	return z, sdkerrors.Wrap(err, "decimal subtraction error")
}

func (x Dec) Int64() (int64, error) {
	return x.dec.Int64()
}

func (x Dec) Cmp(y Dec) int {
	return x.dec.Cmp(&y.dec)
}

func (x Dec) IsEqual(y Dec) bool {
	return x.dec.Cmp(&y.dec) == 0
}

func (x Dec) IsNegative() bool {
	return x.dec.Negative && !x.dec.IsZero()
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
