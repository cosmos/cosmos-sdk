package math_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"sigs.k8s.io/yaml"

	"cosmossdk.io/math"
)

type decimalTestSuite struct {
	suite.Suite
}

func TestDecimalTestSuite(t *testing.T) {
	suite.Run(t, new(decimalTestSuite))
}

func TestDecApproxEq(t *testing.T) {
	// d1 = 0.55, d2 = 0.6, tol = 0.1
	d1 := LegacyNewDecWithPrec(55, 2)
	d2 := LegacyNewDecWithPrec(6, 1)
	tol := LegacyNewDecWithPrec(1, 1)

	require.True(LegacyDecApproxEq(t, d1, d2, tol))

	// d1 = 0.55, d2 = 0.6, tol = 1E-5
	d1 = LegacyNewDecWithPrec(55, 2)
	d2 = LegacyNewDecWithPrec(6, 1)
	tol = LegacyNewDecWithPrec(1, 5)

	require.False(LegacyDecApproxEq(t, d1, d2, tol))

	// d1 = 0.6, d2 = 0.61, tol = 0.01
	d1 = LegacyNewDecWithPrec(6, 1)
	d2 = LegacyNewDecWithPrec(61, 2)
	tol = LegacyNewDecWithPrec(1, 2)

	require.True(LegacyDecApproxEq(t, d1, d2, tol))
}

// create a decimal from a decimal string (ex. "1234.5678")
func (s *decimalTestSuite) mustNewDecFromStr(str string) (d LegacyDec) {
	d, err := LegacyNewDecFromStr(str)
	s.Require().NoError(err)

	return d
}

func (s *decimalTestSuite) TestNewDecFromStr() {
	largeBigInt, ok := new(big.Int).SetString("3144605511029693144278234343371835", 10)
	s.Require().True(ok)

	largerBigInt, ok := new(big.Int).SetString("8888888888888888888888888888888888888888888888888888888888888888888844444440", 10)
	s.Require().True(ok)

	largestBigInt, ok := new(big.Int).SetString("33499189745056880149688856635597007162669032647290798121690100488888732861290034376435130433535", 10)
	s.Require().True(ok)

	tests := []struct {
		decimalStr string
		expErr     bool
		exp        LegacyDec
	}{
		{"", true, LegacyDec{}},
		{"0.-75", true, LegacyDec{}},
		{"0", false, LegacyNewDec(0)},
		{"1", false, LegacyNewDec(1)},
		{"1.1", false, LegacyNewDecWithPrec(11, 1)},
		{"0.75", false, LegacyNewDecWithPrec(75, 2)},
		{"0.8", false, LegacyNewDecWithPrec(8, 1)},
		{"0.11111", false, LegacyNewDecWithPrec(11111, 5)},
		{"314460551102969.3144278234343371835", true, LegacyNewDec(3141203149163817869)},
		{
			"314460551102969314427823434337.1835718092488231350",
			true, LegacyNewDecFromBigIntWithPrec(largeBigInt, 4),
		},
		{
			"314460551102969314427823434337.1835",
			false, LegacyNewDecFromBigIntWithPrec(largeBigInt, 4),
		},
		{".", true, LegacyDec{}},
		{".0", true, LegacyNewDec(0)},
		{"1.", true, LegacyNewDec(1)},
		{"foobar", true, LegacyDec{}},
		{"0.foobar", true, LegacyDec{}},
		{"0.foobar.", true, LegacyDec{}},
		{"8888888888888888888888888888888888888888888888888888888888888888888844444440", false, LegacyNewDecFromBigInt(largerBigInt)},
		{"33499189745056880149688856635597007162669032647290798121690100488888732861290.034376435130433535", false, LegacyNewDecFromBigIntWithPrec(largestBigInt, 18)},
		{"133499189745056880149688856635597007162669032647290798121690100488888732861291", true, LegacyDec{}},
	}

	for tcIndex, tc := range tests {
		res, err := LegacyNewDecFromStr(tc.decimalStr)
		if tc.expErr {
			s.Require().NotNil(err, "error expected, decimalStr %v, tc %v", tc.decimalStr, tcIndex)
		} else {
			s.Require().Nil(err, "unexpected error, decimalStr %v, tc %v", tc.decimalStr, tcIndex)
			s.Require().True(res.Equal(tc.exp), "equality was incorrect, res %v, exp %v, tc %v", res, tc.exp, tcIndex)
		}

		// negative tc
		res, err = LegacyNewDecFromStr("-" + tc.decimalStr)
		if tc.expErr {
			s.Require().NotNil(err, "error expected, decimalStr %v, tc %v", tc.decimalStr, tcIndex)
		} else {
			s.Require().Nil(err, "unexpected error, decimalStr %v, tc %v", tc.decimalStr, tcIndex)
			exp := tc.exp.Mul(LegacyNewDec(-1))
			s.Require().True(res.Equal(exp), "equality was incorrect, res %v, exp %v, tc %v", res, exp, tcIndex)
		}
	}
}

