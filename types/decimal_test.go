package types

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
)

// create a decimal from a decimal string (ex. "1234.5678")
func mustNewDecFromStr(t *testing.T, str string) (d Dec) {
	d, err := NewDecFromStr(str)
	require.NoError(t, err)
	return d
}

//_______________________________________

func TestPrecisionMultiplier(t *testing.T) {
	res := precisionMultiplier(5)
	exp := big.NewInt(10000000000000)
	require.Equal(t, 0, res.Cmp(exp), "equality was incorrect, res %v, exp %v", res, exp)
}

func TestNewDecFromStr(t *testing.T) {
	largeBigInt, success := new(big.Int).SetString("3144605511029693144278234343371835", 10)
	require.True(t, success)
	tests := []struct {
		decimalStr string
		expErr     bool
		exp        Dec
	}{
		{"", true, Dec{}},
		{"0.-75", true, Dec{}},
		{"0", false, NewDec(0)},
		{"1", false, NewDec(1)},
		{"1.1", false, NewDecWithPrec(11, 1)},
		{"0.75", false, NewDecWithPrec(75, 2)},
		{"0.8", false, NewDecWithPrec(8, 1)},
		{"0.11111", false, NewDecWithPrec(11111, 5)},
		{"314460551102969.3144278234343371835", true, NewDec(3141203149163817869)},
		{"314460551102969314427823434337.1835718092488231350",
			true, NewDecFromBigIntWithPrec(largeBigInt, 4)},
		{"314460551102969314427823434337.1835",
			false, NewDecFromBigIntWithPrec(largeBigInt, 4)},
		{".", true, Dec{}},
		{".0", true, NewDec(0)},
		{"1.", true, NewDec(1)},
		{"foobar", true, Dec{}},
		{"0.foobar", true, Dec{}},
		{"0.foobar.", true, Dec{}},
	}

	for tcIndex, tc := range tests {
		res, err := NewDecFromStr(tc.decimalStr)
		if tc.expErr {
			require.NotNil(t, err, "error expected, decimalStr %v, tc %v", tc.decimalStr, tcIndex)
		} else {
			require.Nil(t, err, "unexpected error, decimalStr %v, tc %v", tc.decimalStr, tcIndex)
			require.True(t, res.Equal(tc.exp), "equality was incorrect, res %v, exp %v, tc %v", res, tc.exp, tcIndex)
		}

		// negative tc
		res, err = NewDecFromStr("-" + tc.decimalStr)
		if tc.expErr {
			require.NotNil(t, err, "error expected, decimalStr %v, tc %v", tc.decimalStr, tcIndex)
		} else {
			require.Nil(t, err, "unexpected error, decimalStr %v, tc %v", tc.decimalStr, tcIndex)
			exp := tc.exp.Mul(NewDec(-1))
			require.True(t, res.Equal(exp), "equality was incorrect, res %v, exp %v, tc %v", res, exp, tcIndex)
		}
	}
}

func TestDecString(t *testing.T) {
	tests := []struct {
		d    Dec
		want string
	}{
		{NewDec(0), "0.000000000000000000"},
		{NewDec(1), "1.000000000000000000"},
		{NewDec(10), "10.000000000000000000"},
		{NewDec(12340), "12340.000000000000000000"},
		{NewDecWithPrec(12340, 4), "1.234000000000000000"},
		{NewDecWithPrec(12340, 5), "0.123400000000000000"},
		{NewDecWithPrec(12340, 8), "0.000123400000000000"},
		{NewDecWithPrec(1009009009009009009, 17), "10.090090090090090090"},
	}
	for tcIndex, tc := range tests {
		assert.Equal(t, tc.want, tc.d.String(), "bad String(), index: %v", tcIndex)
	}
}

