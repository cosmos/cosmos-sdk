package types

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"testing"
)

// NOTE: never use new(Dec) or else we will panic unmarshalling into the
// nil embedded big.Int
type Dec struct {
	*big.Int `json:"int"`
}

// number of decimal places
const (
	Precision = 18

	// bytes required to represent the above precision
	// Ceiling[Log2[999 999 999 999 999 999]]
	DecimalPrecisionBits = 60
)

var (
	precisionReuse       = new(big.Int).Exp(big.NewInt(10), big.NewInt(Precision), nil)
	fivePrecision        = new(big.Int).Quo(precisionReuse, big.NewInt(2))
	precisionMultipliers []*big.Int
	zeroInt              = big.NewInt(0)
	oneInt               = big.NewInt(1)
	tenInt               = big.NewInt(10)
)

// Set precision multipliers
func init() {
	precisionMultipliers = make([]*big.Int, Precision+1)
	for i := 0; i <= Precision; i++ {
		precisionMultipliers[i] = calcPrecisionMultiplier(int64(i))
	}
}

func precisionInt() *big.Int {
	return new(big.Int).Set(precisionReuse)
}

// nolint - common values
func ZeroDec() Dec     { return Dec{new(big.Int).Set(zeroInt)} }
func OneDec() Dec      { return Dec{precisionInt()} }
func SmallestDec() Dec { return Dec{new(big.Int).Set(oneInt)} }

// calculate the precision multiplier
func calcPrecisionMultiplier(prec int64) *big.Int {
	if prec > Precision {
		panic(fmt.Sprintf("too much precision, maximum %v, provided %v", Precision, prec))
	}
	zerosToAdd := Precision - prec
	multiplier := new(big.Int).Exp(tenInt, big.NewInt(zerosToAdd), nil)
	return multiplier
}

// get the precision multiplier, do not mutate result
func precisionMultiplier(prec int64) *big.Int {
	if prec > Precision {
		panic(fmt.Sprintf("too much precision, maximum %v, provided %v", Precision, prec))
	}
	return precisionMultipliers[prec]
}

//______________________________________________________________________________________________

// create a new Dec from integer assuming whole number
func NewDec(i int64) Dec {
	return NewDecWithPrec(i, 0)
}

// create a new Dec from integer with decimal place at prec
// CONTRACT: prec <= Precision
func NewDecWithPrec(i, prec int64) Dec {
	return Dec{
		new(big.Int).Mul(big.NewInt(i), precisionMultiplier(prec)),
	}
}

// create a new Dec from big integer assuming whole numbers
// CONTRACT: prec <= Precision
func NewDecFromBigInt(i *big.Int) Dec {
	return NewDecFromBigIntWithPrec(i, 0)
}

// create a new Dec from big integer assuming whole numbers
// CONTRACT: prec <= Precision
func NewDecFromBigIntWithPrec(i *big.Int, prec int64) Dec {
	return Dec{
		new(big.Int).Mul(i, precisionMultiplier(prec)),
	}
}

// create a new Dec from big integer assuming whole numbers
// CONTRACT: prec <= Precision
func NewDecFromInt(i Int) Dec {
	return NewDecFromIntWithPrec(i, 0)
}

// create a new Dec from big integer with decimal place at prec
// CONTRACT: prec <= Precision
func NewDecFromIntWithPrec(i Int, prec int64) Dec {
	return Dec{
		new(big.Int).Mul(i.BigInt(), precisionMultiplier(prec)),
	}
}

