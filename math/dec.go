package math

import (
	"math/big"

	"github.com/cockroachdb/apd/v3"

	"cosmossdk.io/errors"
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
	ErrInvalidDec         = errors.Register(mathCodespace, 1, "invalid decimal string")
	ErrUnexpectedRounding = errors.Register(mathCodespace, 2, "unexpected rounding")
	ErrNonIntegeral       = errors.Register(mathCodespace, 3, "value is non-integral")
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

// NewDecFromString converts a string to a Dec type, supporting standard, scientific, and negative notations.
// It handles non-numeric values and overflow conditions, returning errors for invalid inputs like "NaN" or "Infinity".
//
// Examples:
// - "123" -> Dec{123}
// - "-123.456" -> Dec{-123.456}
// - "1.23e4" -> Dec{12300}
// - "NaN" or "Infinity" -> ErrInvalidDec
//
// The internal representation is an arbitrary-precision decimal: Negative × Coeff × 10*Exponent
// The maximum exponent is 100_000 and must not be exceeded. Following values would be invalid:
// 1e100001 -> ErrInvalidDec
// -1e100001 -> ErrInvalidDec
// 1e-100001 -> ErrInvalidDec
//
// This function is essential for converting textual data into Dec types for numerical operations.
func NewDecFromString(s string) (Dec, error) {
	d, _, err := apd.NewFromString(s)
	if err != nil {
		return Dec{}, ErrInvalidDec.Wrap(err.Error())
	}

	switch d.Form {
	case apd.NaN, apd.NaNSignaling:
		return Dec{}, ErrInvalidDec.Wrap("not a number")
	case apd.Infinite:
		return Dec{}, ErrInvalidDec.Wrap(s)
	case apd.Finite:
		result := Dec{*d}
		return result, nil
	default:
		return Dec{}, ErrInvalidDec.Wrapf("unsupported type: %d", d.Form)
	}
}

// NewDecFromInt64 converts an int64 to a Dec type.
// This function is useful for creating Dec values from integer literals or variables,
// ensuring they can be used in high-precision arithmetic operations defined for Dec types.
//
// Example:
// - NewDecFromInt64(123) returns a Dec representing the value 123.
func NewDecFromInt64(x int64) Dec {
	var res Dec
	res.dec.SetInt64(x)
	return res
}

// NewDecWithPrec creates a Dec from a coefficient and exponent, calculated as coeff * 10^exp.
// Useful for precise decimal representations.
//
// Example:
// - NewDecWithPrec(123, -2) -> Dec representing 1.23.
func NewDecWithPrec(coeff int64, exp int32) Dec {
	var res Dec
	res.dec.SetFinite(coeff, exp)
	return res
}

// Add returns a new Dec representing the sum of `x` and `y` using returning a new Dec, we use apd.BaseContext.
// This function ensures that no arguments are mutated during the operation and checks for overflow conditions.
// If an overflow occurs, an error is returned.
//
// The precision is much higher as long as the max exponent is not exceeded. If the max exponent is exceeded, an error is returned.
// For example:
// - 1e100000 + -1e-1
// - 1e100000 + 9e100000
// - 1e100001 + 0
// We can see that in apd.BaseContext the  max exponent is defined hence we cannot exceed.
//
// This function wraps any internal errors with a context-specific error message for clarity.
func (x Dec) Add(y Dec) (Dec, error) {
	var z Dec
	_, err := apd.BaseContext.Add(&z.dec, &x.dec, &y.dec)
	if err != nil {
		return Dec{}, ErrInvalidDec.Wrap(err.Error())
	}

	return z, nil
}

// Sub returns a new Dec representing the sum of `x` and `y` using returning a new Dec, we use apd.BaseContext.
// This function ensures that no arguments are mutated during the operation and checks for overflow conditions.
// If an overflow occurs, an error is returned.
//
// The precision is much higher as long as the max exponent is not exceeded. If the max exponent is exceeded, an error is returned.
// For example:
// - 1e-100001 - 0
// - 1e100000 - 1e-1
// - 1e100000 - -9e100000
// - 1e100001 - 1e100001 (upper limit exceeded)
// - 1e-100001 - 1e-100001 (lower limit exceeded)
// We can see that in apd.BaseContext the  max exponent is defined hence we cannot exceed.
//
// This function wraps any internal errors with a context-specific error message for clarity.
func (x Dec) Sub(y Dec) (Dec, error) {
	var z Dec
	_, err := apd.BaseContext.Sub(&z.dec, &x.dec, &y.dec)
	if err != nil {
		return Dec{}, ErrInvalidDec.Wrap(err.Error())
	}

	return z, errors.Wrap(err, "decimal subtraction error")
}

// Quo performs division of x by y using the decimal128 context with 34 digits of precision.
// It returns a new Dec or an error if the division is not feasible due to constraints of decimal128.
//
// Within Quo half up rounding may be performed to match the defined precision. If this is unwanted, QuoExact
// should be used instead.
//
// Key error scenarios:
// - Division by zero (e.g., `123 / 0` or `0 / 0`) results in ErrInvalidDec.
// - Non-representable values due to extreme ratios or precision limits.
//
// Examples:
// - `0 / 123` yields `0`.
// - `123 / 123` yields `1.000000000000000000000000000000000`.
// - `-123 / 123` yields `-1.000000000000000000000000000000000`.
// - `4 / 9` yields `0.4444444444444444444444444444444444`.
// - `5 / 9` yields `0.5555555555555555555555555555555556`.
// - `6 / 9` yields `0.6666666666666666666666666666666667`.
// - `1e-100000 / 10`  yields error.
//
// This function is non-mutative and enhances error clarity with specific messages.
func (x Dec) Quo(y Dec) (Dec, error) {
	var z Dec
	_, err := dec128Context.Quo(&z.dec, &x.dec, &y.dec)
	if err != nil {
		return Dec{}, ErrInvalidDec.Wrap(err.Error())
	}

	return z, errors.Wrap(err, "decimal quotient error")
}

// MulExact multiplies two Dec values x and y without rounding, using decimal128 precision.
// It returns an error if rounding is necessary to fit the result within the 34-digit limit.
//
// Example:
// - MulExact(Dec{1.234}, Dec{2.345}) -> Dec{2.893}, or ErrUnexpectedRounding if precision exceeded.
//
// Note:
// - This function does not alter the original Dec values.
func (x Dec) MulExact(y Dec) (Dec, error) {
	var z Dec
	condition, err := dec128Context.Mul(&z.dec, &x.dec, &y.dec)
	if err != nil {
		return z, ErrInvalidDec
	}
	if condition.Rounded() {
		return z, ErrUnexpectedRounding
	}

	return z, nil
}

// QuoExact performs division like Quo and additionally checks for rounding. It returns ErrUnexpectedRounding if
// any rounding occurred during the division. If the division is exact, it returns the result without error.
//
// This function is particularly useful in financial calculations or other scenarios where precision is critical
// and rounding could lead to significant errors.
//
// Key error scenarios:
// - Division by zero (e.g., `123 / 0` or `0 / 0`) results in ErrInvalidDec.
// - Rounding would have occurred, which is not permissible in this context, resulting in ErrUnexpectedRounding.
//
// Examples:
// - `0 / 123` yields `0` without rounding.
// - `123 / 123` yields `1.000000000000000000000000000000000` exactly.
// - `-123 / 123` yields `-1.000000000000000000000000000000000` exactly.
// - `1 / 9` yields error for the precision limit
// - `1e-100000 / 10` yields error for crossing the lower exponent limit.
// - Any division resulting in a non-terminating decimal under decimal128 precision constraints triggers ErrUnexpectedRounding.
//
// This function does not mutate any arguments and wraps any internal errors with a context-specific error message for clarity.
func (x Dec) QuoExact(y Dec) (Dec, error) {
	var z Dec
	condition, err := dec128Context.Quo(&z.dec, &x.dec, &y.dec)
	if err != nil {
		return z, ErrInvalidDec.Wrap(err.Error())
	}
	if condition.Rounded() {
		return z, ErrUnexpectedRounding
	}
	return z, errors.Wrap(err, "decimal quotient error")
}

// QuoInteger performs integer division of x by y, returning a new Dec formatted as decimal128 with 34 digit precision.
// This function returns the integer part of the quotient, discarding any fractional part, and is useful in scenarios
// where only the whole number part of the division result is needed without rounding.
//
// Key error scenarios:
// - Division by zero (e.g., `123 / 0`) results in ErrInvalidDec.
// - Overflow conditions if the result exceeds the storage capacity of a decimal128 formatted number.
//
// Examples:
// - `123 / 50` yields `2` (since the fractional part .46 is discarded).
// - `100 / 3` yields `33` (since the fractional part .3333... is discarded).
// - `50 / 100` yields `0` (since 0.5 is less than 1 and thus discarded).
//
// The function does not mutate any arguments and ensures that errors are wrapped with specific messages for clarity.
func (x Dec) QuoInteger(y Dec) (Dec, error) {
	var z Dec
	_, err := dec128Context.QuoInteger(&z.dec, &x.dec, &y.dec)
	if err != nil {
		return z, ErrInvalidDec.Wrap(err.Error())
	}
	return z, nil
}

// Modulo computes the remainder of division of x by y using decimal128 precision.
// It returns an error if y is zero or if any other error occurs during the computation.
func (x Dec) Modulo(y Dec) (Dec, error) {
	var z Dec
	_, err := dec128Context.Rem(&z.dec, &x.dec, &y.dec)
	if err != nil {
		return z, ErrInvalidDec
	}
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

// SdkIntTrim rounds the decimal number towards zero to the nearest integer, then converts and returns it as `sdkmath.Int`.
// It handles both positive and negative values correctly by truncating towards zero.
// This function panics if the resulting integer is larger than the maximum value that `sdkmath.Int` can represent.
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

// Equal checks if the decimal values of x and y are exactly equal.
// It returns true if both decimals represent the same value, otherwise false.
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

// Reduce returns a copy of x with all trailing zeros removed and the number of zeros that were removed.
// It does not modify the original decimal.
func (x Dec) Reduce() (Dec, int) {
	y := Dec{}
	_, n := y.dec.Reduce(&x.dec)
	return y, n
}

// Marshal serializes the decimal value into a byte slice in a text format.
// This method ensures the decimal is represented in a portable and human-readable form.
// The output may be in scientific notation if the number's magnitude is very large or very small.
//
// Returns:
// - A byte slice of the decimal in text format, which may include scientific notation depending on the value.
func (x Dec) Marshal() ([]byte, error) {
	return x.dec.MarshalText()
}

// Unmarshal parses a byte slice containing a text-formatted decimal and stores the result in the receiver.
// It returns an error if the byte slice does not represent a valid decimal.
func (x *Dec) Unmarshal(data []byte) error {
	var d apd.Decimal
	_, _, err := d.SetString(string(data))
	if err != nil {
		return ErrInvalidDec.Wrap(err.Error())
	}

	switch d.Form {
	case apd.NaN, apd.NaNSignaling:
		return ErrInvalidDec.Wrap("not a number")
	case apd.Infinite:
		return ErrInvalidDec.Wrap("infinite decimal value not allowed")
	default:
		x.dec = d
		return nil
	}
}
