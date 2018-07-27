package types

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

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
		{"0", false, NewDec(0, 0)},
		{"1", false, NewDec(1, 0)},
		{"1.1", false, NewDec(11, 1)},
		{"0.75", false, NewDec(75, 2)},
		{"0.8", false, NewDec(8, 1)},
		{"0.11111", false, NewDec(11111, 5)},
		{"314460551102969.3144278234343371835", true, NewDec(3141203149163817869, 0)},
		{"314460551102969314427823434337.1835718092488231350",
			true, NewDecFromBigInt(largeBigInt, 4)},
		{"314460551102969314427823434337.1835",
			false, NewDecFromBigInt(largeBigInt, 4)},
		{".", true, Dec{}},
		{".0", true, NewDec(0, 0)},
		{"1.", true, NewDec(1, 0)},
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
			exp := tc.exp.Mul(NewDec(-1, 0))
			require.True(t, res.Equal(exp), "equality was incorrect, res %v, exp %v, tc %v", res, exp, tcIndex)
		}
	}
}

func TestDEqualities(t *testing.T) {
	tests := []struct {
		d1, d2     Dec
		gt, lt, eq bool
	}{
		{NewDec(0, 0), NewDec(0, 0), false, false, true},
		{NewDec(0, 2), NewDec(0, 4), false, false, true},
		{NewDec(100, 0), NewDec(100, 0), false, false, true},
		{NewDec(-100, 0), NewDec(-100, 0), false, false, true},
		{NewDec(-1, 1), NewDec(-1, 1), false, false, true},
		{NewDec(3333, 3), NewDec(3333, 3), false, false, true},

		{NewDec(0, 0), NewDec(3333, 3), false, true, false},
		{NewDec(0, 0), NewDec(100, 0), false, true, false},
		{NewDec(-1, 0), NewDec(3333, 3), false, true, false},
		{NewDec(-1, 0), NewDec(100, 0), false, true, false},
		{NewDec(1111, 3), NewDec(100, 0), false, true, false},
		{NewDec(1111, 3), NewDec(3333, 3), false, true, false},
		{NewDec(-3333, 3), NewDec(-1111, 3), false, true, false},

		{NewDec(3333, 3), NewDec(0, 0), true, false, false},
		{NewDec(100, 0), NewDec(0, 0), true, false, false},
		{NewDec(3333, 3), NewDec(-1, 0), true, false, false},
		{NewDec(100, 0), NewDec(-1, 0), true, false, false},
		{NewDec(100, 0), NewDec(1111, 3), true, false, false},
		{NewDec(3333, 3), NewDec(1111, 3), true, false, false},
		{NewDec(-1111, 3), NewDec(-3333, 3), true, false, false},
	}

	for tcIndex, tc := range tests {
		require.Equal(t, tc.gt, tc.d1.GT(tc.d2), "GT result is incorrect, tc #%d", tcIndex)
		require.Equal(t, tc.lt, tc.d1.LT(tc.d2), "LT result is incorrect, tc #%d", tcIndex)
		require.Equal(t, tc.eq, tc.d1.Equal(tc.d2), "equality result is incorrect, tc #%d", tcIndex)
	}

}

//func TestArithmetic(t *testing.T) {
//tests := []struct {
//d1, d2                         Dec
//resMul, resDiv, resAdd, resSub Dec
//}{
//// d1       d2         MUL        DIV        ADD        SUB
//{NewDec(0), NewDec(0), NewDec(0), NewDec(0), NewDec(0), NewDec(0)},
//{NewDec(1), NewDec(0), NewDec(0), NewDec(0), NewDec(1), NewDec(1)},
//{NewDec(0), NewDec(1), NewDec(0), NewDec(0), NewDec(1), NewDec(-1)},
//{NewDec(0), NewDec(-1), NewDec(0), NewDec(0), NewDec(-1), NewDec(1)},
//{NewDec(-1), NewDec(0), NewDec(0), NewDec(0), NewDec(-1), NewDec(-1)},

//{NewDec(1), NewDec(1), NewDec(1), NewDec(1), NewDec(2), NewDec(0)},
//{NewDec(-1), NewDec(-1), NewDec(1), NewDec(1), NewDec(-2), NewDec(0)},
//{NewDec(1), NewDec(-1), NewDec(-1), NewDec(-1), NewDec(0), NewDec(2)},
//{NewDec(-1), NewDec(1), NewDec(-1), NewDec(-1), NewDec(0), NewDec(-2)},

//{NewDec(3), NewDec(7), NewDec(21), NewDec(3, 7), NewDec(10), NewDec(-4)},
//{NewDec(2), NewDec(4), NewDec(8), NewDec(1, 2), NewDec(6), NewDec(-2)},
//{NewDec(100), NewDec(100), NewDec(10000), NewDec(1), NewDec(200), NewDec(0)},