// create a decimal from an input decimal string.
// valid must come in the form:
//   (-) whole integers (.) decimal integers
// examples of acceptable input include:
//   -123.456
//   456.7890
//   345
//   -456789
//
// NOTE - An error will return if more decimal places
// are provided in the string than the constant Precision.
//
// CONTRACT - This function does not mutate the input str.
func NewDecFromStr(str string) (d Dec, err Error) {
	if len(str) == 0 {
		return d, ErrUnknownRequest("decimal string is empty")
	}

	// first extract any negative symbol
	neg := false
	if str[0] == '-' {
		neg = true
		str = str[1:]
	}

	if len(str) == 0 {
		return d, ErrUnknownRequest("decimal string is empty")
	}

	strs := strings.Split(str, ".")
	lenDecs := 0
	combinedStr := strs[0]

	if len(strs) == 2 { // has a decimal place
		lenDecs = len(strs[1])
		if lenDecs == 0 || len(combinedStr) == 0 {
			return d, ErrUnknownRequest("bad decimal length")
		}
		combinedStr = combinedStr + strs[1]

	} else if len(strs) > 2 {
		return d, ErrUnknownRequest("too many periods to be a decimal string")
	}

	if lenDecs > Precision {
		return d, ErrUnknownRequest(
			fmt.Sprintf("too much precision, maximum %v, len decimal %v", Precision, lenDecs))
	}

	// add some extra zero's to correct to the Precision factor
	zerosToAdd := Precision - lenDecs
	zeros := fmt.Sprintf(`%0`+strconv.Itoa(zerosToAdd)+`s`, "")
	combinedStr = combinedStr + zeros

	combined, ok := new(big.Int).SetString(combinedStr, 10) // base 10
	if !ok {
		return d, ErrUnknownRequest(fmt.Sprintf("bad string to integer conversion, combinedStr: %v", combinedStr))
	}
	if neg {
		combined = new(big.Int).Neg(combined)
	}
	return Dec{combined}, nil
}

// Decimal from string, panic on error
func MustNewDecFromStr(s string) Dec {
	dec, err := NewDecFromStr(s)
	if err != nil {
		panic(err)
	}
	return dec
}

//______________________________________________________________________________________________
//nolint
func (d Dec) IsNil() bool       { return d.Int == nil }                 // is decimal nil
func (d Dec) IsZero() bool      { return (d.Int).Sign() == 0 }          // is equal to zero
func (d Dec) IsNegative() bool  { return (d.Int).Sign() == -1 }         // is negative
func (d Dec) IsPositive() bool  { return (d.Int).Sign() == 1 }          // is positive
func (d Dec) Equal(d2 Dec) bool { return (d.Int).Cmp(d2.Int) == 0 }     // equal decimals
func (d Dec) GT(d2 Dec) bool    { return (d.Int).Cmp(d2.Int) > 0 }      // greater than
func (d Dec) GTE(d2 Dec) bool   { return (d.Int).Cmp(d2.Int) >= 0 }     // greater than or equal
func (d Dec) LT(d2 Dec) bool    { return (d.Int).Cmp(d2.Int) < 0 }      // less than
func (d Dec) LTE(d2 Dec) bool   { return (d.Int).Cmp(d2.Int) <= 0 }     // less than or equal
func (d Dec) Neg() Dec          { return Dec{new(big.Int).Neg(d.Int)} } // reverse the decimal sign
func (d Dec) Abs() Dec          { return Dec{new(big.Int).Abs(d.Int)} } // absolute value

// addition
func (d Dec) Add(d2 Dec) Dec {
	res := new(big.Int).Add(d.Int, d2.Int)

	if res.BitLen() > 255+DecimalPrecisionBits {
		panic("Int overflow")
	}
	return Dec{res}
}

// subtraction
func (d Dec) Sub(d2 Dec) Dec {
	res := new(big.Int).Sub(d.Int, d2.Int)

	if res.BitLen() > 255+DecimalPrecisionBits {
		panic("Int overflow")
	}
	return Dec{res}
}

// multiplication
func (d Dec) Mul(d2 Dec) Dec {
	mul := new(big.Int).Mul(d.Int, d2.Int)
	chopped := chopPrecisionAndRound(mul)

	if chopped.BitLen() > 255+DecimalPrecisionBits {
		panic("Int overflow")
	}
	return Dec{chopped}
}

// multiplication truncate
func (d Dec) MulTruncate(d2 Dec) Dec {
	mul := new(big.Int).Mul(d.Int, d2.Int)
	chopped := chopPrecisionAndTruncate(mul)

	if chopped.BitLen() > 255+DecimalPrecisionBits {
		panic("Int overflow")
	}
	return Dec{chopped}
}

// multiplication
func (d Dec) MulInt(i Int) Dec {
	mul := new(big.Int).Mul(d.Int, i.i)

	if mul.BitLen() > 255+DecimalPrecisionBits {
		panic("Int overflow")
	}
	return Dec{mul}
}

