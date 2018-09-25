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
	Precision = 10

	// bytes required to represent the above precision
	// ceil(log2(9999999999))
	DecimalPrecisionBits = 34
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
func ZeroDec() Dec { return Dec{new(big.Int).Set(zeroInt)} }
func OneDec() Dec  { return Dec{precisionInt()} }

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
	if len(strs) == 2 {
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

	combined, ok := new(big.Int).SetString(combinedStr, 10)
	if !ok {
		return d, ErrUnknownRequest(fmt.Sprintf("bad string to integer conversion, combinedStr: %v", combinedStr))
	}
	if neg {
		combined = new(big.Int).Neg(combined)
	}
	return Dec{combined}, nil
}

//______________________________________________________________________________________________
//nolint
func (d Dec) IsNil() bool       { return d.Int == nil }                 // is decimal nil
func (d Dec) IsZero() bool      { return (d.Int).Sign() == 0 }          // is equal to zero
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

func (d Dec) String() string {
	str := d.ToLeftPaddedWithDecimals(Precision)
	placement := len(str) - Precision
	if placement < 0 {
		panic("too few decimal digits")
	}
	return str[:placement] + "." + str[placement:]
}

// TODO panic if negative or if totalDigits < len(initStr)???
// evaluate as an integer and return left padded string
func (d Dec) ToLeftPaddedWithDecimals(totalDigits int8) string {
	intStr := d.Int.String()
	fcode := `%0` + strconv.Itoa(int(totalDigits)) + `s`
	return fmt.Sprintf(fcode, intStr)
}

// TODO panic if negative or if totalDigits < len(initStr)???
// evaluate as an integer and return left padded string
func (d Dec) ToLeftPadded(totalDigits int8) string {
	chopped := chopPrecisionAndRoundNonMutative(d.Int)
	intStr := chopped.String()
	fcode := `%0` + strconv.Itoa(int(totalDigits)) + `s`
	return fmt.Sprintf(fcode, intStr)
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

	// get the trucated quotient and remainder
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

// MarshalJSON defines custom encoding scheme
func (d Dec) MarshalJSON() ([]byte, error) {
	if d.Int == nil {
		return nilJSON, nil
	}

	bz, err := d.Int.MarshalText()
	if err != nil {
		return nil, err
	}
	return json.Marshal(string(bz))
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
	return d.Int.UnmarshalText([]byte(text))
}

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
func DecEq(t *testing.T, exp, got Dec) (*testing.T, bool, string, Dec, Dec) {
	return t, exp.Equal(got), "expected:\t%v\ngot:\t\t%v", exp, got
}
