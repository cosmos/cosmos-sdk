package types

import (
	"math/big"
	"math/rand"
	"testing"

	wire "github.com/cosmos/cosmos-sdk/wire"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	require.Equal(t, NewRat(1), NewRat(1, 1))
	require.Equal(t, NewRat(100), NewRat(100, 1))
	require.Equal(t, NewRat(-1), NewRat(-1, 1))
	require.Equal(t, NewRat(-100), NewRat(-100, 1))
	require.Equal(t, NewRat(0), NewRat(0, 1))

	// do not allow for more than 2 variables
	require.Panics(t, func() { NewRat(1, 1, 1) })
}

func TestNewFromDecimal(t *testing.T) {
	largeBigInt, success := new(big.Int).SetString("3109736052979742687701388262607869", 10)
	require.True(t, success)
	tests := []struct {
		decimalStr string
		expErr     bool
		exp        Rat
	}{
		{"", true, Rat{}},
		{"0", false, NewRat(0)},
		{"1", false, NewRat(1)},
		{"1.1", false, NewRat(11, 10)},
		{"0.75", false, NewRat(3, 4)},
		{"0.8", false, NewRat(4, 5)},
		{"0.11111", true, NewRat(1111, 10000)},
		{"628240629832763.5738930323617075341", true, NewRat(3141203149163817869, 5000)},
		{"621947210595948537540277652521.5738930323617075341",
			true, NewRatFromBigInt(largeBigInt, big.NewInt(5000))},
		{"628240629832763.5738", false, NewRat(3141203149163817869, 5000)},
		{"621947210595948537540277652521.5738",
			false, NewRatFromBigInt(largeBigInt, big.NewInt(5000))},
		{".", true, Rat{}},
		{".0", true, Rat{}},
		{"1.", true, Rat{}},
		{"foobar", true, Rat{}},
		{"0.foobar", true, Rat{}},
		{"0.foobar.", true, Rat{}},
	}

	for tcIndex, tc := range tests {
		res, err := NewRatFromDecimal(tc.decimalStr, 4)
		if tc.expErr {
			require.NotNil(t, err, tc.decimalStr, "error expected, tc #%d", tcIndex)
		} else {
			require.Nil(t, err, tc.decimalStr, "unexpected error, tc #%d", tcIndex)
			require.True(t, res.Equal(tc.exp), tc.decimalStr, "equality was incorrect, tc #%d", tcIndex)
		}

		// negative tc
		res, err = NewRatFromDecimal("-"+tc.decimalStr, 4)
		if tc.expErr {
			require.NotNil(t, err, tc.decimalStr, "error expected (negative case), tc #%d", tcIndex)
		} else {
			require.Nil(t, err, tc.decimalStr, "unexpected error (negative case), tc #%d", tcIndex)
			require.True(t, res.Equal(tc.exp.Mul(NewRat(-1))), tc.decimalStr, "equality was incorrect (negative case), tc #%d", tcIndex)
		}
	}
}

func TestEqualities(t *testing.T) {
	tests := []struct {
		r1, r2     Rat
		gt, lt, eq bool
	}{
		{NewRat(0), NewRat(0), false, false, true},
		{NewRat(0, 100), NewRat(0, 10000), false, false, true},
		{NewRat(100), NewRat(100), false, false, true},
		{NewRat(-100), NewRat(-100), false, false, true},
		{NewRat(-100, -1), NewRat(100), false, false, true},
		{NewRat(-1, 1), NewRat(1, -1), false, false, true},
		{NewRat(1, -1), NewRat(-1, 1), false, false, true},
		{NewRat(3, 7), NewRat(3, 7), false, false, true},

		{NewRat(0), NewRat(3, 7), false, true, false},
		{NewRat(0), NewRat(100), false, true, false},
		{NewRat(-1), NewRat(3, 7), false, true, false},
		{NewRat(-1), NewRat(100), false, true, false},
		{NewRat(1, 7), NewRat(100), false, true, false},
		{NewRat(1, 7), NewRat(3, 7), false, true, false},
		{NewRat(-3, 7), NewRat(-1, 7), false, true, false},

		{NewRat(3, 7), NewRat(0), true, false, false},
		{NewRat(100), NewRat(0), true, false, false},
		{NewRat(3, 7), NewRat(-1), true, false, false},
		{NewRat(100), NewRat(-1), true, false, false},
		{NewRat(100), NewRat(1, 7), true, false, false},
		{NewRat(3, 7), NewRat(1, 7), true, false, false},
		{NewRat(-1, 7), NewRat(-3, 7), true, false, false},
	}

	for tcIndex, tc := range tests {
		require.Equal(t, tc.gt, tc.r1.GT(tc.r2), "GT result is incorrect, tc #%d", tcIndex)
		require.Equal(t, tc.lt, tc.r1.LT(tc.r2), "LT result is incorrect, tc #%d", tcIndex)
		require.Equal(t, tc.eq, tc.r1.Equal(tc.r2), "equality result is incorrect, tc #%d", tcIndex)
	}

}