func TestEqualities(t *testing.T) {
	tests := []struct {
		d1, d2     Dec
		gt, lt, eq bool
	}{
		{NewDec(0), NewDec(0), false, false, true},
		{NewDecWithPrec(0, 2), NewDecWithPrec(0, 4), false, false, true},
		{NewDecWithPrec(100, 0), NewDecWithPrec(100, 0), false, false, true},
		{NewDecWithPrec(-100, 0), NewDecWithPrec(-100, 0), false, false, true},
		{NewDecWithPrec(-1, 1), NewDecWithPrec(-1, 1), false, false, true},
		{NewDecWithPrec(3333, 3), NewDecWithPrec(3333, 3), false, false, true},

		{NewDecWithPrec(0, 0), NewDecWithPrec(3333, 3), false, true, false},
		{NewDecWithPrec(0, 0), NewDecWithPrec(100, 0), false, true, false},
		{NewDecWithPrec(-1, 0), NewDecWithPrec(3333, 3), false, true, false},
		{NewDecWithPrec(-1, 0), NewDecWithPrec(100, 0), false, true, false},
		{NewDecWithPrec(1111, 3), NewDecWithPrec(100, 0), false, true, false},
		{NewDecWithPrec(1111, 3), NewDecWithPrec(3333, 3), false, true, false},
		{NewDecWithPrec(-3333, 3), NewDecWithPrec(-1111, 3), false, true, false},

		{NewDecWithPrec(3333, 3), NewDecWithPrec(0, 0), true, false, false},
		{NewDecWithPrec(100, 0), NewDecWithPrec(0, 0), true, false, false},
		{NewDecWithPrec(3333, 3), NewDecWithPrec(-1, 0), true, false, false},
		{NewDecWithPrec(100, 0), NewDecWithPrec(-1, 0), true, false, false},
		{NewDecWithPrec(100, 0), NewDecWithPrec(1111, 3), true, false, false},
		{NewDecWithPrec(3333, 3), NewDecWithPrec(1111, 3), true, false, false},
		{NewDecWithPrec(-1111, 3), NewDecWithPrec(-3333, 3), true, false, false},
	}

	for tcIndex, tc := range tests {
		require.Equal(t, tc.gt, tc.d1.GT(tc.d2), "GT result is incorrect, tc %d", tcIndex)
		require.Equal(t, tc.lt, tc.d1.LT(tc.d2), "LT result is incorrect, tc %d", tcIndex)
		require.Equal(t, tc.eq, tc.d1.Equal(tc.d2), "equality result is incorrect, tc %d", tcIndex)
	}

}

func TestDecsEqual(t *testing.T) {
	tests := []struct {
		d1s, d2s []Dec
		eq       bool
	}{
		{[]Dec{NewDec(0)}, []Dec{NewDec(0)}, true},
		{[]Dec{NewDec(0)}, []Dec{NewDec(1)}, false},
		{[]Dec{NewDec(0)}, []Dec{}, false},
		{[]Dec{NewDec(0), NewDec(1)}, []Dec{NewDec(0), NewDec(1)}, true},
		{[]Dec{NewDec(1), NewDec(0)}, []Dec{NewDec(1), NewDec(0)}, true},
		{[]Dec{NewDec(1), NewDec(0)}, []Dec{NewDec(0), NewDec(1)}, false},
		{[]Dec{NewDec(1), NewDec(0)}, []Dec{NewDec(1)}, false},
		{[]Dec{NewDec(1), NewDec(2)}, []Dec{NewDec(2), NewDec(4)}, false},
		{[]Dec{NewDec(3), NewDec(18)}, []Dec{NewDec(1), NewDec(6)}, false},
	}

	for tcIndex, tc := range tests {
		require.Equal(t, tc.eq, DecsEqual(tc.d1s, tc.d2s), "equality of decional arrays is incorrect, tc %d", tcIndex)
		require.Equal(t, tc.eq, DecsEqual(tc.d2s, tc.d1s), "equality of decional arrays is incorrect (converse), tc %d", tcIndex)
	}
}

