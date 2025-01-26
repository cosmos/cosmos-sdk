package math_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
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
	d1 := math.LegacyNewDecWithPrec(55, 2)
	d2 := math.LegacyNewDecWithPrec(6, 1)
	tol := math.LegacyNewDecWithPrec(1, 1)

	require.True(math.LegacyDecApproxEq(t, d1, d2, tol))

	// d1 = 0.55, d2 = 0.6, tol = 1E-5
	d1 = math.LegacyNewDecWithPrec(55, 2)
	d2 = math.LegacyNewDecWithPrec(6, 1)
	tol = math.LegacyNewDecWithPrec(1, 5)

	require.False(math.LegacyDecApproxEq(t, d1, d2, tol))

	// d1 = 0.6, d2 = 0.61, tol = 0.01
	d1 = math.LegacyNewDecWithPrec(6, 1)
	d2 = math.LegacyNewDecWithPrec(61, 2)
	tol = math.LegacyNewDecWithPrec(1, 2)

	require.True(math.LegacyDecApproxEq(t, d1, d2, tol))
}

// create a decimal from a decimal string (ex. "1234.5678")
func (s *decimalTestSuite) mustNewDecFromStr(str string) (d math.LegacyDec) {
	d, err := math.LegacyNewDecFromStr(str)
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
		exp        math.LegacyDec
	}{
		{"", true, math.LegacyDec{}},
		{"0.-75", true, math.LegacyDec{}},
		{"0", false, math.LegacyNewDec(0)},
		{"1", false, math.LegacyNewDec(1)},
		{"1.1", false, math.LegacyNewDecWithPrec(11, 1)},
		{"0.75", false, math.LegacyNewDecWithPrec(75, 2)},
		{"0.8", false, math.LegacyNewDecWithPrec(8, 1)},
		{"0.11111", false, math.LegacyNewDecWithPrec(11111, 5)},
		{"314460551102969.3144278234343371835", true, math.LegacyNewDec(3141203149163817869)},
		{
			"314460551102969314427823434337.1835718092488231350",
			true, math.LegacyNewDecFromBigIntWithPrec(largeBigInt, 4),
		},
		{
			"314460551102969314427823434337.1835",
			false, math.LegacyNewDecFromBigIntWithPrec(largeBigInt, 4),
		},
		{".", true, math.LegacyDec{}},
		{".0", true, math.LegacyNewDec(0)},
		{"1.", true, math.LegacyNewDec(1)},
		{"foobar", true, math.LegacyDec{}},
		{"0.foobar", true, math.LegacyDec{}},
		{"0.foobar.", true, math.LegacyDec{}},
		{"8888888888888888888888888888888888888888888888888888888888888888888844444440", false, math.LegacyNewDecFromBigInt(largerBigInt)},
		{"33499189745056880149688856635597007162669032647290798121690100488888732861290.034376435130433535", false, math.LegacyNewDecFromBigIntWithPrec(largestBigInt, 18)},
		{"133499189745056880149688856635597007162669032647290798121690100488888732861291", true, math.LegacyDec{}},
		{"115792089237316195423570985008687907853269984665640564039457584007913129639936", true, math.LegacyDec{}}, // 2^256
	}

	for tcIndex, tc := range tests {
		res, err := math.LegacyNewDecFromStr(tc.decimalStr)
		if tc.expErr {
			s.Require().NotNil(err, "error expected, decimalStr %v, tc %v", tc.decimalStr, tcIndex)
		} else {
			s.Require().Nil(err, "unexpected error, decimalStr %v, tc %v", tc.decimalStr, tcIndex)
			s.Require().True(res.Equal(tc.exp), "equality was incorrect, res %v, expTruncated %v, tc %v", res, tc.exp, tcIndex)
		}

		// negative tc
		res, err = math.LegacyNewDecFromStr("-" + tc.decimalStr)
		if tc.expErr {
			s.Require().NotNil(err, "error expected, decimalStr %v, tc %v", tc.decimalStr, tcIndex)
		} else {
			s.Require().Nil(err, "unexpected error, decimalStr %v, tc %v", tc.decimalStr, tcIndex)
			exp := tc.exp.Mul(math.LegacyNewDec(-1))
			s.Require().True(res.Equal(exp), "equality was incorrect, res %v, expTruncated %v, tc %v", res, exp, tcIndex)
		}
	}
}

func (s *decimalTestSuite) TestDecString() {
	tests := []struct {
		d    math.LegacyDec
		want string
	}{
		{math.LegacyNewDec(0), "0.000000000000000000"},
		{math.LegacyNewDec(1), "1.000000000000000000"},
		{math.LegacyNewDec(10), "10.000000000000000000"},
		{math.LegacyNewDec(12340), "12340.000000000000000000"},
		{math.LegacyNewDecWithPrec(12340, 4), "1.234000000000000000"},
		{math.LegacyNewDecWithPrec(12340, 5), "0.123400000000000000"},
		{math.LegacyNewDecWithPrec(12340, 8), "0.000123400000000000"},
		{math.LegacyNewDecWithPrec(1009009009009009009, 17), "10.090090090090090090"},
	}
	for tcIndex, tc := range tests {
		s.Require().Equal(tc.want, tc.d.String(), "bad String(), index: %v", tcIndex)
	}
}