func (s *decimalTestSuite) TestDecString() {
	tests := []struct {
		d    LegacyDec
		want string
	}{
		{LegacyNewDec(0), "0.000000000000000000"},
		{LegacyNewDec(1), "1.000000000000000000"},
		{LegacyNewDec(10), "10.000000000000000000"},
		{LegacyNewDec(12340), "12340.000000000000000000"},
		{LegacyNewDecWithPrec(12340, 4), "1.234000000000000000"},
		{LegacyNewDecWithPrec(12340, 5), "0.123400000000000000"},
		{LegacyNewDecWithPrec(12340, 8), "0.000123400000000000"},
		{LegacyNewDecWithPrec(1009009009009009009, 17), "10.090090090090090090"},
	}
	for tcIndex, tc := range tests {
		s.Require().Equal(tc.want, tc.d.String(), "bad String(), index: %v", tcIndex)
	}
}

func (s *decimalTestSuite) TestDecFloat64() {
	tests := []struct {
		d    LegacyDec
		want float64
	}{
		{LegacyNewDec(0), 0.000000000000000000},
		{LegacyNewDec(1), 1.000000000000000000},
		{LegacyNewDec(10), 10.000000000000000000},
		{LegacyNewDec(12340), 12340.000000000000000000},
		{LegacyNewDecWithPrec(12340, 4), 1.234000000000000000},
		{LegacyNewDecWithPrec(12340, 5), 0.123400000000000000},
		{LegacyNewDecWithPrec(12340, 8), 0.000123400000000000},
		{LegacyNewDecWithPrec(1009009009009009009, 17), 10.090090090090090090},
	}
	for tcIndex, tc := range tests {
		value, err := tc.d.Float64()
		s.Require().Nil(err, "error getting Float64(), index: %v", tcIndex)
		s.Require().Equal(tc.want, value, "bad Float64(), index: %v", tcIndex)
		s.Require().Equal(tc.want, tc.d.MustFloat64(), "bad MustFloat64(), index: %v", tcIndex)
	}
}

func (s *decimalTestSuite) TestEqualities() {
	tests := []struct {
		d1, d2     LegacyDec
		gt, lt, eq bool
	}{
		{LegacyNewDec(0), LegacyNewDec(0), false, false, true},
		{LegacyNewDecWithPrec(0, 2), LegacyNewDecWithPrec(0, 4), false, false, true},
		{LegacyNewDecWithPrec(100, 0), LegacyNewDecWithPrec(100, 0), false, false, true},
		{LegacyNewDecWithPrec(-100, 0), LegacyNewDecWithPrec(-100, 0), false, false, true},
		{LegacyNewDecWithPrec(-1, 1), LegacyNewDecWithPrec(-1, 1), false, false, true},
		{LegacyNewDecWithPrec(3333, 3), LegacyNewDecWithPrec(3333, 3), false, false, true},

		{LegacyNewDecWithPrec(0, 0), LegacyNewDecWithPrec(3333, 3), false, true, false},
		{LegacyNewDecWithPrec(0, 0), LegacyNewDecWithPrec(100, 0), false, true, false},
		{LegacyNewDecWithPrec(-1, 0), LegacyNewDecWithPrec(3333, 3), false, true, false},
		{LegacyNewDecWithPrec(-1, 0), LegacyNewDecWithPrec(100, 0), false, true, false},
		{LegacyNewDecWithPrec(1111, 3), LegacyNewDecWithPrec(100, 0), false, true, false},
		{LegacyNewDecWithPrec(1111, 3), LegacyNewDecWithPrec(3333, 3), false, true, false},
		{LegacyNewDecWithPrec(-3333, 3), LegacyNewDecWithPrec(-1111, 3), false, true, false},

		{LegacyNewDecWithPrec(3333, 3), LegacyNewDecWithPrec(0, 0), true, false, false},
		{LegacyNewDecWithPrec(100, 0), LegacyNewDecWithPrec(0, 0), true, false, false},
		{LegacyNewDecWithPrec(3333, 3), LegacyNewDecWithPrec(-1, 0), true, false, false},
		{LegacyNewDecWithPrec(100, 0), LegacyNewDecWithPrec(-1, 0), true, false, false},
		{LegacyNewDecWithPrec(100, 0), LegacyNewDecWithPrec(1111, 3), true, false, false},
		{LegacyNewDecWithPrec(3333, 3), LegacyNewDecWithPrec(1111, 3), true, false, false},
		{LegacyNewDecWithPrec(-1111, 3), LegacyNewDecWithPrec(-3333, 3), true, false, false},
	}

	for tcIndex, tc := range tests {
		s.Require().Equal(tc.gt, tc.d1.GT(tc.d2), "GT result is incorrect, tc %d", tcIndex)
		s.Require().Equal(tc.lt, tc.d1.LT(tc.d2), "LT result is incorrect, tc %d", tcIndex)
		s.Require().Equal(tc.eq, tc.d1.Equal(tc.d2), "equality result is incorrect, tc %d", tcIndex)
	}
}

