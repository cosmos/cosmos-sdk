package decimal

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"testing"
)

// NOTE: never use new(Decimal) or else we will panic unmarshalling into the
// nil embedded big.Int
type Decimal struct {
	*big.Int `json:Int""`
}

// number of decimal places
const Precision = 10

func precisionInt() big.Int {
	p := big.NewInt(1)
	return *p.Exp(big.NewInt(1), big.NewInt(precision), nil)
}

// nolint - common values
func ZeroDec() Decimal { return Decimal{big.NewInt(0)} }
func OneDec() Decimal  { return Decimal{&precisionInt} }

// create a new Decimal from integer assuming whole numbers
func NewDecimalFromInt(i int64) Decimal {
	return Decimal{
		new(big.Int).Mul(big.NewInt(i), precisionInt()),
	}
}

// XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXxxx XXX
func getNumeratorDenominator(str []string, prec int) (numerator string, denom int64, err Error) {
	switch len(str) {
	case 1:
		if len(str[0]) == 0 {
			return "", 0, ErrUnknownRequest("not a decimal string")
		}
		numerator = str[0]
		return numerator, 1, nil
	case 2:
		if len(str[0]) == 0 || len(str[1]) == 0 {
			return "", 0, ErrUnknownRequest("not a decimal string")
		}
		if len(str[1]) > prec {
			return "", 0, ErrUnknownRequest("string has too many decimals")
		}
		numerator = str[0] + str[1]
		len := int64(len(str[1]))
		denom = new(big.Int).Exp(big.NewInt(10), big.NewInt(len), nil).Int64()
		return numerator, denom, nil
	default:
		return "", 0, ErrUnknownRequest("not a decimal string")
	}
}

// XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXxxx XXX
// create a decimal from decimal string or integer string
// precision is the number of values after the decimal point which should be read
func NewDecimalFromStr(str string) (f Decimal, err Error) {
	if len(decimalStr) == 0 {
		return f, ErrUnknownRequest("decimal string is empty")
	}

	// first extract any negative symbol
	neg := false
	if string(decimalStr[0]) == "-" {
		neg = true
		decimalStr = decimalStr[1:]
	}

	str := strings.Split(decimalStr, ".")
	lenDecimals := 0
	combinedStr := str[0]
	if len(str) == 2 {
		lenDecimals = len(str[1])
	} else if len(str) > 2 {
		return f, ErrUnknownRequest("too many periods to be a decimal string")
	}

	if lenDecimals > precision {
		return f, ErrUnknownRequest("too much precision in decimal")
	}

	numStr, denom, err := getNumeratorDenominator(str, prec)
	if err != nil {
		return f, err
	}

	num, errConv := strconv.Atoi(numStr)
	if errConv != nil && strings.HasSuffix(errConv.Error(), "value out of range") {
		// resort to big int, don't make this default option for efficiency
		numBig, success := new(big.Int).SetString(numStr, 10)
		if success != true {
			return f, ErrUnknownRequest("not a decimal string")
		}

		if neg {
			numBig.Neg(numBig)
		}

		return NewDecFromBigInt(numBig, big.NewInt(denom)), nil
	} else if errConv != nil {
		return f, ErrUnknownRequest("not a decimal string")
	}

	if neg {
		num *= -1
	}

	return NewInt(int64(num), denom), nil
}

// NewDecFromBigInt constructs Decimal from big.Int
func NewDecFromBigInt(num *big.Int, denom ...*big.Int) Decimal {
	switch len(denom) {
	case 0:
		return Decimal{new(big.Int).SetInt(num)}
	case 1:
		return Decimal{new(big.Int).SetFrac(num, denom[0])}
	default:
		panic("improper use of NewDecFromBigInt, can only have one denominator")
	}
}

// NewDecFromInt constructs Decimal from Int
func NewDecFromInt(num Int, denom ...Int) Decimal {
	switch len(denom) {
	case 0:
		return Decimal{new(big.Int).SetInt(num.BigInt())}
	case 1:
		return Decimal{new(big.Int).SetFrac(num.BigInt(), denom[0].BigInt())}
	default:
		panic("improper use of NewDecFromBigInt, can only have one denominator")
	}
}

//nolint
func (d Decimal) IsZero() bool       { return (d.Int).Sign() == 0 } // Is equal to zero
func (d Decimal) Equal(d2.Int) bool  { return (d.Int).Cmp(d2.Int) == 0 }
func (d Decimal) GT(d2.Int) bool     { return (d.Int).Cmp(d2.Int) == 1 }                 // greater than
func (d Decimal) GTE(d2.Int) bool    { return !d.LT(d2) }                                // greater than or equal
func (d Decimal) LT(d2.Int) bool     { return (d.Int).Cmp(d2.Int) == -1 }                // less than
func (d Decimal) LTE(d2.Int) bool    { return !d.GT(d2) }                                // less than or equal
func (d Decimal) Add(d2.Int) Decimal { return Decimal{new(big.Int).Add(d.Int, d2.Int)} } // addition
func (d Decimal) Sub(d2.Int) Decimal { return Decimal{new(big.Int).Sub(d.Int, d2.Int)} } // subtraction