func TestArithmetic(t *testing.T) {
	tests := []struct {
		r1, r2                         Rat
		resMul, resDiv, resAdd, resSub Rat
	}{
		// r1       r2         MUL        DIV        ADD        SUB
		{NewRat(0), NewRat(0), NewRat(0), NewRat(0), NewRat(0), NewRat(0)},
		{NewRat(1), NewRat(0), NewRat(0), NewRat(0), NewRat(1), NewRat(1)},
		{NewRat(0), NewRat(1), NewRat(0), NewRat(0), NewRat(1), NewRat(-1)},
		{NewRat(0), NewRat(-1), NewRat(0), NewRat(0), NewRat(-1), NewRat(1)},
		{NewRat(-1), NewRat(0), NewRat(0), NewRat(0), NewRat(-1), NewRat(-1)},

		{NewRat(1), NewRat(1), NewRat(1), NewRat(1), NewRat(2), NewRat(0)},
		{NewRat(-1), NewRat(-1), NewRat(1), NewRat(1), NewRat(-2), NewRat(0)},
		{NewRat(1), NewRat(-1), NewRat(-1), NewRat(-1), NewRat(0), NewRat(2)},
		{NewRat(-1), NewRat(1), NewRat(-1), NewRat(-1), NewRat(0), NewRat(-2)},

		{NewRat(3), NewRat(7), NewRat(21), NewRat(3, 7), NewRat(10), NewRat(-4)},
		{NewRat(2), NewRat(4), NewRat(8), NewRat(1, 2), NewRat(6), NewRat(-2)},
		{NewRat(100), NewRat(100), NewRat(10000), NewRat(1), NewRat(200), NewRat(0)},

		{NewRat(3, 2), NewRat(3, 2), NewRat(9, 4), NewRat(1), NewRat(3), NewRat(0)},
		{NewRat(3, 7), NewRat(7, 3), NewRat(1), NewRat(9, 49), NewRat(58, 21), NewRat(-40, 21)},
		{NewRat(1, 21), NewRat(11, 5), NewRat(11, 105), NewRat(5, 231), NewRat(236, 105), NewRat(-226, 105)},
		{NewRat(-21), NewRat(3, 7), NewRat(-9), NewRat(-49), NewRat(-144, 7), NewRat(-150, 7)},
		{NewRat(100), NewRat(1, 7), NewRat(100, 7), NewRat(700), NewRat(701, 7), NewRat(699, 7)},
	}

	for tcIndex, tc := range tests {
		require.True(t, tc.resMul.Equal(tc.r1.Mul(tc.r2)), "r1 %v, r2 %v. tc #%d", tc.r1.Rat, tc.r2.Rat, tcIndex)
		require.True(t, tc.resAdd.Equal(tc.r1.Add(tc.r2)), "r1 %v, r2 %v. tc #%d", tc.r1.Rat, tc.r2.Rat, tcIndex)
		require.True(t, tc.resSub.Equal(tc.r1.Sub(tc.r2)), "r1 %v, r2 %v. tc #%d", tc.r1.Rat, tc.r2.Rat, tcIndex)

		if tc.r2.Num().IsZero() { // panic for divide by zero
			require.Panics(t, func() { tc.r1.Quo(tc.r2) })
		} else {
			require.True(t, tc.resDiv.Equal(tc.r1.Quo(tc.r2)), "r1 %v, r2 %v. tc #%d", tc.r1.Rat, tc.r2.Rat, tcIndex)
		}
	}
}

func TestEvaluate(t *testing.T) {
	tests := []struct {
		r1  Rat
		res int64
	}{
		{NewRat(0), 0},
		{NewRat(1), 1},
		{NewRat(1, 4), 0},
		{NewRat(1, 2), 0},
		{NewRat(3, 4), 1},
		{NewRat(5, 6), 1},
		{NewRat(3, 2), 2},
		{NewRat(5, 2), 2},
		{NewRat(6, 11), 1},  // 0.545-> 1 even though 5 is first decimal and 1 not even
		{NewRat(17, 11), 2}, // 1.545
		{NewRat(5, 11), 0},
		{NewRat(16, 11), 1},
		{NewRat(113, 12), 9},
	}

	for tcIndex, tc := range tests {
		require.Equal(t, tc.res, tc.r1.RoundInt64(), "%v. tc #%d", tc.r1, tcIndex)
		require.Equal(t, tc.res*-1, tc.r1.Mul(NewRat(-1)).RoundInt64(), "%v. tc #%d", tc.r1.Mul(NewRat(-1)), tcIndex)
	}
}