// MulInt64 - multiplication with int64
func (d Dec) MulInt64(i int64) Dec {
	mul := new(big.Int).Mul(d.Int, big.NewInt(i))

	if mul.BitLen() > 255+DecimalPrecisionBits {
		panic("Int overflow")
	}
	return Dec{mul}
}

// quotient
func (d Dec) Quo(d2 Dec) Dec {

	// multiply precision twice
	mul := new(big.Int).Mul(d.Int, precisionReuse)
	mul.Mul(mul, precisionReuse)

	quo := new(big.Int).Quo(mul, d2.Int)
	chopped := chopPrecisionAndRound(quo)

	if chopped.BitLen() > 255+DecimalPrecisionBits {
		panic("Int overflow")
	}
	return Dec{chopped}
}

// quotient truncate
func (d Dec) QuoTruncate(d2 Dec) Dec {

	// multiply precision twice
	mul := new(big.Int).Mul(d.Int, precisionReuse)
	mul.Mul(mul, precisionReuse)

	quo := new(big.Int).Quo(mul, d2.Int)
	chopped := chopPrecisionAndTruncate(quo)

	if chopped.BitLen() > 255+DecimalPrecisionBits {
		panic("Int overflow")
	}
	return Dec{chopped}
}

// quotient, round up
func (d Dec) QuoRoundUp(d2 Dec) Dec {
	// multiply precision twice
	mul := new(big.Int).Mul(d.Int, precisionReuse)
	mul.Mul(mul, precisionReuse)

	quo := new(big.Int).Quo(mul, d2.Int)
	chopped := chopPrecisionAndRoundUp(quo)

	if chopped.BitLen() > 255+DecimalPrecisionBits {
		panic("Int overflow")
	}
	return Dec{chopped}
}

// quotient
func (d Dec) QuoInt(i Int) Dec {
	mul := new(big.Int).Quo(d.Int, i.i)
	return Dec{mul}
}

// QuoInt64 - quotient with int64
func (d Dec) QuoInt64(i int64) Dec {
	mul := new(big.Int).Quo(d.Int, big.NewInt(i))
	return Dec{mul}
}

// is integer, e.g. decimals are zero
func (d Dec) IsInteger() bool {
	return new(big.Int).Rem(d.Int, precisionReuse).Sign() == 0
}

// format decimal state
func (d Dec) Format(s fmt.State, verb rune) {
	_, err := s.Write([]byte(d.String()))
	if err != nil {
		panic(err)
	}
}

func (d Dec) String() string {
	if d.Int == nil {
		return d.Int.String()
	}

	isNeg := d.IsNegative()
	if d.IsNegative() {
		d = d.Neg()
	}

	bzInt, err := d.Int.MarshalText()
	if err != nil {
		return ""
	}
	inputSize := len(bzInt)

	var bzStr []byte

	// TODO: Remove trailing zeros
	// case 1, purely decimal
	if inputSize <= Precision {
		bzStr = make([]byte, Precision+2)

		// 0. prefix
		bzStr[0] = byte('0')
		bzStr[1] = byte('.')

		// set relevant digits to 0
		for i := 0; i < Precision-inputSize; i++ {
			bzStr[i+2] = byte('0')
		}

		// set final digits
		copy(bzStr[2+(Precision-inputSize):], bzInt)

	} else {

		// inputSize + 1 to account for the decimal point that is being added
		bzStr = make([]byte, inputSize+1)
		decPointPlace := inputSize - Precision

		copy(bzStr, bzInt[:decPointPlace])                   // pre-decimal digits
		bzStr[decPointPlace] = byte('.')                     // decimal point
		copy(bzStr[decPointPlace+1:], bzInt[decPointPlace:]) // post-decimal digits
	}

	if isNeg {
		return "-" + string(bzStr)
	}

	return string(bzStr)
}

//     ____
//  __|    |__   "chop 'em
//       ` \     round!"
// ___||  ~  _     -bankers
// |         |      __
// |       | |   __|__|__
// |_____:  /   | $$$    |
//              |________|

