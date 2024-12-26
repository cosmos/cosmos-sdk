package math

import (
	"encoding/json"
	stderrors "errors"
	"math/big"
	"strconv"

	"github.com/cockroachdb/apd/v3"

	"cosmossdk.io/errors"
)

var _ customProtobufType = &Dec{}

const (
	// MaxExponent is the highest exponent supported. Exponents near this range will
	// perform very slowly (many seconds per operation).
	MaxExponent = apd.MaxExponent
	// MinExponent is the lowest exponent supported with the same limitations as
	// MaxExponent.
	MinExponent = apd.MinExponent
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

const mathCodespace = "math"

var (
	ErrInvalidDec         = errors.Register(mathCodespace, 1, "invalid decimal")
	ErrUnexpectedRounding = errors.Register(mathCodespace, 2, "unexpected rounding")
	ErrNonIntegral        = errors.Register(mathCodespace, 3, "value is non-integral")
)

// In cosmos-sdk#7773, decimal128 (with 34 digits of precision) was suggested for performing
// Quo/Mult arithmetic generically across the SDK. Even though the SDK
// has yet to support a GDA with decimal128 (34 digits), we choose to utilize it here.
// https://github.com/cosmos/cosmos-sdk/issues/7773#issuecomment-725006142
var dec128Context = apd.Context{
	Precision:   34,
	MaxExponent: MaxExponent,
	MinExponent: MinExponent,
	Traps:       apd.DefaultTraps,
}

// NewDecFromString converts a string to a Dec type, supporting standard, scientific, and negative notations.
// It handles non-numeric values and overflow conditions, returning errors for invalid inputs like "NaN" or "Infinity".
//
// Examples:
// - "123" -> Dec{123}
// - "-123.456" -> Dec{-123.456}
// - "1.23E4" -> Dec{12300}
// - "NaN" or "Infinity" -> ErrInvalidDec
//
// The internal representation is an arbitrary-precision decimal: Negative × Coeff × 10*Exponent
// The maximum exponent is 100_000 and must not be exceeded. Following values would be invalid:
// 1E100001 -> ErrInvalidDec
// -1E100001 -> ErrInvalidDec
// 1E-100001 -> ErrInvalidDec
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

// NewDecWithExp creates a Dec from a coefficient and exponent, calculated as coeff * 10^exp.
// Useful for precise decimal representations.
// Although this method can be used with a higher than maximum exponent or lower than minimum exponent, further arithmetic
// or other method may fail.
//
// Example:
// - NewDecWithExp(123, -2) -> Dec representing 1.23.
func NewDecWithExp(coeff int64, exp int32) Dec {
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
		if err2 := stderrors.Unwrap(err); err2 != nil {
			// use unwrapped error to not return "add:" prefix from raw apd error
			err = err2
		}
		return Dec{}, ErrInvalidDec.Wrap("sub: " + err.Error())
	}
	return z, nil
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

// Mul returns a new Dec with value `x*y` (formatted as decimal128, with 34 digit precision) without
// mutating any argument and error if there is an overflow.
func (x Dec) Mul(y Dec) (Dec, error) {
	var z Dec
	if _, err := dec128Context.Mul(&z.dec, &x.dec, &y.dec); err != nil {
		return z, ErrInvalidDec.Wrap(err.Error())
	}
	return z, nil
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
		return z, ErrInvalidDec.Wrap(err.Error())
	}
	if condition.Rounded() {
		return z, ErrUnexpectedRounding
	}

	return z, nil
}

// Modulo computes the remainder of division of x by y using decimal128 precision.
// It returns an error if y is zero or if any other error occurs during the computation.
//
// Example:
//   - 7 mod 3 = 1
//   - 6 mod 3 = 0
func (x Dec) Modulo(y Dec) (Dec, error) {
	var z Dec
	_, err := dec128Context.Rem(&z.dec, &x.dec, &y.dec)
	if err != nil {
		return z, ErrInvalidDec.Wrap(err.Error())
	}
	return z, errors.Wrap(err, "decimal remainder error")
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
	z, ok := new(big.Int).SetString(y.Text('f'), 10)
	if !ok {
		return nil, ErrNonIntegral
	}
	return z, nil
}

// SdkIntTrim rounds the decimal number towards zero to the nearest integer, then converts and returns it as `sdkmath.Int`.
// It handles both positive and negative values correctly by truncating towards zero.
// This function returns an ErrNonIntegral error if the resulting integer is larger than the maximum value that `sdkmath.Int` can represent.
func (x Dec) SdkIntTrim() (Int, error) {
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
	bigInt := r.MathBigInt()
	if bigInt.BitLen() > MaxBitLen {
		return ZeroInt(), ErrNonIntegral
	}
	return NewIntFromBigInt(bigInt), nil
}