//{NewDec(3, 2), NewDec(3, 2), NewDec(9, 4), NewDec(1), NewDec(3), NewDec(0)},
//{NewDec(3, 7), NewDec(7, 3), NewDec(1), NewDec(9, 49), NewDec(58, 21), NewDec(-40, 21)},
//{NewDec(1, 21), NewDec(11, 5), NewDec(11, 105), NewDec(5, 231), NewDec(236, 105), NewDec(-226, 105)},
//{NewDec(-21), NewDec(3, 7), NewDec(-9), NewDec(-49), NewDec(-144, 7), NewDec(-150, 7)},
//{NewDec(100), NewDec(1, 7), NewDec(100, 7), NewDec(700), NewDec(701, 7), NewDec(699, 7)},
//}

//for tcIndex, tc := range tests {
//require.True(t, tc.resMul.Equal(tc.d1.Mul(tc.d2)), "d1 %v, d2 %v. tc #%d", tc.d1.Dec, tc.d2.Dec, tcIndex)
//require.True(t, tc.resAdd.Equal(tc.d1.Add(tc.d2)), "d1 %v, d2 %v. tc #%d", tc.d1.Dec, tc.d2.Dec, tcIndex)
//require.True(t, tc.resSub.Equal(tc.d1.Sub(tc.d2)), "d1 %v, d2 %v. tc #%d", tc.d1.Dec, tc.d2.Dec, tcIndex)

//if tc.d2.Num().IsZero() { // panic for divide by zero
//require.Panics(t, func() { tc.d1.Quo(tc.d2) })
//} else {
//require.True(t, tc.resDiv.Equal(tc.d1.Quo(tc.d2)), "d1 %v, d2 %v. tc #%d", tc.d1.Dec, tc.d2.Dec, tcIndex)
//}
//}
//}

//func TestEvaluate(t *testing.T) {
//tests := []struct {
//d1  Dec
//res int64
//}{
//{NewDec(0), 0},
//{NewDec(1), 1},
//{NewDec(1, 4), 0},
//{NewDec(1, 2), 0},
//{NewDec(3, 4), 1},
//{NewDec(5, 6), 1},
//{NewDec(3, 2), 2},
//{NewDec(5, 2), 2},
//{NewDec(6, 11), 1},  // 0.545-> 1 even though 5 is first decimal and 1 not even
//{NewDec(17, 11), 2}, // 1.545
//{NewDec(5, 11), 0},
//{NewDec(16, 11), 1},
//{NewDec(113, 12), 9},
//}

//for tcIndex, tc := range tests {
//require.Equal(t, tc.res, tc.d1.RoundInt64(), "%v. tc #%d", tc.d1, tcIndex)
//require.Equal(t, tc.res*-1, tc.d1.Mul(NewDec(-1)).RoundInt64(), "%v. tc #%d", tc.d1.Mul(NewDec(-1)), tcIndex)
//}
//}

//func TestRound(t *testing.T) {
//many3 := "333333333333333333333333333333333333333333333"
//many7 := "777777777777777777777777777777777777777777777"
//big3, worked := new(big.Int).SetString(many3, 10)
//require.True(t, worked)
//big7, worked := new(big.Int).SetString(many7, 10)
//require.True(t, worked)

//tests := []struct {
//d, res     Dec
//precFactor int64
//}{
//{NewDec(333, 777), NewDec(429, 1000), 1000},
//{Dec{new(big.Dec).SetFrac(big3, big7)}, NewDec(429, 1000), 1000},
//{Dec{new(big.Dec).SetFrac(big3, big7)}, Dec{big.NewDec(4285714286, 10000000000)}, 10000000000},
//{NewDec(1, 2), NewDec(1, 2), 1000},
//}

//for tcIndex, tc := range tests {
//require.Equal(t, tc.res, tc.d.Round(tc.precFactor), "%v", tc.d, "incorrect rounding, tc #%d", tcIndex)
//negR1, negRes := tc.d.Mul(NewDec(-1)), tc.res.Mul(NewDec(-1))
//require.Equal(t, negRes, negR1.Round(tc.precFactor), "%v", negR1, "incorrect rounding (negative case), tc #%d", tcIndex)
//}
//}

//func TestToLeftPadded(t *testing.T) {
//tests := []struct {
//dec    Dec
//digits int8
//res    string
//}{
//{NewDec(100, 3), 8, "00000033"},
//{NewDec(1, 3), 8, "00000000"},
//{NewDec(100, 2), 8, "00000050"},
//{NewDec(1000, 3), 8, "00000333"},
//{NewDec(1000, 3), 12, "000000000333"},
//}
//for tcIndex, tc := range tests {
//require.Equal(t, tc.res, tc.dec.ToLeftPadded(tc.digits), "incorrect left padding, tc #%d", tcIndex)
//}
//}