func (s *decimalTestSuite) TestDecFloat64() {
	tests := []struct {
		d    math.LegacyDec
		want float64
	}{
		{math.LegacyNewDec(0), 0.000000000000000000},
		{math.LegacyNewDec(1), 1.000000000000000000},
		{math.LegacyNewDec(10), 10.000000000000000000},
		{math.LegacyNewDec(12340), 12340.000000000000000000},
		{math.LegacyNewDecWithPrec(12340, 4), 1.234000000000000000},
		{math.LegacyNewDecWithPrec(12340, 5), 0.123400000000000000},
		{math.LegacyNewDecWithPrec(12340, 8), 0.000123400000000000},
		{math.LegacyNewDecWithPrec(1009009009009009009, 17), 10.090090090090090090},
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
		d1, d2     math.LegacyDec
		gt, lt, eq bool
	}{
		{math.LegacyNewDec(0), math.LegacyNewDec(0), false, false, true},
		{math.LegacyNewDecWithPrec(0, 2), math.LegacyNewDecWithPrec(0, 4), false, false, true},
		{math.LegacyNewDecWithPrec(100, 0), math.LegacyNewDecWithPrec(100, 0), false, false, true},
		{math.LegacyNewDecWithPrec(-100, 0), math.LegacyNewDecWithPrec(-100, 0), false, false, true},
		{math.LegacyNewDecWithPrec(-1, 1), math.LegacyNewDecWithPrec(-1, 1), false, false, true},
		{math.LegacyNewDecWithPrec(3333, 3), math.LegacyNewDecWithPrec(3333, 3), false, false, true},

		{math.LegacyNewDecWithPrec(0, 0), math.LegacyNewDecWithPrec(3333, 3), false, true, false},
		{math.LegacyNewDecWithPrec(0, 0), math.LegacyNewDecWithPrec(100, 0), false, true, false},
		{math.LegacyNewDecWithPrec(-1, 0), math.LegacyNewDecWithPrec(3333, 3), false, true, false},
		{math.LegacyNewDecWithPrec(-1, 0), math.LegacyNewDecWithPrec(100, 0), false, true, false},
		{math.LegacyNewDecWithPrec(1111, 3), math.LegacyNewDecWithPrec(100, 0), false, true, false},
		{math.LegacyNewDecWithPrec(1111, 3), math.LegacyNewDecWithPrec(3333, 3), false, true, false},
		{math.LegacyNewDecWithPrec(-3333, 3), math.LegacyNewDecWithPrec(-1111, 3), false, true, false},

		{math.LegacyNewDecWithPrec(3333, 3), math.LegacyNewDecWithPrec(0, 0), true, false, false},
		{math.LegacyNewDecWithPrec(100, 0), math.LegacyNewDecWithPrec(0, 0), true, false, false},
		{math.LegacyNewDecWithPrec(3333, 3), math.LegacyNewDecWithPrec(-1, 0), true, false, false},
		{math.LegacyNewDecWithPrec(100, 0), math.LegacyNewDecWithPrec(-1, 0), true, false, false},
		{math.LegacyNewDecWithPrec(100, 0), math.LegacyNewDecWithPrec(1111, 3), true, false, false},
		{math.LegacyNewDecWithPrec(3333, 3), math.LegacyNewDecWithPrec(1111, 3), true, false, false},
		{math.LegacyNewDecWithPrec(-1111, 3), math.LegacyNewDecWithPrec(-3333, 3), true, false, false},
	}

	for tcIndex, tc := range tests {
		s.Require().Equal(tc.gt, tc.d1.GT(tc.d2), "GT result is incorrect, tc %d", tcIndex)
		s.Require().Equal(tc.lt, tc.d1.LT(tc.d2), "LT result is incorrect, tc %d", tcIndex)
		s.Require().Equal(tc.eq, tc.d1.Equal(tc.d2), "equality result is incorrect, tc %d", tcIndex)
	}
}

func (s *decimalTestSuite) TestDecsEqual() {
	tests := []struct {
		d1s, d2s []math.LegacyDec
		eq       bool
	}{
		{[]math.LegacyDec{math.LegacyNewDec(0)}, []math.LegacyDec{math.LegacyNewDec(0)}, true},
		{[]math.LegacyDec{math.LegacyNewDec(0)}, []math.LegacyDec{math.LegacyNewDec(1)}, false},
		{[]math.LegacyDec{math.LegacyNewDec(0)}, []math.LegacyDec{}, false},
		{[]math.LegacyDec{math.LegacyNewDec(0), math.LegacyNewDec(1)}, []math.LegacyDec{math.LegacyNewDec(0), math.LegacyNewDec(1)}, true},
		{[]math.LegacyDec{math.LegacyNewDec(1), math.LegacyNewDec(0)}, []math.LegacyDec{math.LegacyNewDec(1), math.LegacyNewDec(0)}, true},
		{[]math.LegacyDec{math.LegacyNewDec(1), math.LegacyNewDec(0)}, []math.LegacyDec{math.LegacyNewDec(0), math.LegacyNewDec(1)}, false},
		{[]math.LegacyDec{math.LegacyNewDec(1), math.LegacyNewDec(0)}, []math.LegacyDec{math.LegacyNewDec(1)}, false},
		{[]math.LegacyDec{math.LegacyNewDec(1), math.LegacyNewDec(2)}, []math.LegacyDec{math.LegacyNewDec(2), math.LegacyNewDec(4)}, false},
		{[]math.LegacyDec{math.LegacyNewDec(3), math.LegacyNewDec(18)}, []math.LegacyDec{math.LegacyNewDec(1), math.LegacyNewDec(6)}, false},
	}

	for tcIndex, tc := range tests {
		s.Require().Equal(tc.eq, math.LegacyDecsEqual(tc.d1s, tc.d2s), "equality of decional arrays is incorrect, tc %d", tcIndex)
		s.Require().Equal(tc.eq, math.LegacyDecsEqual(tc.d2s, tc.d1s), "equality of decional arrays is incorrect (converse), tc %d", tcIndex)
	}
}