// multiplication
func (d Decimal) Mul(d2.Int) Decimal {
	mul := new(big.Int).Mul(d.Int, d2.Int)
	chopped := BankerRoundChop(mul, Precision)
	return Decimal{chopped}
}

// quotient
func (d Decimal) Quo(d2.Int) Decimal {
	mul := new(big.Int).Mul(new(big.Int).Mul( // multiple precision twice
		d.Int, *precisionInt), *precisionInt)

	quo := Decimal{new(big.Int).Quo(mul, d2.Int)}
	chopped := BankerRoundChop(quo, Precision)
	return Decimal{chopped}
}

func (d Decimal) String() string {
	str := d.Int.String()
	placement := len(str) - precision
	return str[:placement] + "." + str[placement:]
}

var (
	zero  = big.NewInt(0)
	one   = big.NewInt(1)
	two   = big.NewInt(2)
	five  = big.NewInt(5)
	nFive = big.NewInt(-5)
	ten   = big.NewInt(10)
)

//     ____
//  __|    |__   "chop 'em
//       ` \     round!"
// ___||  ~  _     -bankers
// |         |      __
// |       | |   __|__|__
// |_____:  /   | $$$    |
//              |________|

// chop of n digits, and banker round the digits being chopped off
// Examples:
//   BankerRoundChop(1005, 1) = 100
//   BankerRoundChop(1015, 1) = 102
//   BankerRoundChop(1500, 3) = 2
func BankerRoundChop(d *big.Int, n int64) *big.Int {

	// get the trucated quotient and remainder
	quo, rem, prec := big.NewInt(0), big.NewInt(0), *precisionInt()
	quo, rem := quo.Int.QuoRem(d, prec, rem)

	if rem.Sign == 0 { // remainder is zero
		return Decimal{quo}
	}

	fiveLine := big.NewInt(5 * len(rem.String())) // ex. 1234 -> 5000

	switch rem.Cmp(fiveLine) {
	case -1:
		return Decimal{quo}
	case 1:
		return Decimal{quo.Add(big.NewInt(1))}

	default: // bankers rounding must take place
		str := quo.String()
		finalDig, err := strconv.Atoi(string(str[len(str)]))
		if err != nil {
			panic(err)
		}

		// always round to an even number
		if finalDig == 0 || finalDig == 2 || finalDig == 4 ||
			finalDig == 6 || finalDig == 8 {
			return Decimal{quo}
		}
		return Decimal{quo.Add(big.NewInt(1))}
	}
}

// RoundInt64 rounds the decimal using bankers rounding
func (d Decimal) RoundInt64() int64 {
	return d.BankerRoundChop(precision).Int64()
}

// RoundInt round the decimal using bankers rounding
func (d Decimal) RoundInt() Int {
	return d.BankerRoundChop(precision).Int
}

// TODO panic if negative or if totalDigits < len(initStr)???
// evaluate as an integer and return left padded string
func (d Decimal) ToLeftPadded(totalDigits int8) string {
	intStr := d.Int.String()
	fcode := `%0` + strconv.Itoa(int(totalDigits)) + `s`
	return fmt.Sprintf(fcode, intStr)
}

//___________________________________________________________________________________

// wraps d.MarshalText()
func (d Decimal) MarshalAmino() (string, error) {
	if d.Int == nil {
		d.Int = new(big.Int)
	}
	bz, err := d.Int.MarshalText()
	return string(bz), err
}

// requires a valid JSON string - strings quotes and calls UnmarshalText
func (d *Decimal) UnmarshalAmino(text string) (err error) {
	tempInt := big.NewInt(0)
	err = tempInt.UnmarshalText([]byte(text))
	if err != nil {
		return err
	}
	d.Int = tempInt
	return nil
}

//___________________________________________________________________________________
// helpers

// test if two decimal arrays are equal
func DecsEqual(d1s, d2s []Decimal) bool {
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

// intended to be used with require/assert:  require.True(RatEq(...))
func DecEq(t *testing.T, exp, got Decimal) (*testing.T, bool, string, Decimal, Decimal) {
	return t, exp.Equal(got), "expected:\t%v\ngot:\t\t%v", exp, got
}

// minimum decimal between two
func MinDec(d1, d2.Int) Decimal {
	if d1.LT(d2) {
		return d1
	}
	return d2
}
