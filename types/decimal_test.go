package types

import (
	"math/big"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/stretchr/testify/require"
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
	exp := big.NewInt(100000)
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
		d1, d2                         Dec
		expMul, expDiv, expAdd, expSub Dec
	}{
		// d1          d2            MUL           DIV           ADD           SUB
		{NewDec(0), NewDec(0), NewDec(0), NewDec(0), NewDec(0), NewDec(0)},
		{NewDec(1), NewDec(0), NewDec(0), NewDec(0), NewDec(1), NewDec(1)},
		{NewDec(0), NewDec(1), NewDec(0), NewDec(0), NewDec(1), NewDec(-1)},
		{NewDec(0), NewDec(-1), NewDec(0), NewDec(0), NewDec(-1), NewDec(1)},
		{NewDec(-1), NewDec(0), NewDec(0), NewDec(0), NewDec(-1), NewDec(-1)},

		{NewDec(1), NewDec(1), NewDec(1), NewDec(1), NewDec(2), NewDec(0)},
		{NewDec(-1), NewDec(-1), NewDec(1), NewDec(1), NewDec(-2), NewDec(0)},
		{NewDec(1), NewDec(-1), NewDec(-1), NewDec(-1), NewDec(0), NewDec(2)},
		{NewDec(-1), NewDec(1), NewDec(-1), NewDec(-1), NewDec(0), NewDec(-2)},

		{NewDec(3), NewDec(7), NewDec(21), NewDecWithPrec(4285714286, 10), NewDec(10), NewDec(-4)},
		{NewDec(2), NewDec(4), NewDec(8), NewDecWithPrec(5, 1), NewDec(6), NewDec(-2)},
		{NewDec(100), NewDec(100), NewDec(10000), NewDec(1), NewDec(200), NewDec(0)},

		{NewDecWithPrec(15, 1), NewDecWithPrec(15, 1), NewDecWithPrec(225, 2),
			NewDec(1), NewDec(3), NewDec(0)},
		{NewDecWithPrec(3333, 4), NewDecWithPrec(333, 4), NewDecWithPrec(1109889, 8),
			NewDecWithPrec(10009009009, 9), NewDecWithPrec(3666, 4), NewDecWithPrec(3, 1)},
	}

	for tcIndex, tc := range tests {
		resAdd := tc.d1.Add(tc.d2)
		resSub := tc.d1.Sub(tc.d2)
		resMul := tc.d1.Mul(tc.d2)
		require.True(t, tc.expAdd.Equal(resAdd), "exp %v, res %v, tc %d", tc.expAdd, resAdd, tcIndex)
		require.True(t, tc.expSub.Equal(resSub), "exp %v, res %v, tc %d", tc.expSub, resSub, tcIndex)
		require.True(t, tc.expMul.Equal(resMul), "exp %v, res %v, tc %d", tc.expMul, resMul, tcIndex)

		if tc.d2.IsZero() { // panic for divide by zero
			require.Panics(t, func() { tc.d1.Quo(tc.d2) })
		} else {
			resDiv := tc.d1.Quo(tc.d2)
			require.True(t, tc.expDiv.Equal(resDiv), "exp %v, res %v, tc %d", tc.expDiv.String(), resDiv.String(), tcIndex)
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

func TestToLeftPadded(t *testing.T) {
	tests := []struct {
		dec    Dec
		digits int8
		exp    string
	}{
		{mustNewDecFromStr(t, "33.3"), 8, "00000033"},
		{mustNewDecFromStr(t, "50"), 8, "00000050"},
		{mustNewDecFromStr(t, "333"), 8, "00000333"},
		{mustNewDecFromStr(t, "333"), 12, "000000000333"},
		{mustNewDecFromStr(t, "0.3333"), 8, "00000000"},
	}
	for tcIndex, tc := range tests {
		res := tc.dec.ToLeftPadded(tc.digits)
		require.Equal(t, tc.exp, res, "incorrect left padding, tc %d", tcIndex)
	}
}

var cdc = codec.New()

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

	bz, err := cdc.MarshalBinary(d)
	require.NoError(t, err)

	var d2 Dec
	err = cdc.UnmarshalBinary(bz, &d2)
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
	bz, err := cdc.MarshalBinary(obj)
	require.Nil(t, err)

	var obj2 testDEmbedStruct
	err = cdc.UnmarshalBinary(bz, &obj2)
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
		"19844653375691057515930281852116324640.0000000000",
		dec3.String(),
	)
}