func (s *decimalTestSuite) TestArithmetic() {
	tests := []struct {
		d1, d2                                math.LegacyDec
		expMul, expMulTruncate, expMulRoundUp math.LegacyDec
		expQuo, expQuoRoundUp, expQuoTruncate math.LegacyDec
		expAdd, expSub                        math.LegacyDec
	}{
		//  d1         d2         MUL    MulTruncate   MulRoundUp    QUO    QUORoundUp QUOTrunctate  ADD         SUB
		{math.LegacyNewDec(0), math.LegacyNewDec(0), math.LegacyNewDec(0), math.LegacyNewDec(0), math.LegacyNewDec(0), math.LegacyNewDec(0), math.LegacyNewDec(0), math.LegacyNewDec(0), math.LegacyNewDec(0), math.LegacyNewDec(0)},
		{math.LegacyNewDec(1), math.LegacyNewDec(0), math.LegacyNewDec(0), math.LegacyNewDec(0), math.LegacyNewDec(0), math.LegacyNewDec(0), math.LegacyNewDec(0), math.LegacyNewDec(0), math.LegacyNewDec(1), math.LegacyNewDec(1)},
		{math.LegacyNewDec(0), math.LegacyNewDec(1), math.LegacyNewDec(0), math.LegacyNewDec(0), math.LegacyNewDec(0), math.LegacyNewDec(0), math.LegacyNewDec(0), math.LegacyNewDec(0), math.LegacyNewDec(1), math.LegacyNewDec(-1)},
		{math.LegacyNewDec(0), math.LegacyNewDec(-1), math.LegacyNewDec(0), math.LegacyNewDec(0), math.LegacyNewDec(0), math.LegacyNewDec(0), math.LegacyNewDec(0), math.LegacyNewDec(0), math.LegacyNewDec(-1), math.LegacyNewDec(1)},
		{math.LegacyNewDec(-1), math.LegacyNewDec(0), math.LegacyNewDec(0), math.LegacyNewDec(0), math.LegacyNewDec(0), math.LegacyNewDec(0), math.LegacyNewDec(0), math.LegacyNewDec(0), math.LegacyNewDec(-1), math.LegacyNewDec(-1)},

		{math.LegacyNewDec(1), math.LegacyNewDec(1), math.LegacyNewDec(1), math.LegacyNewDec(1), math.LegacyNewDec(1), math.LegacyNewDec(1), math.LegacyNewDec(1), math.LegacyNewDec(1), math.LegacyNewDec(2), math.LegacyNewDec(0)},
		{math.LegacyNewDec(-1), math.LegacyNewDec(-1), math.LegacyNewDec(1), math.LegacyNewDec(1), math.LegacyNewDec(1), math.LegacyNewDec(1), math.LegacyNewDec(1), math.LegacyNewDec(1), math.LegacyNewDec(-2), math.LegacyNewDec(0)},
		{math.LegacyNewDec(1), math.LegacyNewDec(-1), math.LegacyNewDec(-1), math.LegacyNewDec(-1), math.LegacyNewDec(-1), math.LegacyNewDec(-1), math.LegacyNewDec(-1), math.LegacyNewDec(-1), math.LegacyNewDec(0), math.LegacyNewDec(2)},
		{math.LegacyNewDec(-1), math.LegacyNewDec(1), math.LegacyNewDec(-1), math.LegacyNewDec(-1), math.LegacyNewDec(-1), math.LegacyNewDec(-1), math.LegacyNewDec(-1), math.LegacyNewDec(-1), math.LegacyNewDec(0), math.LegacyNewDec(-2)},

		{
			math.LegacyNewDec(3), math.LegacyNewDec(7), math.LegacyNewDec(21), math.LegacyNewDec(21), math.LegacyNewDec(21),
			math.LegacyNewDecWithPrec(428571428571428571, 18), math.LegacyNewDecWithPrec(428571428571428572, 18), math.LegacyNewDecWithPrec(428571428571428571, 18),
			math.LegacyNewDec(10), math.LegacyNewDec(-4),
		},
		{
			math.LegacyNewDec(2), math.LegacyNewDec(4), math.LegacyNewDec(8), math.LegacyNewDec(8), math.LegacyNewDec(8), math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDecWithPrec(5, 1),
			math.LegacyNewDec(6), math.LegacyNewDec(-2),
		},

		{math.LegacyNewDec(100), math.LegacyNewDec(100), math.LegacyNewDec(10000), math.LegacyNewDec(10000), math.LegacyNewDec(10000), math.LegacyNewDec(1), math.LegacyNewDec(1), math.LegacyNewDec(1), math.LegacyNewDec(200), math.LegacyNewDec(0)},

		{
			math.LegacyNewDecWithPrec(15, 1), math.LegacyNewDecWithPrec(15, 1), math.LegacyNewDecWithPrec(225, 2), math.LegacyNewDecWithPrec(225, 2), math.LegacyNewDecWithPrec(225, 2),
			math.LegacyNewDec(1), math.LegacyNewDec(1), math.LegacyNewDec(1), math.LegacyNewDec(3), math.LegacyNewDec(0),
		},
		{
			math.LegacyNewDecWithPrec(3333, 4), math.LegacyNewDecWithPrec(333, 4), math.LegacyNewDecWithPrec(1109889, 8), math.LegacyNewDecWithPrec(1109889, 8), math.LegacyNewDecWithPrec(1109889, 8),
			math.LegacyMustNewDecFromStr("10.009009009009009009"), math.LegacyMustNewDecFromStr("10.009009009009009010"), math.LegacyMustNewDecFromStr("10.009009009009009009"),
			math.LegacyNewDecWithPrec(3666, 4), math.LegacyNewDecWithPrec(3, 1),
		},
	}

	for tcIndex, tc := range tests {

		resAdd := tc.d1.Add(tc.d2)
		resSub := tc.d1.Sub(tc.d2)
		resMul := tc.d1.Mul(tc.d2)
		resMulTruncate := tc.d1.MulTruncate(tc.d2)
		resMulRoundUp := tc.d1.MulRoundUp(tc.d2)
		s.Require().True(tc.expAdd.Equal(resAdd), "expTruncated %v, res %v, tc %d", tc.expAdd, resAdd, tcIndex)
		s.Require().True(tc.expSub.Equal(resSub), "expTruncated %v, res %v, tc %d", tc.expSub, resSub, tcIndex)
		s.Require().True(tc.expMul.Equal(resMul), "expTruncated %v, res %v, tc %d", tc.expMul, resMul, tcIndex)
		s.Require().True(tc.expMulTruncate.Equal(resMulTruncate), "expTruncated %v, res %v, tc %d", tc.expMulTruncate, resMulTruncate, tcIndex)
		s.Require().True(tc.expMulRoundUp.Equal(resMulRoundUp), "expTruncated %v, res %v, tc %d", tc.expMulRoundUp, resMulRoundUp, tcIndex)

		if tc.d2.IsZero() { // panic for divide by zero
			s.Require().Panics(func() { tc.d1.Quo(tc.d2) })
		} else {
			resQuo := tc.d1.Quo(tc.d2)
			s.Require().True(tc.expQuo.Equal(resQuo), "expTruncated %v, res %v, tc %d", tc.expQuo.String(), resQuo.String(), tcIndex)

			resQuoRoundUp := tc.d1.QuoRoundUp(tc.d2)
			s.Require().True(tc.expQuoRoundUp.Equal(resQuoRoundUp), "expTruncated %v, res %v, tc %d",
				tc.expQuoRoundUp.String(), resQuoRoundUp.String(), tcIndex)

			resQuoTruncate := tc.d1.QuoTruncate(tc.d2)
			s.Require().True(tc.expQuoTruncate.Equal(resQuoTruncate), "expTruncated %v, res %v, tc %d",
				tc.expQuoTruncate.String(), resQuoTruncate.String(), tcIndex)
		}
	}
}

func (s *decimalTestSuite) TestMulRoundUp_RoundingAtPrecisionEnd() {
	var (
		a                = math.LegacyMustNewDecFromStr("0.000000000000000009")
		b                = math.LegacyMustNewDecFromStr("0.000000000000000009")
		expectedRoundUp  = math.LegacyMustNewDecFromStr("0.000000000000000001")
		expectedTruncate = math.LegacyMustNewDecFromStr("0.000000000000000000")
	)

	actualRoundUp := a.MulRoundUp(b)
	s.Require().Equal(expectedRoundUp.String(), actualRoundUp.String(), "expTruncated %v, res %v", expectedRoundUp, actualRoundUp)

	actualTruncate := a.MulTruncate(b)
	s.Require().Equal(expectedTruncate.String(), actualTruncate.String(), "expTruncated %v, res %v", expectedRoundUp, actualTruncate)
}