func (s *decimalTestSuite) TestDecsEqual() {
	tests := []struct {
		d1s, d2s []LegacyDec
		eq       bool
	}{
		{[]LegacyDec{LegacyNewDec(0)}, []LegacyDec{LegacyNewDec(0)}, true},
		{[]LegacyDec{LegacyNewDec(0)}, []LegacyDec{LegacyNewDec(1)}, false},
		{[]LegacyDec{LegacyNewDec(0)}, []LegacyDec{}, false},
		{[]LegacyDec{LegacyNewDec(0), LegacyNewDec(1)}, []LegacyDec{LegacyNewDec(0), LegacyNewDec(1)}, true},
		{[]LegacyDec{LegacyNewDec(1), LegacyNewDec(0)}, []LegacyDec{LegacyNewDec(1), LegacyNewDec(0)}, true},
		{[]LegacyDec{LegacyNewDec(1), LegacyNewDec(0)}, []LegacyDec{LegacyNewDec(0), LegacyNewDec(1)}, false},
		{[]LegacyDec{LegacyNewDec(1), LegacyNewDec(0)}, []LegacyDec{LegacyNewDec(1)}, false},
		{[]LegacyDec{LegacyNewDec(1), LegacyNewDec(2)}, []LegacyDec{LegacyNewDec(2), LegacyNewDec(4)}, false},
		{[]LegacyDec{LegacyNewDec(3), LegacyNewDec(18)}, []LegacyDec{LegacyNewDec(1), LegacyNewDec(6)}, false},
	}

	for tcIndex, tc := range tests {
		s.Require().Equal(tc.eq, LegacyDecsEqual(tc.d1s, tc.d2s), "equality of decional arrays is incorrect, tc %d", tcIndex)
		s.Require().Equal(tc.eq, LegacyDecsEqual(tc.d2s, tc.d1s), "equality of decional arrays is incorrect (converse), tc %d", tcIndex)
	}
}

func (s *decimalTestSuite) TestArithmetic() {
	tests := []struct {
		d1, d2                                LegacyDec
		expMul, expMulTruncate, expMulRoundUp LegacyDec
		expQuo, expQuoRoundUp, expQuoTruncate LegacyDec
		expAdd, expSub                        LegacyDec
	}{
		//  d1         d2         MUL    MulTruncate   MulRoundUp    QUO    QUORoundUp QUOTrunctate  ADD         SUB
		{LegacyNewDec(0), LegacyNewDec(0), LegacyNewDec(0), LegacyNewDec(0), LegacyNewDec(0), LegacyNewDec(0), LegacyNewDec(0), LegacyNewDec(0), LegacyNewDec(0), LegacyNewDec(0)},
		{LegacyNewDec(1), LegacyNewDec(0), LegacyNewDec(0), LegacyNewDec(0), LegacyNewDec(0), LegacyNewDec(0), LegacyNewDec(0), LegacyNewDec(0), LegacyNewDec(1), LegacyNewDec(1)},
		{LegacyNewDec(0), LegacyNewDec(1), LegacyNewDec(0), LegacyNewDec(0), LegacyNewDec(0), LegacyNewDec(0), LegacyNewDec(0), LegacyNewDec(0), LegacyNewDec(1), LegacyNewDec(-1)},
		{LegacyNewDec(0), LegacyNewDec(-1), LegacyNewDec(0), LegacyNewDec(0), LegacyNewDec(0), LegacyNewDec(0), LegacyNewDec(0), LegacyNewDec(0), LegacyNewDec(-1), LegacyNewDec(1)},
		{LegacyNewDec(-1), LegacyNewDec(0), LegacyNewDec(0), LegacyNewDec(0), LegacyNewDec(0), LegacyNewDec(0), LegacyNewDec(0), LegacyNewDec(0), LegacyNewDec(-1), LegacyNewDec(-1)},

		{LegacyNewDec(1), LegacyNewDec(1), LegacyNewDec(1), LegacyNewDec(1), LegacyNewDec(1), LegacyNewDec(1), LegacyNewDec(1), LegacyNewDec(1), LegacyNewDec(2), LegacyNewDec(0)},
		{LegacyNewDec(-1), LegacyNewDec(-1), LegacyNewDec(1), LegacyNewDec(1), LegacyNewDec(1), LegacyNewDec(1), LegacyNewDec(1), LegacyNewDec(1), LegacyNewDec(-2), LegacyNewDec(0)},
		{LegacyNewDec(1), LegacyNewDec(-1), LegacyNewDec(-1), LegacyNewDec(-1), LegacyNewDec(-1), LegacyNewDec(-1), LegacyNewDec(-1), LegacyNewDec(-1), LegacyNewDec(0), LegacyNewDec(2)},
		{LegacyNewDec(-1), LegacyNewDec(1), LegacyNewDec(-1), LegacyNewDec(-1), LegacyNewDec(-1), LegacyNewDec(-1), LegacyNewDec(-1), LegacyNewDec(-1), LegacyNewDec(0), LegacyNewDec(-2)},

		{
			LegacyNewDec(3), LegacyNewDec(7), LegacyNewDec(21), LegacyNewDec(21), LegacyNewDec(21),
			LegacyNewDecWithPrec(428571428571428571, 18), LegacyNewDecWithPrec(428571428571428572, 18), LegacyNewDecWithPrec(428571428571428571, 18),
			LegacyNewDec(10), LegacyNewDec(-4),
		},
		{
			LegacyNewDec(2), LegacyNewDec(4), LegacyNewDec(8), LegacyNewDec(8), LegacyNewDec(8), LegacyNewDecWithPrec(5, 1), LegacyNewDecWithPrec(5, 1), LegacyNewDecWithPrec(5, 1),
			LegacyNewDec(6), LegacyNewDec(-2),
		},

		{LegacyNewDec(100), LegacyNewDec(100), LegacyNewDec(10000), LegacyNewDec(10000), LegacyNewDec(10000), LegacyNewDec(1), LegacyNewDec(1), LegacyNewDec(1), LegacyNewDec(200), LegacyNewDec(0)},

		{
			LegacyNewDecWithPrec(15, 1), LegacyNewDecWithPrec(15, 1), LegacyNewDecWithPrec(225, 2), LegacyNewDecWithPrec(225, 2), LegacyNewDecWithPrec(225, 2),
			LegacyNewDec(1), LegacyNewDec(1), LegacyNewDec(1), LegacyNewDec(3), LegacyNewDec(0),
		},
		{
			LegacyNewDecWithPrec(3333, 4), LegacyNewDecWithPrec(333, 4), LegacyNewDecWithPrec(1109889, 8), LegacyNewDecWithPrec(1109889, 8), LegacyNewDecWithPrec(1109889, 8),
			LegacyMustNewDecFromStr("10.009009009009009009"), LegacyMustNewDecFromStr("10.009009009009009010"), LegacyMustNewDecFromStr("10.009009009009009009"),
			LegacyNewDecWithPrec(3666, 4), LegacyNewDecWithPrec(3, 1),
		},
	}

	for tcIndex, tc := range tests {
		tc := tc
		resAdd := tc.d1.Add(tc.d2)
		resSub := tc.d1.Sub(tc.d2)
		resMul := tc.d1.Mul(tc.d2)
		resMulTruncate := tc.d1.MulTruncate(tc.d2)
		resMulRoundUp := tc.d1.MulRoundUp(tc.d2)
		s.Require().True(tc.expAdd.Equal(resAdd), "exp %v, res %v, tc %d", tc.expAdd, resAdd, tcIndex)
		s.Require().True(tc.expSub.Equal(resSub), "exp %v, res %v, tc %d", tc.expSub, resSub, tcIndex)
		s.Require().True(tc.expMul.Equal(resMul), "exp %v, res %v, tc %d", tc.expMul, resMul, tcIndex)
		s.Require().True(tc.expMulTruncate.Equal(resMulTruncate), "exp %v, res %v, tc %d", tc.expMulTruncate, resMulTruncate, tcIndex)
		s.Require().True(tc.expMulRoundUp.Equal(resMulRoundUp), "exp %v, res %v, tc %d", tc.expMulRoundUp, resMulRoundUp, tcIndex)

		if tc.d2.IsZero() { // panic for divide by zero
			s.Require().Panics(func() { tc.d1.Quo(tc.d2) })
		} else {
			resQuo := tc.d1.Quo(tc.d2)
			s.Require().True(tc.expQuo.Equal(resQuo), "exp %v, res %v, tc %d", tc.expQuo.String(), resQuo.String(), tcIndex)

			resQuoRoundUp := tc.d1.QuoRoundUp(tc.d2)
			s.Require().True(tc.expQuoRoundUp.Equal(resQuoRoundUp), "exp %v, res %v, tc %d",
				tc.expQuoRoundUp.String(), resQuoRoundUp.String(), tcIndex)

			resQuoTruncate := tc.d1.QuoTruncate(tc.d2)
			s.Require().True(tc.expQuoTruncate.Equal(resQuoTruncate), "exp %v, res %v, tc %d",
				tc.expQuoTruncate.String(), resQuoTruncate.String(), tcIndex)
		}
	}
}