// String formatted in decimal notation: '-ddddd.dddd', no exponent
func (x Dec) String() string {
	return string(fmtE(x.dec, 'E'))
}

// Text converts the floating-point number x to a string according
// to the given format. The format is one of:
//
//	'e'	-d.dddde±dd, decimal exponent, exponent digits
//	'E'	-d.ddddE±dd, decimal exponent, exponent digits
//	'f'	-ddddd.dddd, no exponent
//	'g'	like 'e' for large exponents, like 'f' otherwise
//	'G'	like 'E' for large exponents, like 'f' otherwise
//
// If format is a different character, Text returns a "%" followed by the
// unrecognized.Format character. The 'f' format has the possibility of
// displaying precision that is not present in the Decimal when it appends
// zeros (the 'g' format avoids the use of 'f' in this case). All other
// formats always show the exact precision of the Decimal.
func (x Dec) Text(format byte) string {
	return x.dec.Text(format)
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

// Marshal serializes the decimal value into a byte slice in text format.
// This method represents the decimal in a portable and compact hybrid notation.
// Based on the exponent value, the number is formatted into decimal: -ddddd.ddddd, no exponent
// or scientific notation: -d.ddddE±dd
//
// For example, the following transformations are made:
//   - 0 -> 0
//   - 123 -> 123
//   - 10000 -> 10000
//   - -0.001 -> -0.001
//   - -0.000000001 -> -1E-9
//
// Returns:
//   - A byte slice of the decimal in text format.
//   - An error if the decimal cannot be reduced or marshaled properly.
func (x Dec) Marshal() ([]byte, error) {
	var d apd.Decimal
	if _, _, err := dec128Context.Reduce(&d, &x.dec); err != nil {
		return nil, ErrInvalidDec.Wrap(err.Error())
	}
	return fmtE(d, 'E'), nil
}

// fmtE formats a decimal number into a byte slice in scientific notation or fixed-point notation depending on the exponent.
// If the adjusted exponent is between -6 and 6 inclusive, it uses fixed-point notation, otherwise it uses scientific notation.
func fmtE(d apd.Decimal, fmt byte) []byte {
	var scratch, dest [16]byte
	buf := dest[:0]
	digits := d.Coeff.Append(scratch[:0], 10)
	totalDigits := int64(len(digits))
	adj := int64(d.Exponent) + totalDigits - 1
	if adj > -6 && adj < 6 {
		return []byte(d.Text('f'))
	}
	switch {
	case totalDigits > 5:
		beforeComma := digits[0 : totalDigits-6]
		adj -= int64(len(beforeComma) - 1)
		buf = append(buf, beforeComma...)
		buf = append(buf, '.')
		buf = append(buf, digits[totalDigits-6:]...)
	case totalDigits > 1:
		buf = append(buf, digits[0])
		buf = append(buf, '.')
		buf = append(buf, digits[1:]...)
	default:
		buf = append(buf, digits[0:]...)
	}

	buf = append(buf, fmt)
	var ch byte
	if adj < 0 {
		ch = '-'
		adj = -adj
	} else {
		ch = '+'
	}
	buf = append(buf, ch)
	return strconv.AppendInt(buf, adj, 10)
}

// Unmarshal parses a byte slice containing a text-formatted decimal and stores the result in the receiver.
// It returns an error if the byte slice does not represent a valid decimal.
func (x *Dec) Unmarshal(data []byte) error {
	result, err := NewDecFromString(string(data))
	if err != nil {
		return ErrInvalidDec.Wrap(err.Error())
	}

	if result.dec.Form != apd.Finite {
		return ErrInvalidDec.Wrap("unknown decimal form")
	}

	x.dec = result.dec
	return nil
}

// MarshalTo encodes the receiver into the provided byte slice and returns the number of bytes written and any error encountered.
func (x Dec) MarshalTo(data []byte) (n int, err error) {
	bz, err := x.Marshal()
	if err != nil {
		return 0, err
	}

	return copy(data, bz), nil
}

// Size returns the number of bytes required to encode the Dec value, which is useful for determining storage requirements.
func (x Dec) Size() int {
	bz, _ := x.Marshal()
	return len(bz)
}

// MarshalJSON serializes the Dec struct into a JSON-encoded byte slice using scientific notation.
func (x Dec) MarshalJSON() ([]byte, error) {
	return json.Marshal(fmtE(x.dec, 'E'))
}

// UnmarshalJSON implements the json.Unmarshaler interface for the Dec type, converting JSON strings to Dec objects.
func (x *Dec) UnmarshalJSON(data []byte) error {
	var text string
	err := json.Unmarshal(data, &text)
	if err != nil {
		return err
	}
	val, err := NewDecFromString(text)
	if err != nil {
		return err
	}
	*x = val
	return nil
}