func (s *decimalTestSuite) TestBankerRoundChop() {
	tests := []struct {
		d1  math.LegacyDec
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
		d1  math.LegacyDec
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
	dec1, err := math.LegacyNewDecFromStr("51643150036226787134389711697696177267")
	s.Require().NoError(err)
	dec2, err := math.LegacyNewDecFromStr("-31798496660535729618459429845579852627")
	s.Require().NoError(err)
	dec3 := dec1.Add(dec2)
	s.Require().Equal(
		"19844653375691057515930281852116324640.000000000000000000",
		dec3.String(),
	)
}

func (s *decimalTestSuite) TestDecMulInt() {
	tests := []struct {
		sdkDec math.LegacyDec
		sdkInt math.Int
		want   math.LegacyDec
	}{
		{math.LegacyNewDec(10), math.NewInt(2), math.LegacyNewDec(20)},
		{math.LegacyNewDec(1000000), math.NewInt(100), math.LegacyNewDec(100000000)},
		{math.LegacyNewDecWithPrec(1, 1), math.NewInt(10), math.LegacyNewDec(1)},
		{math.LegacyNewDecWithPrec(1, 5), math.NewInt(20), math.LegacyNewDecWithPrec(2, 4)},
	}
	for i, tc := range tests {
		got := tc.sdkDec.MulInt(tc.sdkInt)
		s.Require().Equal(tc.want, got, "Incorrect result on test case %d", i)
	}
}

func (s *decimalTestSuite) TestDecCeil() {
	testCases := []struct {
		input    math.LegacyDec
		expected math.LegacyDec
	}{
		{math.LegacyNewDecWithPrec(1000000000000000, math.LegacyPrecision), math.LegacyNewDec(1)},      // 0.001 => 1.0
		{math.LegacyNewDecWithPrec(-1000000000000000, math.LegacyPrecision), math.LegacyZeroDec()},     // -0.001 => 0.0
		{math.LegacyZeroDec(), math.LegacyZeroDec()},                                                   // 0.0 => 0.0
		{math.LegacyNewDecWithPrec(900000000000000000, math.LegacyPrecision), math.LegacyNewDec(1)},    // 0.9 => 1.0
		{math.LegacyNewDecWithPrec(4001000000000000000, math.LegacyPrecision), math.LegacyNewDec(5)},   // 4.001 => 5.0
		{math.LegacyNewDecWithPrec(-4001000000000000000, math.LegacyPrecision), math.LegacyNewDec(-4)}, // -4.001 => -4.0
		{math.LegacyNewDecWithPrec(4700000000000000000, math.LegacyPrecision), math.LegacyNewDec(5)},   // 4.7 => 5.0
		{math.LegacyNewDecWithPrec(-4700000000000000000, math.LegacyPrecision), math.LegacyNewDec(-4)}, // -4.7 => -4.0
	}

	for i, tc := range testCases {
		res := tc.input.Ceil()
		s.Require().Equal(tc.expected, res, "unexpected result for test case %d, input: %v", i, tc.input)
	}
}

func (s *decimalTestSuite) TestCeilOverflow() {
	// (2^256 * 10^18 -1) / 10^18
	d, err := math.LegacyNewDecFromStr("115792089237316195423570985008687907853269984665640564039457584007913129639935.999999999999999999")
	s.Require().NoError(err)
	s.Require().True(d.IsInValidRange())
	// this call panics because the value is too large
	s.Require().Panics(func() { d.Ceil() }, "Ceil should panic on overflow")
}

func (s *decimalTestSuite) TestPower() {
	testCases := []struct {
		input    math.LegacyDec
		power    uint64
		expected math.LegacyDec
	}{
		{math.LegacyNewDec(100), 0, math.LegacyOneDec()},                                                  // 10 ^ (0) => 1.0
		{math.LegacyOneDec(), 10, math.LegacyOneDec()},                                                    // 1.0 ^ (10) => 1.0
		{math.LegacyNewDecWithPrec(5, 1), 2, math.LegacyNewDecWithPrec(25, 2)},                            // 0.5 ^ 2 => 0.25
		{math.LegacyNewDecWithPrec(2, 1), 2, math.LegacyNewDecWithPrec(4, 2)},                             // 0.2 ^ 2 => 0.04
		{math.LegacyNewDecFromInt(math.NewInt(3)), 3, math.LegacyNewDecFromInt(math.NewInt(27))},          // 3 ^ 3 => 27
		{math.LegacyNewDecFromInt(math.NewInt(-3)), 4, math.LegacyNewDecFromInt(math.NewInt(81))},         // -3 ^ 4 = 81
		{math.LegacyNewDecWithPrec(1414213562373095049, 18), 2, math.LegacyNewDecFromInt(math.NewInt(2))}, // 1.414213562373095049 ^ 2 = 2
	}

	for i, tc := range testCases {
		res := tc.input.Power(tc.power)
		s.Require().True(tc.expected.Sub(res).Abs().LTE(math.LegacySmallestDec()), "unexpected result for test case %d, normal power, input: %v", i, tc.input)

		mutableInput := tc.input
		mutableInput.PowerMut(tc.power)
		s.Require().True(tc.expected.Sub(mutableInput).Abs().LTE(math.LegacySmallestDec()),
			"unexpected result for test case %d, input %v", i, tc.input)
		s.Require().True(res.Equal(tc.input), "unexpected result for test case %d, mutable power, input: %v", i, tc.input)
	}
}

func (s *decimalTestSuite) TestApproxRoot() {
	testCases := []struct {
		input    math.LegacyDec
		root     uint64
		expected math.LegacyDec
	}{
		{math.LegacyNewDecFromInt(math.NewInt(2)), 0, math.LegacyOneDec()},                                   // 2 ^ 0 => 1.0
		{math.LegacyNewDecWithPrec(4, 2), 0, math.LegacyOneDec()},                                            // 0.04 ^ 0 => 1.0
		{math.LegacyNewDec(0), 1, math.LegacyNewDec(0)},                                                      // 0 ^ 1 => 0
		{math.LegacyOneDec(), 10, math.LegacyOneDec()},                                                       // 1.0 ^ (0.1) => 1.0
		{math.LegacyNewDecWithPrec(25, 2), 2, math.LegacyNewDecWithPrec(5, 1)},                               // 0.25 ^ (0.5) => 0.5
		{math.LegacyNewDecWithPrec(4, 2), 2, math.LegacyNewDecWithPrec(2, 1)},                                // 0.04 ^ (0.5) => 0.2
		{math.LegacyNewDecFromInt(math.NewInt(27)), 3, math.LegacyNewDecFromInt(math.NewInt(3))},             // 27 ^ (1/3) => 3
		{math.LegacyNewDecFromInt(math.NewInt(-81)), 4, math.LegacyNewDecFromInt(math.NewInt(-3))},           // -81 ^ (0.25) => -3
		{math.LegacyNewDecFromInt(math.NewInt(2)), 2, math.LegacyNewDecWithPrec(1414213562373095049, 18)},    // 2 ^ (0.5) => 1.414213562373095049
		{math.LegacyNewDecWithPrec(1005, 3), 31536000, math.LegacyMustNewDecFromStr("1.000000000158153904")}, // 1.005 ^ (1/31536000) ≈ 1.00000000016
		{math.LegacySmallestDec(), 2, math.LegacyNewDecWithPrec(1, 9)},                                       // 1e-18 ^ (0.5) => 1e-9
		{math.LegacySmallestDec(), 3, math.LegacyMustNewDecFromStr("0.000000999999999997")},                  // 1e-18 ^ (1/3) => 1e-6
		{math.LegacyNewDecWithPrec(1, 8), 3, math.LegacyMustNewDecFromStr("0.002154434690031900")},           // 1e-8 ^ (1/3) ≈ 0.00215443469
		{math.LegacyMustNewDecFromStr("9000002314687921634000000000000000000021394871242000000000000000"), 2, math.LegacyMustNewDecFromStr("94868342004527103646332858502867.899477053226766107")},
	}

	// In the case of 1e-8 ^ (1/3), the result repeats every 5 iterations starting from iteration 24
	// (i.e. 24, 29, 34, ... give the same result) and never converges enough. The maximum number of
	// iterations (300) causes the result at iteration 300 to be returned, regardless of convergence.

	for i, tc := range testCases {
		res, err := tc.input.ApproxRoot(tc.root)
		s.Require().NoError(err)
		s.Require().True(tc.expected.Sub(res).Abs().LTE(math.LegacySmallestDec()), "unexpected result for test case %d, input: %v", i, tc.input)
	}
}

func (s *decimalTestSuite) TestApproxSqrt() {
	testCases := []struct {
		input    math.LegacyDec
		expected math.LegacyDec
	}{
		{math.LegacyOneDec(), math.LegacyOneDec()},                                 // 1.0 => 1.0
		{math.LegacyNewDecWithPrec(25, 2), math.LegacyNewDecWithPrec(5, 1)},        // 0.25 => 0.5
		{math.LegacyNewDecWithPrec(4, 2), math.LegacyNewDecWithPrec(2, 1)},         // 0.09 => 0.3
		{math.LegacyNewDec(9), math.LegacyNewDecFromInt(math.NewInt(3))},           // 9 => 3
		{math.LegacyNewDec(-9), math.LegacyNewDecFromInt(math.NewInt(-3))},         // -9 => -3
		{math.LegacyNewDec(2), math.LegacyNewDecWithPrec(1414213562373095049, 18)}, // 2 => 1.414213562373095049
		{ // 2^127 - 1 => 13043817825332782212.3495718062525083688 which rounds to 13043817825332782212.3495718062525083689
			math.LegacyNewDec(2).Power(127).Sub(math.LegacyOneDec()),
			math.LegacyMustNewDecFromStr("13043817825332782212.349571806252508369"),
		},
		{math.LegacyMustNewDecFromStr("1.000000011823380862"), math.LegacyMustNewDecFromStr("1.000000005911690414")},
	}

	for i, tc := range testCases {
		res, err := tc.input.ApproxSqrt()
		s.Require().NoError(err)
		s.Require().Equal(tc.expected, res, "unexpected result for test case %d, input: %v", i, tc.input)
	}
}

func (s *decimalTestSuite) TestDecSortableBytes() {
	tests := []struct {
		d    math.LegacyDec
		want []byte
	}{
		{math.LegacyNewDec(0), []byte("000000000000000000.000000000000000000")},
		{math.LegacyNewDec(1), []byte("000000000000000001.000000000000000000")},
		{math.LegacyNewDec(10), []byte("000000000000000010.000000000000000000")},
		{math.LegacyNewDec(12340), []byte("000000000000012340.000000000000000000")},
		{math.LegacyNewDecWithPrec(12340, 4), []byte("000000000000000001.234000000000000000")},
		{math.LegacyNewDecWithPrec(12340, 5), []byte("000000000000000000.123400000000000000")},
		{math.LegacyNewDecWithPrec(12340, 8), []byte("000000000000000000.000123400000000000")},
		{math.LegacyNewDecWithPrec(1009009009009009009, 17), []byte("000000000000000010.090090090090090090")},
		{math.LegacyNewDecWithPrec(-1009009009009009009, 17), []byte("-000000000000000010.090090090090090090")},
		{math.LegacyNewDec(1000000000000000000), []byte("max")},
		{math.LegacyNewDec(-1000000000000000000), []byte("--")},
	}
	for tcIndex, tc := range tests {
		s.Require().Equal(tc.want, math.LegacySortableDecBytes(tc.d), "bad String(), index: %v", tcIndex)
	}

	s.Require().Panics(func() { math.LegacySortableDecBytes(math.LegacyNewDec(1000000000000000001)) })
	s.Require().Panics(func() { math.LegacySortableDecBytes(math.LegacyNewDec(-1000000000000000001)) })
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
		input   math.LegacyDec
		rawBz   string
		jsonStr string
		yamlStr string
	}{
		{
			math.LegacyNewDec(0), "30",
			"\"0.000000000000000000\"",
			"\"0.000000000000000000\"\n",
		},
		{
			math.LegacyNewDecWithPrec(4, 2),
			"3430303030303030303030303030303030",
			"\"0.040000000000000000\"",
			"\"0.040000000000000000\"\n",
		},
		{
			math.LegacyNewDecWithPrec(-4, 2),
			"2D3430303030303030303030303030303030",
			"\"-0.040000000000000000\"",
			"\"-0.040000000000000000\"\n",
		},
		{
			math.LegacyNewDecWithPrec(1414213562373095049, 18),
			"31343134323133353632333733303935303439",
			"\"1.414213562373095049\"",
			"\"1.414213562373095049\"\n",
		},
		{
			math.LegacyNewDecWithPrec(-1414213562373095049, 18),
			"2D31343134323133353632333733303935303439",
			"\"-1.414213562373095049\"",
			"\"-1.414213562373095049\"\n",
		},
		{
			math.LegacyNewDecFromBigIntWithPrec(largestBigInt, 18),
			"3333343939313839373435303536383830313439363838383536363335353937303037313632363639303332363437323930373938313231363930313030343838383838373332383631323930303334333736343335313330343333353335",
			"\"33499189745056880149688856635597007162669032647290798121690100488888732861290.034376435130433535\"",
			"\"33499189745056880149688856635597007162669032647290798121690100488888732861290.034376435130433535\"\n",
		},
		{
			math.LegacyNewDecFromBigIntWithPrec(smallestBigInt, 18),
			"2D3333343939313839373435303536383830313439363838383536363335353937303037313632363639303332363437323930373938313231363930313030343838383838373332383631323930303334333736343335313330343333353335",
			"\"-33499189745056880149688856635597007162669032647290798121690100488888732861290.034376435130433535\"",
			"\"-33499189745056880149688856635597007162669032647290798121690100488888732861290.034376435130433535\"\n",
		},
		{
			math.LegacyNewDecFromBigIntWithPrec(maxInt, 18),
			"3636373439353934383732353238343430303734383434343238333137373938353033353831333334353136333233363435333939303630383435303530323434343434333636343330363435303137313838323137353635323136373637",
			"\"66749594872528440074844428317798503581334516323645399060845050244444366430645.017188217565216767\"",
			"\"66749594872528440074844428317798503581334516323645399060845050244444366430645.017188217565216767\"\n",
		},
	}

	for _, tc := range testCases {
		bz, err := tc.input.Marshal()
		s.Require().NoError(err)
		s.Require().Equal(tc.rawBz, fmt.Sprintf("%X", bz))

		var other math.LegacyDec
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
	n1 := math.LegacyNewDec(10)
	n2 := math.LegacyNewDec(1000000010)
	s.Require().Equal(n1.Mul(n2).Quo(n2), math.LegacyNewDec(10))
	s.Require().NotEqual(n1.Mul(n2).Quo(n2), n1.Quo(n2).Mul(n2))
}

func BenchmarkMarshalTo(b *testing.B) {
	b.ReportAllocs()
	bis := []struct {
		in   math.LegacyDec
		want []byte
	}{
		{
			math.LegacyNewDec(1e8), []byte{
				0x31, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
				0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
				0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
			},
		},
		{math.LegacyNewDec(0), []byte{0x30}},
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
	b1 := math.LegacyNewDec(17e2 + 8371)
	b2 := math.LegacyNewDec(4371)
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
	b1 := math.LegacyNewDec(17e2 + 8371)
	baseArr := make([]math.LegacyDec, b.N)
	for i := 0; i < b.N; i++ {
		baseArr[i] = b1.Clone()
	}
	b2 := math.LegacyNewDec(4371)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink = baseArr[i].QuoTruncateMut(b2)
	}

	if sink == nil {
		b.Fatal("Benchmark did not run")
	}
	sink = (interface{})(nil)
}

func BenchmarkLegacySqrtOnMersennePrime(b *testing.B) {
	b1 := math.LegacyNewDec(2).Power(127).Sub(math.LegacyOneDec())
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
	b1 := math.LegacyNewDec(17e2 + 8371)
	baseArr := make([]math.LegacyDec, b.N)
	for i := 0; i < b.N; i++ {
		baseArr[i] = b1.Clone()
	}
	b2 := math.LegacyNewDec(4371)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink = baseArr[i].QuoRoundupMut(b2)
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
		t.Run(tc[0], func(t *testing.T) {
			out, err := math.FormatDec(tc[0])
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
		t.Run(value, func(t *testing.T) {
			s, err := math.FormatDec(value)
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
		math.LegacyNewDecWithPrec(10, -1)
	})
}

func (s *decimalTestSuite) TestConvertToBigIntMutativeForLegacyDec() {
	r := big.NewInt(30)
	i := math.LegacyNewDecFromBigInt(r)

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

func TestQuoMut(t *testing.T) {
	specs := map[string]struct {
		dividend, divisor          math.LegacyDec
		expTruncated, expRoundedUp string
		expPanic                   bool
	}{
		"0.0000000000000000001": {
			dividend:     math.LegacyNewDecWithPrec(1, 18),
			divisor:      math.LegacyMustNewDecFromStr("10"),
			expRoundedUp: "0.000000000000000001",
			expTruncated: "0.000000000000000000",
		},
		"0.0000000000000000002": {
			dividend:     math.LegacyNewDecWithPrec(1, 18),
			divisor:      math.LegacyMustNewDecFromStr("5"),
			expRoundedUp: "0.000000000000000001",
			expTruncated: "0.000000000000000000",
		},
		"0.0000000000000000003": {
			dividend:     math.LegacyNewDecWithPrec(1, 18),
			divisor:      math.LegacyMustNewDecFromStr("3.333333333333333"),
			expRoundedUp: "0.000000000000000001",
			expTruncated: "0.000000000000000000",
		},
		"0.0000000000000000004": {
			dividend:     math.LegacyNewDecWithPrec(1, 18),
			divisor:      math.LegacyMustNewDecFromStr("2.5"),
			expRoundedUp: "0.000000000000000001",
			expTruncated: "0.000000000000000000",
		},
		"0.0000000000000000005": {
			dividend:     math.LegacyNewDecWithPrec(1, 18),
			divisor:      math.LegacyMustNewDecFromStr("2"),
			expRoundedUp: "0.000000000000000001",

			expTruncated: "0.000000000000000000",
		},
		"0.0000000000000000006": {
			dividend:     math.LegacyNewDecWithPrec(1, 18),
			divisor:      math.LegacyMustNewDecFromStr("1.666666666666666666"),
			expRoundedUp: "0.000000000000000001",

			expTruncated: "0.000000000000000000",
		},
		"0.0000000000000000007": {
			dividend:     math.LegacyNewDecWithPrec(1, 18),
			divisor:      math.LegacyMustNewDecFromStr("1.428571428571429"),
			expRoundedUp: "0.000000000000000001",

			expTruncated: "0.000000000000000000",
		},
		"0.0000000000000000008": {
			dividend:     math.LegacyNewDecWithPrec(1, 18),
			divisor:      math.LegacyMustNewDecFromStr("1.25"),
			expRoundedUp: "0.000000000000000001",

			expTruncated: "0.000000000000000000",
		},
		"0.0000000000000000009": {
			dividend:     math.LegacyNewDecWithPrec(1, 18),
			divisor:      math.LegacyMustNewDecFromStr("1.111111111111111"),
			expRoundedUp: "0.000000000000000001",

			expTruncated: "0.000000000000000000",
		},
		"-0.0000000000000000001": {
			dividend:     math.LegacyNewDecWithPrec(1, 18).Neg(),
			divisor:      math.LegacyMustNewDecFromStr("10"),
			expRoundedUp: "0.000000000000000000",
			expTruncated: "0.000000000000000000",
		},
		"-0.0000000000000000002": {
			dividend:     math.LegacyNewDecWithPrec(1, 18).Neg(),
			divisor:      math.LegacyMustNewDecFromStr("5"),
			expRoundedUp: "0.000000000000000000",
			expTruncated: "0.000000000000000000",
		},
		"-0.0000000000000000003": {
			dividend:     math.LegacyNewDecWithPrec(1, 18).Neg(),
			divisor:      math.LegacyMustNewDecFromStr("3.333333333333333"),
			expRoundedUp: "0.000000000000000000",
			expTruncated: "0.000000000000000000",
		},
		"-0.0000000000000000004": {
			dividend:     math.LegacyNewDecWithPrec(1, 18).Neg(),
			divisor:      math.LegacyMustNewDecFromStr("2.5"),
			expRoundedUp: "0.000000000000000000",
			expTruncated: "0.000000000000000000",
		},
		"-0.0000000000000000005": {
			dividend:     math.LegacyNewDecWithPrec(1, 18).Neg(),
			divisor:      math.LegacyMustNewDecFromStr("2"),
			expRoundedUp: "0.000000000000000000",
			expTruncated: "0.000000000000000000",
		},
		"-0.0000000000000000006": {
			dividend:     math.LegacyNewDecWithPrec(1, 18).Neg(),
			divisor:      math.LegacyMustNewDecFromStr("1.666666666666666666"),
			expRoundedUp: "0.000000000000000000",
			expTruncated: "0.000000000000000000",
		},
		"-0.0000000000000000007": {
			dividend:     math.LegacyNewDecWithPrec(1, 18).Neg(),
			divisor:      math.LegacyMustNewDecFromStr("1.428571428571429"),
			expRoundedUp: "0.000000000000000000",
			expTruncated: "0.000000000000000000",
		},
		"-0.0000000000000000008": {
			dividend:     math.LegacyNewDecWithPrec(1, 18).Neg(),
			divisor:      math.LegacyMustNewDecFromStr("1.25"),
			expRoundedUp: "0.000000000000000000",
			expTruncated: "0.000000000000000000",
		},
		"-0.0000000000000000009": {
			dividend:     math.LegacyNewDecWithPrec(1, 18).Neg(),
			divisor:      math.LegacyMustNewDecFromStr("1.111111111111111"),
			expRoundedUp: "0.000000000000000000",
			expTruncated: "0.000000000000000000",
		},
		"--0.0000000000000000001": {
			dividend:     math.LegacyNewDecWithPrec(1, 18).Neg(),
			divisor:      math.LegacyMustNewDecFromStr("-10"),
			expRoundedUp: "0.000000000000000001",
			expTruncated: "0.000000000000000000",
		},
		"--0.0000000000000000002": {
			dividend:     math.LegacyNewDecWithPrec(1, 18).Neg(),
			divisor:      math.LegacyMustNewDecFromStr("-5"),
			expRoundedUp: "0.000000000000000001",
			expTruncated: "0.000000000000000000",
		},
		"--0.0000000000000000003": {
			dividend:     math.LegacyNewDecWithPrec(1, 18).Neg(),
			divisor:      math.LegacyMustNewDecFromStr("-3.333333333333333"),
			expRoundedUp: "0.000000000000000001",
			expTruncated: "0.000000000000000000",
		},
		"--0.0000000000000000004": {
			dividend:     math.LegacyNewDecWithPrec(1, 18).Neg(),
			divisor:      math.LegacyMustNewDecFromStr("-2.5"),
			expRoundedUp: "0.000000000000000001",
			expTruncated: "0.000000000000000000",
		},
		"--0.0000000000000000005": {
			dividend:     math.LegacyNewDecWithPrec(1, 18).Neg(),
			divisor:      math.LegacyMustNewDecFromStr("-2"),
			expRoundedUp: "0.000000000000000001",
			expTruncated: "0.000000000000000000",
		},
		"--0.0000000000000000006": {
			dividend:     math.LegacyNewDecWithPrec(1, 18).Neg(),
			divisor:      math.LegacyMustNewDecFromStr("-1.666666666666666666"),
			expRoundedUp: "0.000000000000000001",
			expTruncated: "0.000000000000000000",
		},
		"--0.0000000000000000007": {
			dividend:     math.LegacyNewDecWithPrec(1, 18).Neg(),
			divisor:      math.LegacyMustNewDecFromStr("-1.428571428571429"),
			expRoundedUp: "0.000000000000000001",
			expTruncated: "0.000000000000000000",
		},
		"--0.0000000000000000008": {
			dividend:     math.LegacyNewDecWithPrec(1, 18).Neg(),
			divisor:      math.LegacyMustNewDecFromStr("-1.25"),
			expRoundedUp: "0.000000000000000001",
			expTruncated: "0.000000000000000000",
		},
		"--0.0000000000000000009": {
			dividend:     math.LegacyNewDecWithPrec(1, 18).Neg(),
			divisor:      math.LegacyMustNewDecFromStr("-1.111111111111111"),
			expRoundedUp: "0.000000000000000001",
			expTruncated: "0.000000000000000000",
		},
		"big / small": {
			dividend:     math.LegacyMustNewDecFromStr("999999999999999999"),
			divisor:      math.LegacyNewDecWithPrec(1, 18),
			expRoundedUp: "999999999999999999000000000000000000.000000000000000000",
			expTruncated: "999999999999999999000000000000000000.000000000000000000",
		},
		"divide by dividend": {
			dividend:     math.LegacyNewDecWithPrec(123, 0),
			divisor:      math.LegacyMustNewDecFromStr("123"),
			expRoundedUp: "1.000000000000000000",
			expTruncated: "1.000000000000000000",
		},
		"zero divided": {
			dividend:     math.LegacyNewDecWithPrec(0, 0),
			divisor:      math.LegacyMustNewDecFromStr("1"),
			expRoundedUp: "0.000000000000000000",
			expTruncated: "0.000000000000000000",
		},
		"zero divided by negative value": {
			dividend:     math.LegacyNewDecWithPrec(0, 0),
			divisor:      math.LegacyMustNewDecFromStr("-1"),
			expRoundedUp: "0.000000000000000000",
			expTruncated: "0.000000000000000000",
		},
		"zero divided by zero": {
			dividend: math.LegacyNewDecWithPrec(0, 0),
			divisor:  math.LegacyMustNewDecFromStr("0"),
			expPanic: true,
		},
		"divide by zero": {
			dividend: math.LegacyNewDecWithPrec(1, 0),
			divisor:  math.LegacyMustNewDecFromStr("0"),
			expPanic: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			t.Run("round up", func(t *testing.T) {
				t.Parallel()
				if !spec.expPanic {
					got := spec.dividend.Clone().QuoRoundupMut(spec.divisor.Clone())
					require.Equal(t, spec.expRoundedUp, got.String())
					return
				}
				require.Panics(t, func() {
					_ = spec.dividend.Clone().QuoRoundupMut(spec.divisor.Clone())
				})
			})
			t.Run("truncate", func(t *testing.T) {
				t.Parallel()
				if !spec.expPanic {
					got := spec.dividend.Clone().QuoTruncateMut(spec.divisor.Clone())
					require.Equal(t, spec.expTruncated, got.String())
					return
				}
				require.Panics(t, func() {
					_ = spec.dividend.Clone().QuoTruncateMut(spec.divisor.Clone())
				})
			})
		})
	}
}

func Test_DocumentLegacyAsymmetry(t *testing.T) {
	zeroDec := math.LegacyZeroDec()
	emptyDec := math.LegacyDec{}

	zeroDecBz, err := zeroDec.Marshal()
	require.NoError(t, err)
	zeroDecJSON, err := zeroDec.MarshalJSON()
	require.NoError(t, err)

	emptyDecBz, err := emptyDec.Marshal()
	require.NoError(t, err)
	emptyDecJSON, err := emptyDec.MarshalJSON()
	require.NoError(t, err)

	// makes sense, zero and empty are semantically different and render differently
	require.NotEqual(t, zeroDecJSON, emptyDecJSON)
	// but on the proto wire they encode to the same bytes
	require.Equal(t, zeroDecBz, emptyDecBz)

	// zero values are symmetrical
	zeroDecRoundTrip := math.LegacyDec{}
	err = zeroDecRoundTrip.Unmarshal(zeroDecBz)
	require.NoError(t, err)
	zeroDecRoundTripJSON, err := zeroDecRoundTrip.MarshalJSON()
	require.NoError(t, err)
	require.Equal(t, zeroDecJSON, zeroDecRoundTripJSON)
	require.Equal(t, zeroDec, zeroDecRoundTrip)

	// empty values are not
	emptyDecRoundTrip := math.LegacyDec{}
	err = emptyDecRoundTrip.Unmarshal(emptyDecBz)
	require.NoError(t, err)
	emptyDecRoundTripJSON, err := emptyDecRoundTrip.MarshalJSON()
	require.NoError(t, err)

	// !!! this is the key point, they are not equal, it looks like a bug
	require.NotEqual(t, emptyDecJSON, emptyDecRoundTripJSON)
	require.NotEqual(t, emptyDec, emptyDecRoundTrip)
}

// 2^256 * 10^18 -1
const maxValidDecNumber = "115792089237316195423570985008687907853269984665640564039457584007913129639935999999999999999999"

func TestDecOpsWithinLimits(t *testing.T) {
	maxValid, ok := new(big.Int).SetString(maxValidDecNumber, 10)
	require.True(t, ok)
	minValid := new(big.Int).Neg(maxValid)
	specs := map[string]struct {
		src    *big.Int
		expErr bool
	}{
		"max": {
			src: maxValid,
		},
		"max + 1": {
			src:    new(big.Int).Add(maxValid, big.NewInt(1)),
			expErr: true,
		},
		"min": {
			src: minValid,
		},
		"min - 1": {
			src:    new(big.Int).Sub(minValid, big.NewInt(1)),
			expErr: true,
		},
		"max Int": {
			// max Int is 2^256 -1
			src: math.NewIntFromBigInt(new(big.Int).Sub(new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil), big.NewInt(1))).BigIntMut(),
		},
		"min Int": {
			// max Int is -1 *(2^256 -1)
			src: math.NewIntFromBigInt(new(big.Int).Neg(new(big.Int).Sub(new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil), big.NewInt(1)))).BigIntMut(),
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			src := math.LegacyNewDecFromBigIntWithPrec(spec.src, 18)

			ops := map[string]struct {
				fn func(src math.LegacyDec) math.LegacyDec
			}{
				"AddMut": {
					fn: func(src math.LegacyDec) math.LegacyDec { return src.AddMut(math.LegacyNewDec(0)) },
				},
				"SubMut": {
					fn: func(src math.LegacyDec) math.LegacyDec { return src.SubMut(math.LegacyNewDec(0)) },
				},
				"MulMut": {
					fn: func(src math.LegacyDec) math.LegacyDec { return src.MulMut(math.LegacyNewDec(1)) },
				},
				"MulTruncateMut": {
					fn: func(src math.LegacyDec) math.LegacyDec { return src.MulTruncateMut(math.LegacyNewDec(1)) },
				},
				"MulRoundUpMut": {
					fn: func(src math.LegacyDec) math.LegacyDec { return src.MulRoundUpMut(math.LegacyNewDec(1)) },
				},
				"MulIntMut": {
					fn: func(src math.LegacyDec) math.LegacyDec { return src.MulIntMut(math.NewInt(1)) },
				},
				"MulInt64Mut": {
					fn: func(src math.LegacyDec) math.LegacyDec { return src.MulInt64Mut(1) },
				},
				"QuoMut": {
					fn: func(src math.LegacyDec) math.LegacyDec { return src.QuoMut(math.LegacyNewDec(1)) },
				},
				"QuoTruncateMut": {
					fn: func(src math.LegacyDec) math.LegacyDec { return src.QuoTruncateMut(math.LegacyNewDec(1)) },
				},
				"QuoRoundupMut": {
					fn: func(src math.LegacyDec) math.LegacyDec { return src.QuoRoundupMut(math.LegacyNewDec(1)) },
				},
			}
			for name, op := range ops {
				t.Run(name, func(t *testing.T) {
					if spec.expErr {
						assert.Panics(t, func() {
							got := op.fn(src)
							t.Log(got.String())
						})
						return
					}
					exp := src.String()
					// exp no panics
					got := op.fn(src)
					assert.Equal(t, exp, got.String())
				})
			}
		})
	}
}