func TestArithmetic(t *testing.T) {
	tests := []struct {
		d1, d2                                Dec
		expMul, expMulTruncate                Dec
		expQuo, expQuoRoundUp, expQuoTruncate Dec
		expAdd, expSub                        Dec
	}{
		//  d1         d2         MUL    MulTruncate    QUO    QUORoundUp QUOTrunctate  ADD         SUB
		{NewDec(0), NewDec(0), NewDec(0), NewDec(0), NewDec(0), NewDec(0), NewDec(0), NewDec(0), NewDec(0)},
		{NewDec(1), NewDec(0), NewDec(0), NewDec(0), NewDec(0), NewDec(0), NewDec(0), NewDec(1), NewDec(1)},
		{NewDec(0), NewDec(1), NewDec(0), NewDec(0), NewDec(0), NewDec(0), NewDec(0), NewDec(1), NewDec(-1)},
		{NewDec(0), NewDec(-1), NewDec(0), NewDec(0), NewDec(0), NewDec(0), NewDec(0), NewDec(-1), NewDec(1)},
		{NewDec(-1), NewDec(0), NewDec(0), NewDec(0), NewDec(0), NewDec(0), NewDec(0), NewDec(-1), NewDec(-1)},

		{NewDec(1), NewDec(1), NewDec(1), NewDec(1), NewDec(1), NewDec(1), NewDec(1), NewDec(2), NewDec(0)},
		{NewDec(-1), NewDec(-1), NewDec(1), NewDec(1), NewDec(1), NewDec(1), NewDec(1), NewDec(-2), NewDec(0)},
		{NewDec(1), NewDec(-1), NewDec(-1), NewDec(-1), NewDec(-1), NewDec(-1), NewDec(-1), NewDec(0), NewDec(2)},
		{NewDec(-1), NewDec(1), NewDec(-1), NewDec(-1), NewDec(-1), NewDec(-1), NewDec(-1), NewDec(0), NewDec(-2)},

		{NewDec(3), NewDec(7), NewDec(21), NewDec(21),
			NewDecWithPrec(428571428571428571, 18), NewDecWithPrec(428571428571428572, 18), NewDecWithPrec(428571428571428571, 18),
			NewDec(10), NewDec(-4)},
		{NewDec(2), NewDec(4), NewDec(8), NewDec(8), NewDecWithPrec(5, 1), NewDecWithPrec(5, 1), NewDecWithPrec(5, 1),
			NewDec(6), NewDec(-2)},

		{NewDec(100), NewDec(100), NewDec(10000), NewDec(10000), NewDec(1), NewDec(1), NewDec(1), NewDec(200), NewDec(0)},

		{NewDecWithPrec(15, 1), NewDecWithPrec(15, 1), NewDecWithPrec(225, 2), NewDecWithPrec(225, 2),
			NewDec(1), NewDec(1), NewDec(1), NewDec(3), NewDec(0)},
		{NewDecWithPrec(3333, 4), NewDecWithPrec(333, 4), NewDecWithPrec(1109889, 8), NewDecWithPrec(1109889, 8),
			MustNewDecFromStr("10.009009009009009009"), MustNewDecFromStr("10.009009009009009010"), MustNewDecFromStr("10.009009009009009009"),
			NewDecWithPrec(3666, 4), NewDecWithPrec(3, 1)},
	}

	for tcIndex, tc := range tests {
		resAdd := tc.d1.Add(tc.d2)
		resSub := tc.d1.Sub(tc.d2)
		resMul := tc.d1.Mul(tc.d2)
		resMulTruncate := tc.d1.MulTruncate(tc.d2)
		require.True(t, tc.expAdd.Equal(resAdd), "exp %v, res %v, tc %d", tc.expAdd, resAdd, tcIndex)
		require.True(t, tc.expSub.Equal(resSub), "exp %v, res %v, tc %d", tc.expSub, resSub, tcIndex)
		require.True(t, tc.expMul.Equal(resMul), "exp %v, res %v, tc %d", tc.expMul, resMul, tcIndex)
		require.True(t, tc.expMulTruncate.Equal(resMulTruncate), "exp %v, res %v, tc %d", tc.expMulTruncate, resMulTruncate, tcIndex)

		if tc.d2.IsZero() { // panic for divide by zero
			require.Panics(t, func() { tc.d1.Quo(tc.d2) })
		} else {
			resQuo := tc.d1.Quo(tc.d2)
			require.True(t, tc.expQuo.Equal(resQuo), "exp %v, res %v, tc %d", tc.expQuo.String(), resQuo.String(), tcIndex)

			resQuoRoundUp := tc.d1.QuoRoundUp(tc.d2)
			require.True(t, tc.expQuoRoundUp.Equal(resQuoRoundUp), "exp %v, res %v, tc %d",
				tc.expQuoRoundUp.String(), resQuoRoundUp.String(), tcIndex)

			resQuoTruncate := tc.d1.QuoTruncate(tc.d2)
			require.True(t, tc.expQuoTruncate.Equal(resQuoTruncate), "exp %v, res %v, tc %d",
				tc.expQuoTruncate.String(), resQuoTruncate.String(), tcIndex)
		}
	}
}