func (s *decimalTestSuite) TestMulRoundUp_RoundingAtPrecisionEnd() {
	var (
		a                = LegacyMustNewDecFromStr("0.000000000000000009")
		b                = LegacyMustNewDecFromStr("0.000000000000000009")
		expectedRoundUp  = LegacyMustNewDecFromStr("0.000000000000000001")
		expectedTruncate = LegacyMustNewDecFromStr("0.000000000000000000")
	)

	actualRoundUp := a.MulRoundUp(b)
	s.Require().Equal(expectedRoundUp.String(), actualRoundUp.String(), "exp %v, res %v", expectedRoundUp, actualRoundUp)

	actualTruncate := a.MulTruncate(b)
	s.Require().Equal(expectedTruncate.String(), actualTruncate.String(), "exp %v, res %v", expectedRoundUp, actualTruncate)
}

func (s *decimalTestSuite) TestBankerRoundChop() {
	tests := []struct {
		d1  LegacyDec
		exp int64
	}{
		{s.mustNewDecFromStr("0.25"), 0},
		{s.mustNewDecFromStr("0"), 0},
		{s.mustNewDecFromStr("1"), 1},
		{s.mustNewDecFromStr("0.75"), 1},
		{s.mustNewDecFromStr("0.5"), 0},
		{s.mustNewDecFromStr("7.5"), 8},
		{s.mustNewDecFromStr("1.5"), 2},
		{s.mustNewDecFromStr("2.5"), 2},
		{s.mustNewDecFromStr("0.545"), 1}, // 0.545-> 1 even though 5 is first decimal and 1 not even
		{s.mustNewDecFromStr("1.545"), 2},
	}

	for tcIndex, tc := range tests {
		resNeg := tc.d1.Neg().RoundInt64()
		s.Require().Equal(-1*tc.exp, resNeg, "negative tc %d", tcIndex)

		resPos := tc.d1.RoundInt64()
		s.Require().Equal(tc.exp, resPos, "positive tc %d", tcIndex)
	}
}