func TestDecCeilLimits(t *testing.T) {
	maxValid, ok := new(big.Int).SetString(maxValidDecNumber, 10)
	require.True(t, ok)
	minValid := new(big.Int).Neg(maxValid)

	specs := map[string]struct {
		src    *big.Int
		exp    string
		expErr bool
	}{
		"max": {
			src:    maxValid,
			expErr: true,
		},
		"max + 1": {
			src:    new(big.Int).Add(maxValid, big.NewInt(1)),
			expErr: true,
		},
		"max - 1e18, previous full number": {
			src: new(big.Int).Sub(maxValid, new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)),
			exp: "115792089237316195423570985008687907853269984665640564039457584007913129639935.000000000000000000",
		},
		"min": {
			src: minValid,
			exp: "-115792089237316195423570985008687907853269984665640564039457584007913129639935.000000000000000000",
		},
		"min - 1": {
			src:    new(big.Int).Sub(minValid, big.NewInt(1)),
			expErr: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			src := math.LegacyNewDecFromBigIntWithPrec(spec.src, 18)
			if spec.expErr {
				assert.Panics(t, func() {
					got := src.Ceil()
					t.Log(got.String())
				})
				return
			}
			got := src.Ceil()
			assert.Equal(t, spec.exp, got.String())
		})
	}
}