func TestBankerRoundChop(t *testing.T) {
	tests := []struct {
		d1  Dec
		exp int64
	}{
		{mustNewDecFromStr(t, "0.25"), 0},
		{mustNewDecFromStr(t, "0"), 0},
		{mustNewDecFromStr(t, "1"), 1},
		{mustNewDecFromStr(t, "0.75"), 1},
		{mustNewDecFromStr(t, "0.5"), 0},
		{mustNewDecFromStr(t, "7.5"), 8},
		{mustNewDecFromStr(t, "1.5"), 2},
		{mustNewDecFromStr(t, "2.5"), 2},
		{mustNewDecFromStr(t, "0.545"), 1}, // 0.545-> 1 even though 5 is first decimal and 1 not even
		{mustNewDecFromStr(t, "1.545"), 2},
	}

	for tcIndex, tc := range tests {
		resNeg := tc.d1.Neg().RoundInt64()
		require.Equal(t, -1*tc.exp, resNeg, "negative tc %d", tcIndex)

		resPos := tc.d1.RoundInt64()
		require.Equal(t, tc.exp, resPos, "positive tc %d", tcIndex)
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		d1  Dec
		exp int64
	}{
		{mustNewDecFromStr(t, "0"), 0},
		{mustNewDecFromStr(t, "0.25"), 0},
		{mustNewDecFromStr(t, "0.75"), 0},
		{mustNewDecFromStr(t, "1"), 1},
		{mustNewDecFromStr(t, "1.5"), 1},
		{mustNewDecFromStr(t, "7.5"), 7},
		{mustNewDecFromStr(t, "7.6"), 7},
		{mustNewDecFromStr(t, "7.4"), 7},
		{mustNewDecFromStr(t, "100.1"), 100},
		{mustNewDecFromStr(t, "1000.1"), 1000},
	}

	for tcIndex, tc := range tests {
		resNeg := tc.d1.Neg().TruncateInt64()
		require.Equal(t, -1*tc.exp, resNeg, "negative tc %d", tcIndex)

		resPos := tc.d1.TruncateInt64()
		require.Equal(t, tc.exp, resPos, "positive tc %d", tcIndex)
	}
}

var cdc = codec.New()

func TestDecMarshalJSON(t *testing.T) {
	decimal := func(i int64) Dec {
		d := NewDec(0)
		d.Int = new(big.Int).SetInt64(i)
		return d
	}
	tests := []struct {
		name    string
		d       Dec
		want    string
		wantErr bool // if wantErr = false, will also attempt unmarshaling
	}{
		{"zero", decimal(0), "\"0.000000000000000000\"", false},
		{"one", decimal(1), "\"0.000000000000000001\"", false},
		{"ten", decimal(10), "\"0.000000000000000010\"", false},
		{"12340", decimal(12340), "\"0.000000000000012340\"", false},
		{"zeroInt", NewDec(0), "\"0.000000000000000000\"", false},
		{"oneInt", NewDec(1), "\"1.000000000000000000\"", false},
		{"tenInt", NewDec(10), "\"10.000000000000000000\"", false},
		{"12340Int", NewDec(12340), "\"12340.000000000000000000\"", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.d.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("Dec.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				assert.Equal(t, tt.want, string(got), "incorrect marshalled value")
				unmarshalledDec := NewDec(0)
				unmarshalledDec.UnmarshalJSON(got)
				assert.Equal(t, tt.d, unmarshalledDec, "incorrect unmarshalled value")
			}
		})
	}
}

func TestZeroDeserializationJSON(t *testing.T) {
	d := Dec{new(big.Int)}
	err := cdc.UnmarshalJSON([]byte(`"0"`), &d)
	require.Nil(t, err)
	err = cdc.UnmarshalJSON([]byte(`"{}"`), &d)
	require.NotNil(t, err)
}

func TestSerializationText(t *testing.T) {
	d := mustNewDecFromStr(t, "0.333")

	bz, err := d.MarshalText()
	require.NoError(t, err)

	d2 := Dec{new(big.Int)}
	err = d2.UnmarshalText(bz)
	require.NoError(t, err)
	require.True(t, d.Equal(d2), "original: %v, unmarshalled: %v", d, d2)
}

func TestSerializationGocodecJSON(t *testing.T) {
	d := mustNewDecFromStr(t, "0.333")

	bz, err := cdc.MarshalJSON(d)
	require.NoError(t, err)

	d2 := Dec{new(big.Int)}
	err = cdc.UnmarshalJSON(bz, &d2)
	require.NoError(t, err)
	require.True(t, d.Equal(d2), "original: %v, unmarshalled: %v", d, d2)
}

func TestSerializationGocodecBinary(t *testing.T) {
	d := mustNewDecFromStr(t, "0.333")

	bz, err := cdc.MarshalBinaryLengthPrefixed(d)
	require.NoError(t, err)

	var d2 Dec
	err = cdc.UnmarshalBinaryLengthPrefixed(bz, &d2)
	require.NoError(t, err)
	require.True(t, d.Equal(d2), "original: %v, unmarshalled: %v", d, d2)
}

type testDEmbedStruct struct {
	Field1 string `json:"f1"`
	Field2 int    `json:"f2"`
	Field3 Dec    `json:"f3"`
}