func (s *decimalTestSuite) TestTruncate() {
	tests := []struct {
		d1  LegacyDec
		exp int64
	}{
		{s.mustNewDecFromStr("0"), 0},
		{s.mustNewDecFromStr("0.25"), 0},
		{s.mustNewDecFromStr("0.75"), 0},
		{s.mustNewDecFromStr("1"), 1},
		{s.mustNewDecFromStr("1.5"), 1},
		{s.mustNewDecFromStr("7.5"), 7},
		{s.mustNewDecFromStr("7.6"), 7},
		{s.mustNewDecFromStr("7.4"), 7},
		{s.mustNewDecFromStr("100.1"), 100},
		{s.mustNewDecFromStr("1000.1"), 1000},
	}

	for tcIndex, tc := range tests {
		resNeg := tc.d1.Neg().TruncateInt64()
		s.Require().Equal(-1*tc.exp, resNeg, "negative tc %d", tcIndex)

		resPos := tc.d1.TruncateInt64()
		s.Require().Equal(tc.exp, resPos, "positive tc %d", tcIndex)
	}
}

func (s *decimalTestSuite) TestStringOverflow() {
	// two random 64 bit primes
	dec1, err := LegacyNewDecFromStr("51643150036226787134389711697696177267")
	s.Require().NoError(err)
	dec2, err := LegacyNewDecFromStr("-31798496660535729618459429845579852627")
	s.Require().NoError(err)
	dec3 := dec1.Add(dec2)
	s.Require().Equal(
		"19844653375691057515930281852116324640.000000000000000000",
		dec3.String(),
	)
}

func (s *decimalTestSuite) TestDecMulInt() {
	tests := []struct {
		sdkDec LegacyDec
		sdkInt math.Int
		want   LegacyDec
	}{
		{LegacyNewDec(10), math.NewInt(2), LegacyNewDec(20)},
		{LegacyNewDec(1000000), math.NewInt(100), LegacyNewDec(100000000)},
		{LegacyNewDecWithPrec(1, 1), math.NewInt(10), LegacyNewDec(1)},
		{LegacyNewDecWithPrec(1, 5), math.NewInt(20), LegacyNewDecWithPrec(2, 4)},
	}
	for i, tc := range tests {
		got := tc.sdkDec.MulInt(tc.sdkInt)
		s.Require().Equal(tc.want, got, "Incorrect result on test case %d", i)
	}
}

func (s *decimalTestSuite) TestDecCeil() {
	testCases := []struct {
		input    LegacyDec
		expected LegacyDec
	}{
		{LegacyNewDecWithPrec(1000000000000000, LegacyPrecision), LegacyNewDec(1)},      // 0.001 => 1.0
		{LegacyNewDecWithPrec(-1000000000000000, LegacyPrecision), LegacyZeroDec()},     // -0.001 => 0.0
		{LegacyZeroDec(), LegacyZeroDec()},                                              // 0.0 => 0.0
		{LegacyNewDecWithPrec(900000000000000000, LegacyPrecision), LegacyNewDec(1)},    // 0.9 => 1.0
		{LegacyNewDecWithPrec(4001000000000000000, LegacyPrecision), LegacyNewDec(5)},   // 4.001 => 5.0
		{LegacyNewDecWithPrec(-4001000000000000000, LegacyPrecision), LegacyNewDec(-4)}, // -4.001 => -4.0
		{LegacyNewDecWithPrec(4700000000000000000, LegacyPrecision), LegacyNewDec(5)},   // 4.7 => 5.0
		{LegacyNewDecWithPrec(-4700000000000000000, LegacyPrecision), LegacyNewDec(-4)}, // -4.7 => -4.0
	}

	for i, tc := range testCases {
		res := tc.input.Ceil()
		s.Require().Equal(tc.expected, res, "unexpected result for test case %d, input: %v", i, tc.input)
	}
}

func (s *decimalTestSuite) TestCeilOverflow() {
	d, err := LegacyNewDecFromStr("66749594872528440074844428317798503581334516323645399060845050244444366430645.000000000000000001")
	s.Require().NoError(err)
	s.Require().True(d.BigInt().BitLen() <= 315, "d is too large")
	// this call panics because the value is too large
	s.Require().Panics(func() { d.Ceil() }, "Ceil should panic on overflow")
}