func TestRound(t *testing.T) {
	many3 := "333333333333333333333333333333333333333333333"
	many7 := "777777777777777777777777777777777777777777777"
	big3, worked := new(big.Int).SetString(many3, 10)
	require.True(t, worked)
	big7, worked := new(big.Int).SetString(many7, 10)
	require.True(t, worked)

	tests := []struct {
		r, res     Rat
		precFactor int64
	}{
		{NewRat(333, 777), NewRat(429, 1000), 1000},
		{Rat{new(big.Rat).SetFrac(big3, big7)}, NewRat(429, 1000), 1000},
		{Rat{new(big.Rat).SetFrac(big3, big7)}, Rat{big.NewRat(4285714286, 10000000000)}, 10000000000},
		{NewRat(1, 2), NewRat(1, 2), 1000},
	}

	for tcIndex, tc := range tests {
		require.Equal(t, tc.res, tc.r.Round(tc.precFactor), "%v", tc.r, "incorrect rounding, tc #%d", tcIndex)
		negR1, negRes := tc.r.Mul(NewRat(-1)), tc.res.Mul(NewRat(-1))
		require.Equal(t, negRes, negR1.Round(tc.precFactor), "%v", negR1, "incorrect rounding (negative case), tc #%d", tcIndex)
	}
}

func TestToLeftPadded(t *testing.T) {
	tests := []struct {
		rat    Rat
		digits int8
		res    string
	}{
		{NewRat(100, 3), 8, "00000033"},
		{NewRat(1, 3), 8, "00000000"},
		{NewRat(100, 2), 8, "00000050"},
		{NewRat(1000, 3), 8, "00000333"},
		{NewRat(1000, 3), 12, "000000000333"},
	}
	for tcIndex, tc := range tests {
		require.Equal(t, tc.res, tc.rat.ToLeftPadded(tc.digits), "incorrect left padding, tc #%d", tcIndex)
	}
}

var cdc = wire.NewCodec() //var jsonCdc JSONCodec // TODO wire.Codec

func TestZeroSerializationJSON(t *testing.T) {
	r := NewRat(0, 1)
	err := cdc.UnmarshalJSON([]byte(`"0/1"`), &r)
	require.Nil(t, err)
	err = cdc.UnmarshalJSON([]byte(`"0/0"`), &r)
	require.NotNil(t, err)
	err = cdc.UnmarshalJSON([]byte(`"1/0"`), &r)
	require.NotNil(t, err)
	err = cdc.UnmarshalJSON([]byte(`"{}"`), &r)
	require.NotNil(t, err)
}

func TestSerializationText(t *testing.T) {
	r := NewRat(1, 3)

	bz, err := r.MarshalText()
	require.NoError(t, err)

	var r2 = Rat{new(big.Rat)}
	err = r2.UnmarshalText(bz)
	require.NoError(t, err)
	require.True(t, r.Equal(r2), "original: %v, unmarshalled: %v", r, r2)
}

func TestSerializationGoWireJSON(t *testing.T) {
	r := NewRat(1, 3)
	bz, err := cdc.MarshalJSON(r)
	require.NoError(t, err)

	var r2 Rat
	err = cdc.UnmarshalJSON(bz, &r2)
	require.NoError(t, err)
	require.True(t, r.Equal(r2), "original: %v, unmarshalled: %v", r, r2)
}

func TestSerializationGoWireBinary(t *testing.T) {
	r := NewRat(1, 3)
	bz, err := cdc.MarshalBinary(r)
	require.NoError(t, err)

	var r2 Rat
	err = cdc.UnmarshalBinary(bz, &r2)
	require.NoError(t, err)
	require.True(t, r.Equal(r2), "original: %v, unmarshalled: %v", r, r2)
}

type testEmbedStruct struct {
	Field1 string `json:"f1"`
	Field2 int    `json:"f2"`
	Field3 Rat    `json:"f3"`
}

func TestEmbeddedStructSerializationGoWire(t *testing.T) {
	obj := testEmbedStruct{"foo", 10, NewRat(1, 3)}
	bz, err := cdc.MarshalJSON(obj)
	require.Nil(t, err)

	var obj2 testEmbedStruct
	err = cdc.UnmarshalJSON(bz, &obj2)
	require.Nil(t, err)

	require.Equal(t, obj.Field1, obj2.Field1)
	require.Equal(t, obj.Field2, obj2.Field2)
	require.True(t, obj.Field3.Equal(obj2.Field3), "original: %v, unmarshalled: %v", obj, obj2)
}