// TODO make work for UnmarshalJSON
func TestEmbeddedStructSerializationGocodec(t *testing.T) {
	obj := testDEmbedStruct{"foo", 10, NewDecWithPrec(1, 3)}
	bz, err := cdc.MarshalBinaryLengthPrefixed(obj)
	require.Nil(t, err)

	var obj2 testDEmbedStruct
	err = cdc.UnmarshalBinaryLengthPrefixed(bz, &obj2)
	require.Nil(t, err)

	require.Equal(t, obj.Field1, obj2.Field1)
	require.Equal(t, obj.Field2, obj2.Field2)
	require.True(t, obj.Field3.Equal(obj2.Field3), "original: %v, unmarshalled: %v", obj, obj2)
}

func TestStringOverflow(t *testing.T) {
	// two random 64 bit primes
	dec1, err := NewDecFromStr("51643150036226787134389711697696177267")
	require.NoError(t, err)
	dec2, err := NewDecFromStr("-31798496660535729618459429845579852627")
	require.NoError(t, err)
	dec3 := dec1.Add(dec2)
	require.Equal(t,
		"19844653375691057515930281852116324640.000000000000000000",
		dec3.String(),
	)
}

func TestDecMulInt(t *testing.T) {
	tests := []struct {
		sdkDec Dec
		sdkInt Int
		want   Dec
	}{
		{NewDec(10), NewInt(2), NewDec(20)},
		{NewDec(1000000), NewInt(100), NewDec(100000000)},
		{NewDecWithPrec(1, 1), NewInt(10), NewDec(1)},
		{NewDecWithPrec(1, 5), NewInt(20), NewDecWithPrec(2, 4)},
	}
	for i, tc := range tests {
		got := tc.sdkDec.MulInt(tc.sdkInt)
		require.Equal(t, tc.want, got, "Incorrect result on test case %d", i)
	}
}

func TestDecCeil(t *testing.T) {
	testCases := []struct {
		input    Dec
		expected Dec
	}{
		{NewDecWithPrec(1000000000000000, Precision), NewDec(1)},  // 0.001 => 1.0
		{NewDecWithPrec(-1000000000000000, Precision), ZeroDec()}, // -0.001 => 0.0
		{ZeroDec(), ZeroDec()}, // 0.0 => 0.0
		{NewDecWithPrec(900000000000000000, Precision), NewDec(1)},    // 0.9 => 1.0
		{NewDecWithPrec(4001000000000000000, Precision), NewDec(5)},   // 4.001 => 5.0
		{NewDecWithPrec(-4001000000000000000, Precision), NewDec(-4)}, // -4.001 => -4.0
		{NewDecWithPrec(4700000000000000000, Precision), NewDec(5)},   // 4.7 => 5.0
		{NewDecWithPrec(-4700000000000000000, Precision), NewDec(-4)}, // -4.7 => -4.0
	}

	for i, tc := range testCases {
		res := tc.input.Ceil()
		require.Equal(t, tc.expected, res, "unexpected result for test case %d, input: %v", i, tc.input)
	}
}

func TestDecSortableBytes(t *testing.T) {
	tests := []struct {
		d    Dec
		want []byte
	}{
		{NewDec(0), []byte("000000000000000000.000000000000000000")},
		{NewDec(1), []byte("000000000000000001.000000000000000000")},
		{NewDec(10), []byte("000000000000000010.000000000000000000")},
		{NewDec(12340), []byte("000000000000012340.000000000000000000")},
		{NewDecWithPrec(12340, 4), []byte("000000000000000001.234000000000000000")},
		{NewDecWithPrec(12340, 5), []byte("000000000000000000.123400000000000000")},
		{NewDecWithPrec(12340, 8), []byte("000000000000000000.000123400000000000")},
		{NewDecWithPrec(1009009009009009009, 17), []byte("000000000000000010.090090090090090090")},
		{NewDecWithPrec(-1009009009009009009, 17), []byte("-000000000000000010.090090090090090090")},
		{NewDec(1000000000000000000), []byte("max")},
		{NewDec(-1000000000000000000), []byte("--")},
	}
	for tcIndex, tc := range tests {
		assert.Equal(t, tc.want, SortableDecBytes(tc.d), "bad String(), index: %v", tcIndex)
	}

	assert.Panics(t, func() { SortableDecBytes(NewDec(1000000000000000001)) })
	assert.Panics(t, func() { SortableDecBytes(NewDec(-1000000000000000001)) })
}