func (s *decimalTestSuite) TestPower() {
	testCases := []struct {
		input    LegacyDec
		power    uint64
		expected LegacyDec
	}{
		{LegacyNewDec(100), 0, LegacyOneDec()},                                                  // 10 ^ (0) => 1.0
		{LegacyOneDec(), 10, LegacyOneDec()},                                                    // 1.0 ^ (10) => 1.0
		{LegacyNewDecWithPrec(5, 1), 2, LegacyNewDecWithPrec(25, 2)},                            // 0.5 ^ 2 => 0.25
		{LegacyNewDecWithPrec(2, 1), 2, LegacyNewDecWithPrec(4, 2)},                             // 0.2 ^ 2 => 0.04
		{LegacyNewDecFromInt(math.NewInt(3)), 3, LegacyNewDecFromInt(math.NewInt(27))},          // 3 ^ 3 => 27
		{LegacyNewDecFromInt(math.NewInt(-3)), 4, LegacyNewDecFromInt(math.NewInt(81))},         // -3 ^ 4 = 81
		{LegacyNewDecWithPrec(1414213562373095049, 18), 2, LegacyNewDecFromInt(math.NewInt(2))}, // 1.414213562373095049 ^ 2 = 2
	}

	for i, tc := range testCases {
		res := tc.input.Power(tc.power)
		s.Require().True(tc.expected.Sub(res).Abs().LTE(LegacySmallestDec()), "unexpected result for test case %d, normal power, input: %v", i, tc.input)

		mutableInput := tc.input
		mutableInput.PowerMut(tc.power)
		s.Require().True(tc.expected.Sub(mutableInput).Abs().LTE(LegacySmallestDec()),
			"unexpected result for test case %d, input %v", i, tc.input)
		s.Require().True(res.Equal(tc.input), "unexpected result for test case %d, mutable power, input: %v", i, tc.input)
	}
}

func (s *decimalTestSuite) TestApproxRoot() {
	testCases := []struct {
		input    LegacyDec
		root     uint64
		expected LegacyDec
	}{
		{LegacyOneDec(), 10, LegacyOneDec()},                                                       // 1.0 ^ (0.1) => 1.0
		{LegacyNewDecWithPrec(25, 2), 2, LegacyNewDecWithPrec(5, 1)},                               // 0.25 ^ (0.5) => 0.5
		{LegacyNewDecWithPrec(4, 2), 2, LegacyNewDecWithPrec(2, 1)},                                // 0.04 ^ (0.5) => 0.2
		{LegacyNewDecFromInt(math.NewInt(27)), 3, LegacyNewDecFromInt(math.NewInt(3))},             // 27 ^ (1/3) => 3
		{LegacyNewDecFromInt(math.NewInt(-81)), 4, LegacyNewDecFromInt(math.NewInt(-3))},           // -81 ^ (0.25) => -3
		{LegacyNewDecFromInt(math.NewInt(2)), 2, LegacyNewDecWithPrec(1414213562373095049, 18)},    // 2 ^ (0.5) => 1.414213562373095049
		{LegacyNewDecWithPrec(1005, 3), 31536000, LegacyMustNewDecFromStr("1.000000000158153904")}, // 1.005 ^ (1/31536000) ≈ 1.00000000016
		{LegacySmallestDec(), 2, LegacyNewDecWithPrec(1, 9)},                                       // 1e-18 ^ (0.5) => 1e-9
		{LegacySmallestDec(), 3, LegacyMustNewDecFromStr("0.000000999999999997")},                  // 1e-18 ^ (1/3) => 1e-6
		{LegacyNewDecWithPrec(1, 8), 3, LegacyMustNewDecFromStr("0.002154434690031900")},           // 1e-8 ^ (1/3) ≈ 0.00215443469
		{LegacyMustNewDecFromStr("9000002314687921634000000000000000000021394871242000000000000000"), 2, LegacyMustNewDecFromStr("94868342004527103646332858502867.899477053226766107")},
	}

	// In the case of 1e-8 ^ (1/3), the result repeats every 5 iterations starting from iteration 24
	// (i.e. 24, 29, 34, ... give the same result) and never converges enough. The maximum number of
	// iterations (300) causes the result at iteration 300 to be returned, regardless of convergence.

	for i, tc := range testCases {
		res, err := tc.input.ApproxRoot(tc.root)
		s.Require().NoError(err)
		s.Require().True(tc.expected.Sub(res).Abs().LTE(LegacySmallestDec()), "unexpected result for test case %d, input: %v", i, tc.input)
	}
}

func (s *decimalTestSuite) TestApproxSqrt() {
	testCases := []struct {
		input    LegacyDec
		expected LegacyDec
	}{
		{LegacyOneDec(), LegacyOneDec()},                                 // 1.0 => 1.0
		{LegacyNewDecWithPrec(25, 2), LegacyNewDecWithPrec(5, 1)},        // 0.25 => 0.5
		{LegacyNewDecWithPrec(4, 2), LegacyNewDecWithPrec(2, 1)},         // 0.09 => 0.3
		{LegacyNewDec(9), LegacyNewDecFromInt(math.NewInt(3))},           // 9 => 3
		{LegacyNewDec(-9), LegacyNewDecFromInt(math.NewInt(-3))},         // -9 => -3
		{LegacyNewDec(2), LegacyNewDecWithPrec(1414213562373095049, 18)}, // 2 => 1.414213562373095049
		{ // 2^127 - 1 => 13043817825332782212.3495718062525083688 which rounds to 13043817825332782212.3495718062525083689
			LegacyNewDec(2).Power(127).Sub(LegacyOneDec()),
			LegacyMustNewDecFromStr("13043817825332782212.349571806252508369"),
		},
		{LegacyMustNewDecFromStr("1.000000011823380862"), LegacyMustNewDecFromStr("1.000000005911690414")},
	}

	for i, tc := range testCases {
		res, err := tc.input.ApproxSqrt()
		s.Require().NoError(err)
		s.Require().Equal(tc.expected, res, "unexpected result for test case %d, input: %v", i, tc.input)
	}
}