// nolint - go-cyclo
// Remove a Precision amount of rightmost digits and perform bankers rounding
// on the remainder (gaussian rounding) on the digits which have been removed.
//
// Mutates the input. Use the non-mutative version if that is undesired
func chopPrecisionAndRound(d *big.Int) *big.Int {

	// remove the negative and add it back when returning
	if d.Sign() == -1 {
		// make d positive, compute chopped value, and then un-mutate d
		d = d.Neg(d)
		d = chopPrecisionAndRound(d)
		d = d.Neg(d)
		return d
	}

	// get the truncated quotient and remainder
	quo, rem := d, big.NewInt(0)
	quo, rem = quo.QuoRem(d, precisionReuse, rem)

	if rem.Sign() == 0 { // remainder is zero
		return quo
	}

	switch rem.Cmp(fivePrecision) {
	case -1:
		return quo
	case 1:
		return quo.Add(quo, oneInt)
	default: // bankers rounding must take place
		// always round to an even number
		if quo.Bit(0) == 0 {
			return quo
		}
		return quo.Add(quo, oneInt)
	}
}

func chopPrecisionAndRoundUp(d *big.Int) *big.Int {

	// remove the negative and add it back when returning
	if d.Sign() == -1 {
		// make d positive, compute chopped value, and then un-mutate d
		d = d.Neg(d)
		// truncate since d is negative...
		d = chopPrecisionAndTruncate(d)
		d = d.Neg(d)
		return d
	}

	// get the truncated quotient and remainder
	quo, rem := d, big.NewInt(0)
	quo, rem = quo.QuoRem(d, precisionReuse, rem)

	if rem.Sign() == 0 { // remainder is zero
		return quo
	}

	return quo.Add(quo, oneInt)
}

func chopPrecisionAndRoundNonMutative(d *big.Int) *big.Int {
	tmp := new(big.Int).Set(d)
	return chopPrecisionAndRound(tmp)
}

// RoundInt64 rounds the decimal using bankers rounding
func (d Dec) RoundInt64() int64 {
	chopped := chopPrecisionAndRoundNonMutative(d.Int)
	if !chopped.IsInt64() {
		panic("Int64() out of bound")
	}
	return chopped.Int64()
}

// RoundInt round the decimal using bankers rounding
func (d Dec) RoundInt() Int {
	return NewIntFromBigInt(chopPrecisionAndRoundNonMutative(d.Int))
}

//___________________________________________________________________________________

// similar to chopPrecisionAndRound, but always rounds down
func chopPrecisionAndTruncate(d *big.Int) *big.Int {
	return d.Quo(d, precisionReuse)
}

func chopPrecisionAndTruncateNonMutative(d *big.Int) *big.Int {
	tmp := new(big.Int).Set(d)
	return chopPrecisionAndTruncate(tmp)
}

// TruncateInt64 truncates the decimals from the number and returns an int64
func (d Dec) TruncateInt64() int64 {
	chopped := chopPrecisionAndTruncateNonMutative(d.Int)
	if !chopped.IsInt64() {
		panic("Int64() out of bound")
	}
	return chopped.Int64()
}

// TruncateInt truncates the decimals from the number and returns an Int
func (d Dec) TruncateInt() Int {
	return NewIntFromBigInt(chopPrecisionAndTruncateNonMutative(d.Int))
}

// TruncateDec truncates the decimals from the number and returns a Dec
func (d Dec) TruncateDec() Dec {
	return NewDecFromBigInt(chopPrecisionAndTruncateNonMutative(d.Int))
}

// Ceil returns the smallest interger value (as a decimal) that is greater than
// or equal to the given decimal.
func (d Dec) Ceil() Dec {
	tmp := new(big.Int).Set(d.Int)

	quo, rem := tmp, big.NewInt(0)
	quo, rem = quo.QuoRem(tmp, precisionReuse, rem)

	// no need to round with a zero remainder regardless of sign
	if rem.Cmp(zeroInt) == 0 {
		return NewDecFromBigInt(quo)
	}

	if rem.Sign() == -1 {
		return NewDecFromBigInt(quo)
	}

	return NewDecFromBigInt(quo.Add(quo, oneInt))
}

//___________________________________________________________________________________

// MaxSortableDec is the largest Dec that can be passed into SortableDecBytes()
// Its negative form is the least Dec that can be passed in.
var MaxSortableDec = OneDec().Quo(SmallestDec())

