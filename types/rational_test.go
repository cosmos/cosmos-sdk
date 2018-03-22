package types

import (
	"math/big"
	"testing"

	wire "github.com/cosmos/cosmos-sdk/wire"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	assert.Equal(t, NewRat(1), NewRat(1, 1))
	assert.Equal(t, NewRat(100), NewRat(100, 1))
	assert.Equal(t, NewRat(-1), NewRat(-1, 1))
	assert.Equal(t, NewRat(-100), NewRat(-100, 1))
	assert.Equal(t, NewRat(0), NewRat(0, 1))

	// do not allow for more than 2 variables
	assert.Panics(t, func() { NewRat(1, 1, 1) })
}

func TestNewFromDecimal(t *testing.T) {
	tests := []struct {
		decimalStr string
		expErr     bool
		exp        Rat
	}{
		{"0", false, NewRat(0)},
		{"1", false, NewRat(1)},
		{"1.1", false, NewRat(11, 10)},
		{"0.75", false, NewRat(3, 4)},
		{"0.8", false, NewRat(4, 5)},
		{"0.11111", false, NewRat(11111, 100000)},
		{".", true, Rat{}},
		{".0", true, Rat{}},
		{"1.", true, Rat{}},
		{"foobar", true, Rat{}},
		{"0.foobar", true, Rat{}},
		{"0.foobar.", true, Rat{}},
	}

	for _, tc := range tests {

		res, err := NewRatFromDecimal(tc.decimalStr)
		if tc.expErr {
			assert.NotNil(t, err, tc.decimalStr)
		} else {
			assert.Nil(t, err)
			assert.True(t, res.Equal(tc.exp))
		}

		// negative tc
		res, err = NewRatFromDecimal("-" + tc.decimalStr)
		if tc.expErr {
			assert.NotNil(t, err, tc.decimalStr)
		} else {
			assert.Nil(t, err)
			assert.True(t, res.Equal(tc.exp.Mul(NewRat(-1))))
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

	for _, tc := range tests {
		assert.Equal(t, tc.gt, tc.r1.GT(tc.r2))
		assert.Equal(t, tc.lt, tc.r1.LT(tc.r2))
		assert.Equal(t, tc.eq, tc.r1.Equal(tc.r2))
	}

}

func TestArithmatic(t *testing.T) {
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

	for _, tc := range tests {
		assert.True(t, tc.resMul.Equal(tc.r1.Mul(tc.r2)), "r1 %v, r2 %v", tc.r1.GetRat(), tc.r2.GetRat())
		assert.True(t, tc.resAdd.Equal(tc.r1.Add(tc.r2)), "r1 %v, r2 %v", tc.r1.GetRat(), tc.r2.GetRat())
		assert.True(t, tc.resSub.Equal(tc.r1.Sub(tc.r2)), "r1 %v, r2 %v", tc.r1.GetRat(), tc.r2.GetRat())

		if tc.r2.Num == 0 { // panic for divide by zero
			assert.Panics(t, func() { tc.r1.Quo(tc.r2) })
		} else {
			assert.True(t, tc.resDiv.Equal(tc.r1.Quo(tc.r2)), "r1 %v, r2 %v", tc.r1.GetRat(), tc.r2.GetRat())
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

	for _, tc := range tests {
		assert.Equal(t, tc.res, tc.r1.Evaluate(), "%v", tc.r1)
		assert.Equal(t, tc.res*-1, tc.r1.Mul(NewRat(-1)).Evaluate(), "%v", tc.r1.Mul(NewRat(-1)))
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
		{ToRat(new(big.Rat).SetFrac(big3, big7)), NewRat(429, 1000), 1000},
		{ToRat(new(big.Rat).SetFrac(big3, big7)), ToRat(big.NewRat(4285714286, 10000000000)), 10000000000},
		{NewRat(1, 2), NewRat(1, 2), 1000},
	}

	for _, tc := range tests {
		assert.Equal(t, tc.res, tc.r.Round(tc.precFactor), "%v", tc.r)
		negR1, negRes := tc.r.Mul(NewRat(-1)), tc.res.Mul(NewRat(-1))
		assert.Equal(t, negRes, negR1.Round(tc.precFactor), "%v", negR1)
	}
}

//func TestZeroSerializationJSON(t *testing.T) {
//r := NewRat(0, 1)
//err := r.UnmarshalJSON([]byte(`"0/1"`))
//assert.Nil(t, err)
//err = r.UnmarshalJSON([]byte(`"0/0"`))
//assert.NotNil(t, err)
//err = r.UnmarshalJSON([]byte(`"1/0"`))
//assert.NotNil(t, err)
//err = r.UnmarshalJSON([]byte(`"{}"`))
//assert.NotNil(t, err)
//}

//func TestSerializationJSON(t *testing.T) {
//r := NewRat(1, 3)

//bz, err := r.MarshalText()
//require.Nil(t, err)

//r2 := NewRat(0, 1)
//err = r2.UnmarshalText(bz)
//require.Nil(t, err)

//assert.True(t, r.Equal(r2), "original: %v, unmarshalled: %v", r, r2)
//}

var cdc = wire.NewCodec() //var jsonCdc JSONCodec // TODO wire.Codec

func TestSerializationGoWire(t *testing.T) {
	r := NewRat(1, 3)

	bz, err := cdc.MarshalBinary(r)
	require.Nil(t, err)

	//str, err := r.MarshalJSON()
	//require.Nil(t, err)

	r2 := NewRat(0, 1)
	err = cdc.UnmarshalBinary([]byte(bz), &r2)
	//panic(fmt.Sprintf("debug bz: %v\n", string(bz)))
	require.Nil(t, err)

	assert.True(t, r.Equal(r2), "original: %v, unmarshalled: %v", r, r2)
}

type testEmbedStruct struct {
	Field1 string `json:"f1"`
	Field2 int    `json:"f2"`
	Field3 Rat    `json:"f3"`
}

func TestEmbeddedStructSerializationGoWire(t *testing.T) {
	obj := testEmbedStruct{"foo", 10, NewRat(1, 3)}

	bz, err := cdc.MarshalBinary(obj)
	require.Nil(t, err)

	var obj2 testEmbedStruct
	obj2.Field3 = NewRat(0, 1) // ... needs to be initialized
	err = cdc.UnmarshalBinary(bz, &obj2)
	require.Nil(t, err)

	assert.Equal(t, obj.Field1, obj2.Field1)
	assert.Equal(t, obj.Field2, obj2.Field2)
	assert.True(t, obj.Field3.Equal(obj2.Field3), "original: %v, unmarshalled: %v", obj, obj2)

}