func (s *decimalTestSuite) TestDecSortableBytes() {
	tests := []struct {
		d    LegacyDec
		want []byte
	}{
		{LegacyNewDec(0), []byte("000000000000000000.000000000000000000")},
		{LegacyNewDec(1), []byte("000000000000000001.000000000000000000")},
		{LegacyNewDec(10), []byte("000000000000000010.000000000000000000")},
		{LegacyNewDec(12340), []byte("000000000000012340.000000000000000000")},
		{LegacyNewDecWithPrec(12340, 4), []byte("000000000000000001.234000000000000000")},
		{LegacyNewDecWithPrec(12340, 5), []byte("000000000000000000.123400000000000000")},
		{LegacyNewDecWithPrec(12340, 8), []byte("000000000000000000.000123400000000000")},
		{LegacyNewDecWithPrec(1009009009009009009, 17), []byte("000000000000000010.090090090090090090")},
		{LegacyNewDecWithPrec(-1009009009009009009, 17), []byte("-000000000000000010.090090090090090090")},
		{LegacyNewDec(1000000000000000000), []byte("max")},
		{LegacyNewDec(-1000000000000000000), []byte("--")},
	}
	for tcIndex, tc := range tests {
		s.Require().Equal(tc.want, LegacySortableDecBytes(tc.d), "bad String(), index: %v", tcIndex)
	}

	s.Require().Panics(func() { LegacySortableDecBytes(LegacyNewDec(1000000000000000001)) })
	s.Require().Panics(func() { LegacySortableDecBytes(LegacyNewDec(-1000000000000000001)) })
}

func (s *decimalTestSuite) TestDecEncoding() {
	largestBigInt, ok := new(big.Int).SetString("33499189745056880149688856635597007162669032647290798121690100488888732861290034376435130433535", 10)
	s.Require().True(ok)

	smallestBigInt, ok := new(big.Int).SetString("-33499189745056880149688856635597007162669032647290798121690100488888732861290034376435130433535", 10)
	s.Require().True(ok)

	const maxDecBitLen = 315
	maxInt, ok := new(big.Int).SetString(strings.Repeat("1", maxDecBitLen), 2)
	s.Require().True(ok)

	testCases := []struct {
		input   LegacyDec
		rawBz   string
		jsonStr string
		yamlStr string
	}{
		{
			LegacyNewDec(0), "30",
			"\"0.000000000000000000\"",
			"\"0.000000000000000000\"\n",
		},
		{
			LegacyNewDecWithPrec(4, 2),
			"3430303030303030303030303030303030",
			"\"0.040000000000000000\"",
			"\"0.040000000000000000\"\n",
		},
		{
			LegacyNewDecWithPrec(-4, 2),
			"2D3430303030303030303030303030303030",
			"\"-0.040000000000000000\"",
			"\"-0.040000000000000000\"\n",
		},
		{
			LegacyNewDecWithPrec(1414213562373095049, 18),
			"31343134323133353632333733303935303439",
			"\"1.414213562373095049\"",
			"\"1.414213562373095049\"\n",
		},
		{
			LegacyNewDecWithPrec(-1414213562373095049, 18),
			"2D31343134323133353632333733303935303439",
			"\"-1.414213562373095049\"",
			"\"-1.414213562373095049\"\n",
		},
		{
			LegacyNewDecFromBigIntWithPrec(largestBigInt, 18),
			"3333343939313839373435303536383830313439363838383536363335353937303037313632363639303332363437323930373938313231363930313030343838383838373332383631323930303334333736343335313330343333353335",
			"\"33499189745056880149688856635597007162669032647290798121690100488888732861290.034376435130433535\"",
			"\"33499189745056880149688856635597007162669032647290798121690100488888732861290.034376435130433535\"\n",
		},
		{
			LegacyNewDecFromBigIntWithPrec(smallestBigInt, 18),
			"2D3333343939313839373435303536383830313439363838383536363335353937303037313632363639303332363437323930373938313231363930313030343838383838373332383631323930303334333736343335313330343333353335",
			"\"-33499189745056880149688856635597007162669032647290798121690100488888732861290.034376435130433535\"",
			"\"-33499189745056880149688856635597007162669032647290798121690100488888732861290.034376435130433535\"\n",
		},
		{
			LegacyNewDecFromBigIntWithPrec(maxInt, 18),
			"3636373439353934383732353238343430303734383434343238333137373938353033353831333334353136333233363435333939303630383435303530323434343434333636343330363435303137313838323137353635323136373637",
			"\"66749594872528440074844428317798503581334516323645399060845050244444366430645.017188217565216767\"",
			"\"66749594872528440074844428317798503581334516323645399060845050244444366430645.017188217565216767\"\n",
		},
	}

	for _, tc := range testCases {
		bz, err := tc.input.Marshal()
		s.Require().NoError(err)
		s.Require().Equal(tc.rawBz, fmt.Sprintf("%X", bz))

		var other LegacyDec
		s.Require().NoError((&other).Unmarshal(bz))
		s.Require().True(tc.input.Equal(other))

		bz, err = json.Marshal(tc.input)
		s.Require().NoError(err)
		s.Require().Equal(tc.jsonStr, string(bz))
		s.Require().NoError(json.Unmarshal(bz, &other))
		s.Require().True(tc.input.Equal(other))

		bz, err = yaml.Marshal(tc.input)
		s.Require().NoError(err)
		s.Require().Equal(tc.yamlStr, string(bz))
	}
}

