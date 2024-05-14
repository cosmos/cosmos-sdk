package math

import (
	"math/big"

	"cosmossdk.io/errors"
	"github.com/cockroachdb/apd/v3"
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

// constants for more convenient intent behind dec.Cmp values.
const (
	GreaterThan = 1
	LessThan    = -1
	EqualTo     = 0
)

const mathCodespace = "math"

var (
	ErrInvalidDecString   = errors.Register(mathCodespace, 1, "invalid decimal string")
	ErrUnexpectedRounding = errors.Register(mathCodespace, 2, "unexpected rounding")
	ErrNonIntegeral       = errors.Register(mathCodespace, 3, "value is non-integral")
	ErrInfiniteString     = errors.Register(mathCodespace, 4, "value is infinite")
)

// In cosmos-sdk#7773, decimal128 (with 34 digits of precision) was suggested for performing
// Quo/Mult arithmetic generically across the SDK. Even though the SDK
// has yet to support a GDA with decimal128 (34 digits), we choose to utilize it here.
// https://github.com/cosmos/cosmos-sdk/issues/7773#issuecomment-725006142
var dec128Context = apd.Context{
	Precision:   34,
	MaxExponent: apd.MaxExponent,
	MinExponent: apd.MinExponent,
	Traps:       apd.DefaultTraps,
}

type SetupConstraint func(Dec) error

// AssertNotNegative greater or equal 0
func AssertNotNegative() SetupConstraint {
	return func(d Dec) error {
		if d.IsNegative() {
			return ErrInvalidDecString.Wrap("is negative")
		}
		return nil
	}
}

// AssertGreaterThanZero greater than 0
func AssertGreaterThanZero() SetupConstraint {
	return func(d Dec) error {
		if !d.IsPositive() {
			return ErrInvalidDecString.Wrap("is negative")
		}
		return nil
	}
}

// AssertMaxDecimals limit the decimal places
func AssertMaxDecimals(max uint32) SetupConstraint {
	return func(d Dec) error {
		if d.NumDecimalPlaces() > max {
			return ErrInvalidDecString.Wrapf("exceeds maximum decimal places: %d", max)
		}
		return nil
	}
}

// NewDecFromString constructor
func NewDecFromString(s string, c ...SetupConstraint) (Dec, error) {
	d, _, err := apd.NewFromString(s)
	if err != nil {
		return Dec{}, ErrInvalidDecString.Wrap(err.Error())
	}

	switch d.Form {
	case apd.NaN, apd.NaNSignaling:
		return Dec{}, ErrInvalidDecString.Wrap("not a number")
	case apd.Infinite:
		return Dec{}, ErrInfiniteString.Wrapf(s)
	default:
		result := Dec{*d}
		for _, v := range c {
			if err := v(result); err != nil {
				return Dec{}, err
			}
		}
		return result, nil
	}
}

// NewNonNegativeDecFromString constructor
func NewNonNegativeDecFromString(s string, c ...SetupConstraint) (Dec, error) {
	return NewDecFromString(s, append(c, AssertNotNegative())...)
}

func NewDecFromInt64(x int64) Dec {
	var res Dec
	res.dec.SetInt64(x)
	return res
}

// NewDecFinite returns a decimal with a value of coeff * 10^exp.
func NewDecFinite(coeff int64, exp int32) Dec {
	var res Dec
	res.dec.SetFinite(coeff, exp)
	return res
}

// Add returns a new Dec with value `x+y` without mutating any argument and error if
// there is an overflow.
func (x Dec) Add(y Dec) (Dec, error) {
	var z Dec
	_, err := apd.BaseContext.Add(&z.dec, &x.dec, &y.dec)
	return z, errors.Wrap(err, "decimal addition error")
}

// Sub returns a new Dec with value `x-y` without mutating any argument and error if
// there is an overflow.
func (x Dec) Sub(y Dec) (Dec, error) {
	var z Dec
	_, err := apd.BaseContext.Sub(&z.dec, &x.dec, &y.dec)
	return z, errors.Wrap(err, "decimal subtraction error")
}

// Quo returns a new Dec with value `x/y` (formatted as decimal128, 34 digit precision) without mutating any
// argument and error if there is an overflow.
func (x Dec) Quo(y Dec) (Dec, error) {
	var z Dec
	_, err := dec128Context.Quo(&z.dec, &x.dec, &y.dec)
	return z, errors.Wrap(err, "decimal quotient error")
}

// MulExact returns a new dec with value x * y. The product must not round or
// ErrUnexpectedRounding will be returned.
func (x Dec) MulExact(y Dec) (Dec, error) {
	var z Dec
	condition, err := dec128Context.Mul(&z.dec, &x.dec, &y.dec)
	if err != nil {
		return z, err
	}
	if condition.Rounded() {
		return z, ErrUnexpectedRounding
	}
	return z, nil
}

// QuoExact is a version of Quo that returns ErrUnexpectedRounding if any rounding occurred.
func (x Dec) QuoExact(y Dec) (Dec, error) {
	var z Dec
	condition, err := dec128Context.Quo(&z.dec, &x.dec, &y.dec)
	if err != nil {
		return z, err
	}
	if condition.Rounded() {
		return z, ErrUnexpectedRounding
	}
	return z, errors.Wrap(err, "decimal quotient error")
}

// QuoInteger returns a new integral Dec with value `x/y` (formatted as decimal128, with 34 digit precision)
// without mutating any argument and error if there is an overflow.
func (x Dec) QuoInteger(y Dec) (Dec, error) {
	var z Dec
	_, err := dec128Context.QuoInteger(&z.dec, &x.dec, &y.dec)
	return z, errors.Wrap(err, "decimal quotient error")
}

// Rem returns the integral remainder from `x/y` (formatted as decimal128, with 34 digit precision) without
// mutating any argument and error if the integer part of x/y cannot fit in 34 digit precision
func (x Dec) Rem(y Dec) (Dec, error) {
	var z Dec
	_, err := dec128Context.Rem(&z.dec, &x.dec, &y.dec)
	return z, errors.Wrap(err, "decimal remainder error")
}

// Mul returns a new Dec with value `x*y` (formatted as decimal128, with 34 digit precision) without
// mutating any argument and error if there is an overflow.
func (x Dec) Mul(y Dec) (Dec, error) {
	var z Dec
	_, err := dec128Context.Mul(&z.dec, &x.dec, &y.dec)
	return z, errors.Wrap(err, "decimal multiplication error")
}

// Int64 converts x to an int64 or returns an error if x cannot
// fit precisely into an int64.
func (x Dec) Int64() (int64, error) {
	return x.dec.Int64()
}

// BigInt converts x to a *big.Int or returns an error if x cannot
// fit precisely into an *big.Int.
func (x Dec) BigInt() (*big.Int, error) {
	y, _ := x.Reduce()
	z := &big.Int{}
	z, ok := z.SetString(y.String(), 10)
	if !ok {
		return nil, ErrNonIntegeral
	}
	return z, nil
}

// SdkIntTrim rounds decimal number to the integer towards zero and converts it to `sdkmath.Int`.
// Panics if x is bigger the SDK Int max value
func (x Dec) SdkIntTrim() Int {
	y, _ := x.Reduce()
	r := y.dec.Coeff
	if y.dec.Exponent != 0 {
		decs := apd.NewBigInt(10)
		if y.dec.Exponent > 0 {
			decs.Exp(decs, apd.NewBigInt(int64(y.dec.Exponent)), nil)
			r.Mul(&y.dec.Coeff, decs)
		} else {
			decs.Exp(decs, apd.NewBigInt(int64(-y.dec.Exponent)), nil)
			r.Quo(&y.dec.Coeff, decs)
		}
	}
	if x.dec.Negative {
		r.Neg(&r)
	}
	return NewIntFromBigInt(r.MathBigInt())
}

func (x Dec) String() string {
	return x.dec.Text('f')
}

// Cmp compares x and y and returns:
// -1 if x <  y
// 0 if x == y
// +1 if x >  y
// undefined if d or x are NaN
func (x Dec) Cmp(y Dec) int {
	return x.dec.Cmp(&y.dec)
}

func (x Dec) Equal(y Dec) bool {
	return x.dec.Cmp(&y.dec) == 0
}

// IsZero returns true if the decimal is zero.
func (x Dec) IsZero() bool {
	return x.dec.IsZero()
}

// IsNegative returns true if the decimal is negative.
func (x Dec) IsNegative() bool {
	return x.dec.Negative && !x.dec.IsZero()
}

// IsPositive returns true if the decimal is positive.
func (x Dec) IsPositive() bool {
	return !x.dec.Negative && !x.dec.IsZero()
}

// IsFinite returns true if the decimal is finite.
func (x Dec) IsFinite() bool {
	return x.dec.Form == apd.Finite
}

// NumDecimalPlaces returns the number of decimal places in x.
func (x Dec) NumDecimalPlaces() uint32 {
	exp := x.dec.Exponent
	if exp >= 0 {
		return 0
	}
	return uint32(-exp)
}

// Reduce returns a copy of x with all trailing zeros removed and the number
// of trailing zeros removed.
func (x Dec) Reduce() (Dec, int) {
	y := Dec{}
	_, n := y.dec.Reduce(&x.dec)
	return y, n
}
