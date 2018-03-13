package types

import (
	"errors"
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
type Rat struct {
	*big.Rat `json:"rat"`
}

// Rational - big Rat with additional functionality
type Rational interface {
	GetRat() *big.Rat
	Num() int64
	Denom() int64
	GT(Rational) bool
	LT(Rational) bool
	Equal(Rational) bool
	IsZero() bool
	Inv() Rational
	Mul(Rational) Rational
	Quo(Rational) Rational
	Add(Rational) Rational
	Sub(Rational) Rational
	Round(int64) Rational
	Evaluate() int64
}

var _ Rational = Rat{} // enforce at compile time

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
func NewRatFromDecimal(decimalStr string) (f Rat, err error) {

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
			return f, errors.New("not a decimal string")
		}
		numStr = str[0]
	case 2:
		if len(str[0]) == 0 || len(str[1]) == 0 {
			return f, errors.New("not a decimal string")
		}
		numStr = str[0] + str[1]
		len := int64(len(str[1]))
		denom = new(big.Int).Exp(big.NewInt(10), big.NewInt(len), nil).Int64()
	default:
		return f, errors.New("not a decimal string")
	}

	num, err := strconv.Atoi(numStr)
	if err != nil {
		return f, err
	}

	if neg {
		num *= -1
	}

	return Rat{big.NewRat(int64(num), denom)}, nil
}

//nolint
func (r Rat) GetRat() *big.Rat         { return r.Rat }                                     // GetRat - get big.Rat
func (r Rat) Num() int64               { return r.Rat.Num().Int64() }                       // Num - return the numerator
func (r Rat) Denom() int64             { return r.Rat.Denom().Int64() }                     // Denom  - return the denominator
func (r Rat) IsZero() bool             { return r.Num() == 0 }                              // IsZero - Is the Rat equal to zero
func (r Rat) Equal(r2 Rational) bool   { return r.Rat.Cmp(r2.GetRat()) == 0 }               // Equal - rationals are equal
func (r Rat) GT(r2 Rational) bool      { return r.Rat.Cmp(r2.GetRat()) == 1 }               // GT - greater than
func (r Rat) LT(r2 Rational) bool      { return r.Rat.Cmp(r2.GetRat()) == -1 }              // LT - less than
func (r Rat) Inv() Rational            { return Rat{new(big.Rat).Inv(r.Rat)} }              // Inv - inverse
func (r Rat) Mul(r2 Rational) Rational { return Rat{new(big.Rat).Mul(r.Rat, r2.GetRat())} } // Mul - multiplication
func (r Rat) Quo(r2 Rational) Rational { return Rat{new(big.Rat).Quo(r.Rat, r2.GetRat())} } // Quo - quotient
func (r Rat) Add(r2 Rational) Rational { return Rat{new(big.Rat).Add(r.Rat, r2.GetRat())} } // Add - addition
func (r Rat) Sub(r2 Rational) Rational { return Rat{new(big.Rat).Sub(r.Rat, r2.GetRat())} } // Sub - subtraction

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
func (r Rat) Round(precisionFactor int64) Rational {
	rTen := Rat{new(big.Rat).Mul(r.Rat, big.NewRat(precisionFactor, 1))}
	return Rat{big.NewRat(rTen.Evaluate(), precisionFactor)}
}

//___________________________________________________________________________________

//var ratCdc = RegisterWire(wire.NewCodec())
//// add rational codec elements to provided codec
//func RegisterWire(cdc *wire.Codec) *wire.Codec {
//cdc.RegisterInterface((*Rational)(nil), nil)
//cdc.RegisterConcrete(Rat{}, "rat", nil)
//return cdc
//}

//TODO there has got to be a better way using native MarshalText and UnmarshalText

// RatMarshal - Marshable Rat Struct
//type RatMarshal struct {
//Numerator   int64 `json:"numerator"`
//Denominator int64 `json:"denominator"`
//}

//// MarshalJSON - custom implementation of JSON Marshal
//func (r Rat) MarshalJSON() ([]byte, error) {
//return ratCdc.MarshalJSON(RatMarshal{r.Num(), r.Denom()})
//}

//// UnmarshalJSON - custom implementation of JSON Unmarshal
//func (r *Rat) UnmarshalJSON(data []byte) (err error) {
//defer func() {
//if rcv := recover(); rcv != nil {
//err = fmt.Errorf("Panic during UnmarshalJSON: %v", rcv)
//}
//}()

//ratMar := new(RatMarshal)
//if err := ratCdc.UnmarshalJSON(data, ratMar); err != nil {
//return err
//}
//r.Rat = big.NewRat(ratMar.Numerator, ratMar.Denominator)

//return nil
//}

//nolint
func (r Rat) MarshalJSON() ([]byte, error)           { return r.MarshalText() }
func (r *Rat) UnmarshalJSON(data []byte) (err error) { return r.UnmarshalText(data) }
