package types

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

//   "that's one big rat!"
//          ______
//         / / /\ \____oo
//     __ /___...._____ _\o
//  __|     |_    |_

// Rat - extend big.Rat
// NOTE: never use new(Rat) or else
// we will panic unmarshalling into the
// nil embedded big.Rat
type Rat struct {
	*big.Rat `json:"rat"`
}

// RatInterface - big Rat with additional functionality
// NOTE: we only have one implementation of this interface
// and don't use it anywhere, but it might come in handy
// if we want to provide Rat types that include
// the units of the value in the type system.
//type RatInterface interface {
//GetRat() *big.Rat
//Num() int64
//Denom() int64
//GT(Rat) bool
//LT(Rat) bool
//Equal(Rat) bool
//IsZero() bool
//Inv() Rat
//Mul(Rat) Rat
//Quo(Rat) Rat
//Add(Rat) Rat
//Sub(Rat) Rat
//Round(int64) Rat
//Evaluate() int64
//}
//var _ Rat = Rat{} // enforce at compile time

// nolint - common values
var (
	ZeroRat = Rat{big.NewRat(0, 1)}
	OneRat  = Rat{big.NewRat(1, 1)}
)

// New - create a new Rat from integers
func NewRat(Numerator int64, Denominator ...int64) Rat {
	switch len(Denominator) {
	case 0:
		return Rat{big.NewRat(Numerator, 1)}
	case 1:
		return Rat{big.NewRat(Numerator, Denominator[0])}
	default:
		panic("improper use of New, can only have one denominator")
	}
}

//NewFromDecimal - create a rational from decimal string or integer string
func NewRatFromDecimal(decimalStr string) (f Rat, err Error) {

	// first extract any negative symbol
	neg := false
	if string(decimalStr[0]) == "-" {
		neg = true
		decimalStr = decimalStr[1:]
	}

	str := strings.Split(decimalStr, ".")

	var numStr string
	var denom int64 = 1
	switch len(str) {
	case 1:
		if len(str[0]) == 0 {
			return f, NewError(CodeUnknownRequest, "not a decimal string")
		}
		numStr = str[0]
	case 2:
		if len(str[0]) == 0 || len(str[1]) == 0 {
			return f, NewError(CodeUnknownRequest, "not a decimal string")
		}
		numStr = str[0] + str[1]
		len := int64(len(str[1]))
		denom = new(big.Int).Exp(big.NewInt(10), big.NewInt(len), nil).Int64()
	default:
		return f, NewError(CodeUnknownRequest, "not a decimal string")
	}

	num, errConv := strconv.Atoi(numStr)
	if errConv != nil {
		return f, NewError(CodeUnknownRequest, errConv.Error())
	}

	if neg {
		num *= -1
	}

	return Rat{big.NewRat(int64(num), denom)}, nil
}

//nolint
func (r Rat) GetRat() *big.Rat  { return r.Rat }                                     // GetRat - get big.Rat
func (r Rat) Num() int64        { return r.Rat.Num().Int64() }                       // Num - return the numerator
func (r Rat) Denom() int64      { return r.Rat.Denom().Int64() }                     // Denom  - return the denominator
func (r Rat) IsZero() bool      { return r.Num() == 0 }                              // IsZero - Is the Rat equal to zero
func (r Rat) Equal(r2 Rat) bool { return r.Rat.Cmp(r2.GetRat()) == 0 }               // Equal - rationals are equal
func (r Rat) GT(r2 Rat) bool    { return r.Rat.Cmp(r2.GetRat()) == 1 }               // GT - greater than
func (r Rat) LT(r2 Rat) bool    { return r.Rat.Cmp(r2.GetRat()) == -1 }              // LT - less than
func (r Rat) Inv() Rat          { return Rat{new(big.Rat).Inv(r.Rat)} }              // Inv - inverse
func (r Rat) Mul(r2 Rat) Rat    { return Rat{new(big.Rat).Mul(r.Rat, r2.GetRat())} } // Mul - multiplication
func (r Rat) Quo(r2 Rat) Rat    { return Rat{new(big.Rat).Quo(r.Rat, r2.GetRat())} } // Quo - quotient
func (r Rat) Add(r2 Rat) Rat    { return Rat{new(big.Rat).Add(r.Rat, r2.GetRat())} } // Add - addition
func (r Rat) Sub(r2 Rat) Rat    { return Rat{new(big.Rat).Sub(r.Rat, r2.GetRat())} } // Sub - subtraction

var zero = big.NewInt(0)
var one = big.NewInt(1)
var two = big.NewInt(2)
var five = big.NewInt(5)
var nFive = big.NewInt(-5)
var ten = big.NewInt(10)

// EvaluateBig - evaluate the rational using bankers rounding
func (r Rat) EvaluateBig() *big.Int {

	num := r.Rat.Num()
	denom := r.Rat.Denom()

	d, rem := new(big.Int), new(big.Int)
	d.QuoRem(num, denom, rem)
	if rem.Cmp(zero) == 0 { // is the remainder zero
		return d
	}

	// evaluate the remainder using bankers rounding
	tenNum := new(big.Int).Mul(num, ten)
	tenD := new(big.Int).Mul(d, ten)
	remainderDigit := new(big.Int).Sub(new(big.Int).Quo(tenNum, denom), tenD) // get the first remainder digit
	isFinalDigit := (new(big.Int).Rem(tenNum, denom).Cmp(zero) == 0)          // is this the final digit in the remainder?

	switch {
	case isFinalDigit && (remainderDigit.Cmp(five) == 0 || remainderDigit.Cmp(nFive) == 0):
		dRem2 := new(big.Int).Rem(d, two)
		return new(big.Int).Add(d, dRem2) // always rounds to the even number
	case remainderDigit.Cmp(five) != -1: //remainderDigit >= 5:
		d.Add(d, one)
	case remainderDigit.Cmp(nFive) != 1: //remainderDigit <= -5:
		d.Sub(d, one)
	}
	return d
}

// Evaluate - evaluate the rational using bankers rounding
func (r Rat) Evaluate() int64 {
	return r.EvaluateBig().Int64()
}

// Round - round Rat with the provided precisionFactor
func (r Rat) Round(precisionFactor int64) Rat {
	rTen := Rat{new(big.Rat).Mul(r.Rat, big.NewRat(precisionFactor, 1))}
	return Rat{big.NewRat(rTen.Evaluate(), precisionFactor)}
}

//___________________________________________________________________________________

var ratCdc JSONCodec // TODO wire.Codec

// Hack to just use json.Marshal for everything until
// we update for amino
type JSONCodec struct{}

func (jc JSONCodec) MarshalJSON(o interface{}) ([]byte, error) {
	return json.Marshal(o)
}

func (jc JSONCodec) UnmarshalJSON(bz []byte, o interface{}) error {
	return json.Unmarshal(bz, o)
}

// Wraps r.MarshalText() in quotes to make it a valid JSON string.
func (r Rat) MarshalJSON() ([]byte, error) {
	bz, err := r.MarshalText()
	if err != nil {
		return bz, err
	}
	return []byte(fmt.Sprintf(`"%s"`, bz)), nil
}

// Requires a valid JSON string - strings quotes and calls UnmarshalText
func (r *Rat) UnmarshalJSON(data []byte) (err error) {
	quote := []byte(`"`)
	if len(data) < 2 ||
		!bytes.HasPrefix(data, quote) ||
		!bytes.HasSuffix(data, quote) {
		return fmt.Errorf("JSON encoded Rat must be a quote-delimitted string")
	}
	data = bytes.Trim(data, `"`)
	return r.UnmarshalText(data)
}