// ValidSortableDec ensures that a Dec is within the sortable bounds,
// a Dec can't have a precision of less than 10^-18.
// Max sortable decimal was set to the reciprocal of SmallestDec.
func ValidSortableDec(dec Dec) bool {
	return dec.Abs().LTE(MaxSortableDec)
}

// SortableDecBytes returns a byte slice representation of a Dec that can be sorted.
// Left and right pads with 0s so there are 18 digits to left and right of the decimal point.
// For this reason, there is a maximum and minimum value for this, enforced by ValidSortableDec.
func SortableDecBytes(dec Dec) []byte {
	if !ValidSortableDec(dec) {
		panic("dec must be within bounds")
	}
	// Instead of adding an extra byte to all sortable decs in order to handle max sortable, we just
	// makes its bytes be "max" which comes after all numbers in ASCIIbetical order
	if dec.Equal(MaxSortableDec) {
		return []byte("max")
	}
	// For the same reason, we make the bytes of minimum sortable dec be --, which comes before all numbers.
	if dec.Equal(MaxSortableDec.Neg()) {
		return []byte("--")
	}
	// We move the negative sign to the front of all the left padded 0s, to make negative numbers come before positive numbers
	if dec.IsNegative() {
		return append([]byte("-"), []byte(fmt.Sprintf(fmt.Sprintf("%%0%ds", Precision*2+1), dec.Abs().String()))...)
	}
	return []byte(fmt.Sprintf(fmt.Sprintf("%%0%ds", Precision*2+1), dec.String()))
}

//___________________________________________________________________________________

// reuse nil values
var (
	nilAmino string
	nilJSON  []byte
)

func init() {
	empty := new(big.Int)
	bz, err := empty.MarshalText()
	if err != nil {
		panic("bad nil amino init")
	}
	nilAmino = string(bz)

	nilJSON, err = json.Marshal(string(bz))
	if err != nil {
		panic("bad nil json init")
	}
}

// wraps d.MarshalText()
func (d Dec) MarshalAmino() (string, error) {
	if d.Int == nil {
		return nilAmino, nil
	}
	bz, err := d.Int.MarshalText()
	return string(bz), err
}

// requires a valid JSON string - strings quotes and calls UnmarshalText
func (d *Dec) UnmarshalAmino(text string) (err error) {
	tempInt := new(big.Int)
	err = tempInt.UnmarshalText([]byte(text))
	if err != nil {
		return err
	}
	d.Int = tempInt
	return nil
}

// MarshalJSON marshals the decimal
func (d Dec) MarshalJSON() ([]byte, error) {
	if d.Int == nil {
		return nilJSON, nil
	}

	return json.Marshal(d.String())
}

// UnmarshalJSON defines custom decoding scheme
func (d *Dec) UnmarshalJSON(bz []byte) error {
	if d.Int == nil {
		d.Int = new(big.Int)
	}

	var text string
	err := json.Unmarshal(bz, &text)
	if err != nil {
		return err
	}
	// TODO: Reuse dec allocation
	newDec, err := NewDecFromStr(text)
	if err != nil {
		return err
	}
	d.Int = newDec.Int
	return nil
}

// MarshalYAML returns Ythe AML representation.
func (d Dec) MarshalYAML() (interface{}, error) { return d.String(), nil }

//___________________________________________________________________________________
// helpers

// test if two decimal arrays are equal
func DecsEqual(d1s, d2s []Dec) bool {
	if len(d1s) != len(d2s) {
		return false
	}

	for i, d1 := range d1s {
		if !d1.Equal(d2s[i]) {
			return false
		}
	}
	return true
}

// minimum decimal between two
func MinDec(d1, d2 Dec) Dec {
	if d1.LT(d2) {
		return d1
	}
	return d2
}

// maximum decimal between two
func MaxDec(d1, d2 Dec) Dec {
	if d1.LT(d2) {
		return d2
	}
	return d1
}

// intended to be used with require/assert:  require.True(DecEq(...))
func DecEq(t *testing.T, exp, got Dec) (*testing.T, bool, string, string, string) {
	return t, exp.Equal(got), "expected:\t%v\ngot:\t\t%v", exp.String(), got.String()
}