func TestRatsEqual(t *testing.T) {
	tests := []struct {
		r1s, r2s []Rat
		eq       bool
	}{
		{[]Rat{NewRat(0)}, []Rat{NewRat(0)}, true},
		{[]Rat{NewRat(0)}, []Rat{NewRat(1)}, false},
		{[]Rat{NewRat(0)}, []Rat{}, false},
		{[]Rat{NewRat(0), NewRat(1)}, []Rat{NewRat(0), NewRat(1)}, true},
		{[]Rat{NewRat(1), NewRat(0)}, []Rat{NewRat(1), NewRat(0)}, true},
		{[]Rat{NewRat(1), NewRat(0)}, []Rat{NewRat(0), NewRat(1)}, false},
		{[]Rat{NewRat(1), NewRat(0)}, []Rat{NewRat(1)}, false},
		{[]Rat{NewRat(1), NewRat(2)}, []Rat{NewRat(2), NewRat(4)}, false},
		{[]Rat{NewRat(3), NewRat(18)}, []Rat{NewRat(1), NewRat(6)}, false},
	}

	for tcIndex, tc := range tests {
		require.Equal(t, tc.eq, RatsEqual(tc.r1s, tc.r2s), "equality of rational arrays is incorrect, tc #%d", tcIndex)
		require.Equal(t, tc.eq, RatsEqual(tc.r2s, tc.r1s), "equality of rational arrays is incorrect (converse), tc #%d", tcIndex)
	}

}

func TestStringOverflow(t *testing.T) {
	// two random 64 bit primes
	rat1 := NewRat(5164315003622678713, 4389711697696177267)
	rat2 := NewRat(-3179849666053572961, 8459429845579852627)
	rat3 := rat1.Add(rat2)
	require.Equal(t,
		"29728537197630860939575850336935951464/37134458148982045574552091851127630409",
		rat3.String(),
	)
}

// Tests below uses randomness
// Since we are using *big.Rat as underlying value
// and (U/)Int is immutable value(see TestImmutability(U/)Int)
// it is safe to use randomness in the tests
func TestArithRat(t *testing.T) {
	for i := 0; i < 20; i++ {
		n1 := NewInt(int64(rand.Int31()))
		d1 := NewInt(int64(rand.Int31()))
		rat1 := NewRatFromInt(n1, d1)

		n2 := NewInt(int64(rand.Int31()))
		d2 := NewInt(int64(rand.Int31()))
		rat2 := NewRatFromInt(n2, d2)

		n1d2 := n1.Mul(d2)
		n2d1 := n2.Mul(d1)

		cases := []struct {
			nres Int
			dres Int
			rres Rat
		}{
			{n1d2.Add(n2d1), d1.Mul(d2), rat1.Add(rat2)},
			{n1d2.Sub(n2d1), d1.Mul(d2), rat1.Sub(rat2)},
			{n1.Mul(n2), d1.Mul(d2), rat1.Mul(rat2)},
			{n1d2, n2d1, rat1.Quo(rat2)},
		}

		for _, tc := range cases {
			require.Equal(t, NewRatFromInt(tc.nres, tc.dres), tc.rres)
		}
	}
}

func TestCompRat(t *testing.T) {
	for i := 0; i < 20; i++ {
		n1 := NewInt(int64(rand.Int31()))
		d1 := NewInt(int64(rand.Int31()))
		rat1 := NewRatFromInt(n1, d1)

		n2 := NewInt(int64(rand.Int31()))
		d2 := NewInt(int64(rand.Int31()))
		rat2 := NewRatFromInt(n2, d2)

		n1d2 := n1.Mul(d2)
		n2d1 := n2.Mul(d1)

		cases := []struct {
			ires bool
			rres bool
		}{
			{n1d2.Equal(n2d1), rat1.Equal(rat2)},
			{n1d2.GT(n2d1), rat1.GT(rat2)},
			{n1d2.LT(n2d1), rat1.LT(rat2)},
			{n1d2.GT(n2d1) || n1d2.Equal(n2d1), rat1.GTE(rat2)},
			{n1d2.LT(n2d1) || n1d2.Equal(n2d1), rat1.LTE(rat2)},
		}

		for _, tc := range cases {
			require.Equal(t, tc.ires, tc.rres)
		}
	}
}

func TestImmutabilityRat(t *testing.T) {
	for i := 0; i < 20; i++ {
		n := int64(rand.Int31())
		r := NewRat(n)
		z := ZeroRat()
		o := OneRat()

		r.Add(z)
		r.Sub(z)
		r.Mul(o)
		r.Quo(o)

		require.Equal(t, n, r.RoundInt64())
		require.True(t, NewRat(n).Equal(r))
	}

}