func TestTruncateIntLimits(t *testing.T) {
	maxValid, ok := new(big.Int).SetString(maxValidDecNumber, 10)
	require.True(t, ok)
	minValid := new(big.Int).Neg(maxValid)

	specs := map[string]struct {
		src    *big.Int
		exp    string
		expErr bool
	}{
		"max": {
			src: maxValid,
			exp: "115792089237316195423570985008687907853269984665640564039457584007913129639935",
		},
		"max + 1": {
			src:    new(big.Int).Add(maxValid, big.NewInt(1)),
			expErr: true,
		},
		"min": {
			src: minValid,
			exp: "-115792089237316195423570985008687907853269984665640564039457584007913129639935",
		},
		"min - 1": {
			src:    new(big.Int).Sub(minValid, big.NewInt(1)),
			expErr: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			src := math.LegacyNewDecFromBigIntWithPrec(spec.src, 18)
			if spec.expErr {
				assert.Panics(t, func() {
					got := src.TruncateInt()
					t.Log(got.String())
				})
				return
			}
			got := src.TruncateInt()
			assert.Equal(t, spec.exp, got.String())
		})
	}
}

func TestRoundIntLimits(t *testing.T) {
	maxValid, ok := new(big.Int).SetString(maxValidDecNumber, 10)
	require.True(t, ok)
	minValid := new(big.Int).Neg(maxValid)
	oneE18 := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)

	specs := map[string]struct {
		src    *big.Int
		exp    string
		expErr bool
	}{
		"max -1e18; previous full number": {
			src: new(big.Int).Sub(maxValid, oneE18),
			exp: "115792089237316195423570985008687907853269984665640564039457584007913129639935",
		},
		"max": {
			src:    maxValid,
			expErr: true,
		},
		"max + 1": {
			src:    new(big.Int).Add(maxValid, big.NewInt(1)),
			expErr: true,
		},
		"min + 1e18; previous full number": {
			src: new(big.Int).Add(minValid, oneE18),
			exp: "-115792089237316195423570985008687907853269984665640564039457584007913129639935",
		},
		"min": {
			src:    minValid,
			expErr: true,
		},
		"min - 1": {
			src:    new(big.Int).Sub(minValid, big.NewInt(1)),
			expErr: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			src := math.LegacyNewDecFromBigIntWithPrec(spec.src, 18)
			t.Log(src.String())
			if spec.expErr {
				assert.Panics(t, func() {
					got := src.RoundInt()
					t.Log(got.String())
				})
				return
			}
			got := src.RoundInt()
			assert.Equal(t, spec.exp, got.String())
		})
	}
}

func BenchmarkIsInValidRange(b *testing.B) {
	maxValid, ok := new(big.Int).SetString(maxValidDecNumber, 10)
	require.True(b, ok)
	souceMax := math.LegacyNewDecFromBigIntWithPrec(maxValid, 18)
	b.ResetTimer()
	specs := map[string]math.LegacyDec{
		"max":         souceMax,
		"greater max": math.LegacyNewDecFromBigIntWithPrec(maxValid, 16),
		"min":         souceMax.Neg(),
		"lower min":   math.LegacyNewDecFromBigIntWithPrec(new(big.Int).Neg(maxValid), 16),
		"zero":        math.LegacyZeroDec(),
		"one":         math.LegacyOneDec(),
	}
	for name, source := range specs {
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = source.IsInValidRange()
			}
		})
	}
}