//var cdc = wire.NewCodec() //var jsonCdc JSONCodec // TODO wire.Codec

//func TestZeroSerializationJSON(t *testing.T) {
//d := NewDec(0, 1)
//err := cdc.UnmarshalJSON([]byte(`"0/1"`), &d)
//require.Nil(t, err)
//err = cdc.UnmarshalJSON([]byte(`"0/0"`), &d)
//require.NotNil(t, err)
//err = cdc.UnmarshalJSON([]byte(`"1/0"`), &d)
//require.NotNil(t, err)
//err = cdc.UnmarshalJSON([]byte(`"{}"`), &d)
//require.NotNil(t, err)
//}

//func TestSerializationText(t *testing.T) {
//d := NewDec(1, 3)

//bz, err := d.MarshalText()
//require.NoError(t, err)

//var d2 = Dec{new(big.Dec)}
//err = d2.UnmarshalText(bz)
//require.NoError(t, err)
//require.True(t, d.Equal(d2), "original: %v, unmarshalled: %v", d, d2)
//}

//func TestSerializationGoWireJSON(t *testing.T) {
//d := NewDec(1, 3)
//bz, err := cdc.MarshalJSON(d)
//require.NoError(t, err)

//var d2 Dec
//err = cdc.UnmarshalJSON(bz, &d2)
//require.NoError(t, err)
//require.True(t, d.Equal(d2), "original: %v, unmarshalled: %v", d, d2)
//}

//func TestSerializationGoWireBinary(t *testing.T) {
//d := NewDec(1, 3)
//bz, err := cdc.MarshalBinary(d)
//require.NoError(t, err)

//var d2 Dec
//err = cdc.UnmarshalBinary(bz, &d2)
//require.NoError(t, err)
//require.True(t, d.Equal(d2), "original: %v, unmarshalled: %v", d, d2)
//}

//type testEmbedStruct struct {
//Field1 string `json:"f1"`
//Field2 int    `json:"f2"`
//Field3 Dec    `json:"f3"`
//}

//func TestEmbeddedStructSerializationGoWire(t *testing.T) {
//obj := testEmbedStruct{"foo", 10, NewDec(1, 3)}
//bz, err := cdc.MarshalJSON(obj)
//require.Nil(t, err)

//var obj2 testEmbedStruct
//err = cdc.UnmarshalJSON(bz, &obj2)
//require.Nil(t, err)

//require.Equal(t, obj.Field1, obj2.Field1)
//require.Equal(t, obj.Field2, obj2.Field2)
//require.True(t, obj.Field3.Equal(obj2.Field3), "original: %v, unmarshalled: %v", obj, obj2)
//}

//func TestDecsEqual(t *testing.T) {
//tests := []struct {
//d1s, d2s []Dec
//eq       bool
//}{
//{[]Dec{NewDec(0)}, []Dec{NewDec(0)}, true},
//{[]Dec{NewDec(0)}, []Dec{NewDec(1)}, false},
//{[]Dec{NewDec(0)}, []Dec{}, false},
//{[]Dec{NewDec(0), NewDec(1)}, []Dec{NewDec(0), NewDec(1)}, true},
//{[]Dec{NewDec(1), NewDec(0)}, []Dec{NewDec(1), NewDec(0)}, true},
//{[]Dec{NewDec(1), NewDec(0)}, []Dec{NewDec(0), NewDec(1)}, false},
//{[]Dec{NewDec(1), NewDec(0)}, []Dec{NewDec(1)}, false},
//{[]Dec{NewDec(1), NewDec(2)}, []Dec{NewDec(2), NewDec(4)}, false},
//{[]Dec{NewDec(3), NewDec(18)}, []Dec{NewDec(1), NewDec(6)}, false},
//}

//for tcIndex, tc := range tests {
//require.Equal(t, tc.eq, DecsEqual(tc.d1s, tc.d2s), "equality of decional arrays is incorrect, tc #%d", tcIndex)
//require.Equal(t, tc.eq, DecsEqual(tc.d2s, tc.d1s), "equality of decional arrays is incorrect (converse), tc #%d", tcIndex)
//}

//}

//func TestStringOverflow(t *testing.T) {
//// two random 64 bit primes
//dec1 := NewDec(5164315003622678713, 4389711697696177267)
//dec2 := NewDec(-3179849666053572961, 8459429845579852627)
//dec3 := dec1.Add(dec2)
//require.Equal(t,
//"29728537197630860939575850336935951464/37134458148982045574552091851127630409",
//dec3.String(),
//)
//}