// Showcase that different orders of operations causes different results.
func (s *decimalTestSuite) TestOperationOrders() {
	n1 := LegacyNewDec(10)
	n2 := LegacyNewDec(1000000010)
	s.Require().Equal(n1.Mul(n2).Quo(n2), LegacyNewDec(10))
	s.Require().NotEqual(n1.Mul(n2).Quo(n2), n1.Quo(n2).Mul(n2))
}

func BenchmarkMarshalTo(b *testing.B) {
	b.ReportAllocs()
	bis := []struct {
		in   LegacyDec
		want []byte
	}{
		{
			LegacyNewDec(1e8), []byte{
				0x31, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
				0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
				0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
			},
		},
		{LegacyNewDec(0), []byte{0x30}},
	}
	data := make([]byte, 100)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, bi := range bis {
			if n, err := bi.in.MarshalTo(data); err != nil {
				b.Fatal(err)
			} else if !bytes.Equal(data[:n], bi.want) {
				b.Fatalf("Mismatch\nGot:  % x\nWant: % x\n", data[:n], bi.want)
			}
		}
	}
}

var sink interface{}

func BenchmarkLegacyQuoMut(b *testing.B) {
	b1 := LegacyNewDec(17e2 + 8371)
	b2 := LegacyNewDec(4371)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink = b1.QuoMut(b2)
	}

	if sink == nil {
		b.Fatal("Benchmark did not run")
	}
	sink = (interface{})(nil)
}

func BenchmarkLegacyQuoTruncateMut(b *testing.B) {
	b1 := LegacyNewDec(17e2 + 8371)
	b2 := LegacyNewDec(4371)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink = b1.QuoTruncateMut(b2)
	}

	if sink == nil {
		b.Fatal("Benchmark did not run")
	}
	sink = (interface{})(nil)
}

func BenchmarkLegacySqrtOnMersennePrime(b *testing.B) {
	b1 := LegacyNewDec(2).Power(127).Sub(LegacyOneDec())
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink, _ = b1.ApproxSqrt()
	}

	if sink == nil {
		b.Fatal("Benchmark did not run")
	}
	sink = (interface{})(nil)
}

func BenchmarkLegacyQuoRoundupMut(b *testing.B) {
	b1 := LegacyNewDec(17e2 + 8371)
	b2 := LegacyNewDec(4371)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink = b1.QuoRoundupMut(b2)
	}

	if sink == nil {
		b.Fatal("Benchmark did not run")
	}
	sink = (interface{})(nil)
}

func TestFormatDec(t *testing.T) {
	type decimalTest []string
	var testcases []decimalTest
	raw, err := os.ReadFile("./testdata/decimals.json")
	require.NoError(t, err)
	err = json.Unmarshal(raw, &testcases)
	require.NoError(t, err)

	for _, tc := range testcases {
		tc := tc
		t.Run(tc[0], func(t *testing.T) {
			out, err := FormatDec(tc[0])
			require.NoError(t, err)
			require.Equal(t, tc[1], out)
		})
	}
}

func TestFormatDecNonDigits(t *testing.T) {
	badCases := []string{
		"10.a",
		"1a.10",
		"p1a10.",
		"0.10p",
		"--10",
		"12.😎😎",
		"11111111111133333333333333333333333333333a",
		"11111111111133333333333333333333333333333 192892",
	}

	for _, value := range badCases {
		value := value
		t.Run(value, func(t *testing.T) {
			s, err := FormatDec(value)
			if err == nil {
				t.Fatal("Expected an error")
			}
			if g, w := err.Error(), "non-digits"; !strings.Contains(g, w) {
				t.Errorf("Error mismatch\nGot:  %q\nWant substring: %q", g, w)
			}
			if s != "" {
				t.Fatalf("Got a non-empty string: %q", s)
			}
		})
	}
}

func TestNegativePrecisionPanic(t *testing.T) {
	require.Panics(t, func() {
		LegacyNewDecWithPrec(10, -1)
	})
}

func (s *decimalTestSuite) TestConvertToBigIntMutativeForLegacyDec() {
	r := big.NewInt(30)
	i := LegacyNewDecFromBigInt(r)

	// Compare value of BigInt & BigIntMut
	s.Require().Equal(i.BigInt(), i.BigIntMut())

	// Modify BigIntMut() pointer and ensure i.BigIntMut() & i.BigInt() change
	p1 := i.BigIntMut()
	p1.SetInt64(40)
	s.Require().Equal(big.NewInt(40), i.BigIntMut())
	s.Require().Equal(big.NewInt(40), i.BigInt())

	// Modify big.Int() pointer and ensure i.BigIntMut() & i.BigInt() don't change
	p2 := i.BigInt()
	p2.SetInt64(50)
	s.Require().NotEqual(big.NewInt(50), i.BigIntMut())
	s.Require().NotEqual(big.NewInt(50), i.BigInt())
}
