package types

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"testing"
)

//   "that's one big rat!"
//          ______
//         / / /\ \____oo
//     __ /___...._____ _\o
//  __|     |_    |_

// NOTE: never use new(Rat) or else
// we will panic unmarshalling into the
// nil embedded big.Rat
type Rat struct {
	big.Rat `json:"rat"`
}

// nolint - common values
func ZeroRat() Rat { return Rat{*big.NewRat(0, 1)} }
func OneRat() Rat  { return Rat{*big.NewRat(1, 1)} }

// New - create a new Rat from integers
func NewRat(Numerator int64, Denominator ...int64) Rat {
	switch len(Denominator) {
	case 0:
		return Rat{*big.NewRat(Numerator, 1)}
	case 1:
		return Rat{*big.NewRat(Numerator, Denominator[0])}
	default:
		panic("improper use of New, can only have one denominator")
	}
}

// create a rational from decimal string or integer string
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
			return f, ErrUnknownRequest("not a decimal string")
		}
		numStr = str[0]
	case 2:
		if len(str[0]) == 0 || len(str[1]) == 0 {
			return f, ErrUnknownRequest("not a decimal string")
		}
		numStr = str[0] + str[1]
		len := int64(len(str[1]))
		denom = new(big.Int).Exp(big.NewInt(10), big.NewInt(len), nil).Int64()
	default:
		return f, ErrUnknownRequest("not a decimal string")
	}

	num, errConv := strconv.Atoi(numStr)
	if errConv != nil {
		return f, ErrUnknownRequest(errConv.Error())
	}

	if neg {
		num *= -1
	}

	return NewRat(int64(num), denom), nil
}

// NewRatFromBigInt constructs Rat from big.Int
func NewRatFromBigInt(num *big.Int, denom ...*big.Int) Rat {
	switch len(denom) {
	case 0:
		return Rat{*new(big.Rat).SetInt(num)}
	case 1:
		return Rat{*new(big.Rat).SetFrac(num, denom[0])}
	default:
		panic("improper use of NewRatFromBigInt, can only have one denominator")
	}
}

// NewRatFromInt constructs Rat from Int
func NewRatFromInt(num Int, denom ...Int) Rat {
	switch len(denom) {
	case 0:
		return Rat{*new(big.Rat).SetInt(num.BigInt())}
	case 1:
		return Rat{*new(big.Rat).SetFrac(num.BigInt(), denom[0].BigInt())}
	default:
		panic("improper use of NewRatFromBigInt, can only have one denominator")
	}
}

//nolint
func (r Rat) Num() int64        { return r.Rat.Num().Int64() }   // Num - return the numerator
func (r Rat) Denom() int64      { return r.Rat.Denom().Int64() } // Denom  - return the denominator
func (r Rat) IsZero() bool      { return r.Num() == 0 }          // IsZero - Is the Rat equal to zero
func (r Rat) Equal(r2 Rat) bool { return (&(r.Rat)).Cmp(&(r2.Rat)) == 0 }
func (r Rat) GT(r2 Rat) bool    { return (&(r.Rat)).Cmp(&(r2.Rat)) == 1 }              // greater than
func (r Rat) LT(r2 Rat) bool    { return (&(r.Rat)).Cmp(&(r2.Rat)) == -1 }             // less than
func (r Rat) Mul(r2 Rat) Rat    { return Rat{*new(big.Rat).Mul(&(r.Rat), &(r2.Rat))} } // Mul - multiplication
func (r Rat) Quo(r2 Rat) Rat    { return Rat{*new(big.Rat).Quo(&(r.Rat), &(r2.Rat))} } // Quo - quotient
func (r Rat) Add(r2 Rat) Rat    { return Rat{*new(big.Rat).Add(&(r.Rat), &(r2.Rat))} } // Add - addition
func (r Rat) Sub(r2 Rat) Rat    { return Rat{*new(big.Rat).Sub(&(r.Rat), &(r2.Rat))} } // Sub - subtraction
func (r Rat) String() string    { return fmt.Sprintf("%v/%v", r.Num(), r.Denom()) }

var (
	zero  = big.NewInt(0)
	one   = big.NewInt(1)
	two   = big.NewInt(2)
	five  = big.NewInt(5)
	nFive = big.NewInt(-5)
	ten   = big.NewInt(10)
)

// evaluate the rational using bankers rounding
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

// evaluate the rational using bankers rounding
func (r Rat) Evaluate() int64 {
	return r.EvaluateBig().Int64()
}

// EvaulateInt evaludates the rational using EvaluateBig
func (r Rat) EvaluateInt() Int {
	return NewIntFromBigInt(r.EvaluateBig())
}

// round Rat with the provided precisionFactor
func (r Rat) Round(precisionFactor int64) Rat {
	rTen := Rat{*new(big.Rat).Mul(&(r.Rat), big.NewRat(precisionFactor, 1))}
	return Rat{*big.NewRat(rTen.Evaluate(), precisionFactor)}
}

// TODO panic if negative or if totalDigits < len(initStr)???
// evaluate as an integer and return left padded string
func (r Rat) ToLeftPadded(totalDigits int8) string {
	intStr := r.EvaluateBig().String()
	fcode := `%0` + strconv.Itoa(int(totalDigits)) + `s`
	return fmt.Sprintf(fcode, intStr)
}

//___________________________________________________________________________________

//Wraps r.MarshalText().
func (r Rat) MarshalAmino() (string, error) {
	bz, err := (&(r.Rat)).MarshalText()
	return string(bz), err
}

// Requires a valid JSON string - strings quotes and calls UnmarshalText
func (r *Rat) UnmarshalAmino(text string) (err error) {
	tempRat := big.NewRat(0, 1)
	err = tempRat.UnmarshalText([]byte(text))
	if err != nil {
		return err
	}
	r.Rat = *tempRat
	return nil
}

//___________________________________________________________________________________
// helpers

// test if two rat arrays are the equal
func RatsEqual(r1s, r2s []Rat) bool {
	if len(r1s) != len(r2s) {
		return false
	}

	for i, r1 := range r1s {
		if !r1.Equal(r2s[i]) {
			return false
		}
	}
	return true
}

// intended to be used with require/assert:  require.True(RatEq(...))
func RatEq(t *testing.T, exp, got Rat) (*testing.T, bool, string, Rat, Rat) {
	return t, exp.Equal(got), "expected:\t%v\ngot:\t\t%v", exp, got
}
