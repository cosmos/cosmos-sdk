package types

import (
	"encoding/json"
	"math/big"
	"testing"

	asrt "github.com/stretchr/testify/assert"
	rqr "github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	assert := asrt.New(t)

	assert.Equal(New(1), New(1, 1))
	assert.Equal(New(100), New(100, 1))
	assert.Equal(New(-1), New(-1, 1))
	assert.Equal(New(-100), New(-100, 1))
	assert.Equal(New(0), New(0, 1))

	// do not allow for more than 2 variables
	assert.Panics(func() { New(1, 1, 1) })
}

func TestNewFromDecimal(t *testing.T) {
	assert := asrt.New(t)

	tests := []struct {
		decimalStr string
		expErr     bool
		exp        Rat
	}{
		{"0", false, New(0)},
		{"1", false, New(1)},
		{"1.1", false, New(11, 10)},
		{"0.75", false, New(3, 4)},
		{"0.8", false, New(4, 5)},
		{"0.11111", false, New(11111, 100000)},
		{".", true, Rat{}},
		{".0", true, Rat{}},
		{"1.", true, Rat{}},
		{"foobar", true, Rat{}},
		{"0.foobar", true, Rat{}},
		{"0.foobar.", true, Rat{}},
	}

	for _, tc := range tests {

		res, err := NewFromDecimal(tc.decimalStr)
		if tc.expErr {
			assert.NotNil(err, tc.decimalStr)
		} else {
			assert.Nil(err)
			assert.True(res.Equal(tc.exp))
		}

		// negative tc
		res, err = NewFromDecimal("-" + tc.decimalStr)
		if tc.expErr {
			assert.NotNil(err, tc.decimalStr)
		} else {
			assert.Nil(err)
			assert.True(res.Equal(tc.exp.Mul(New(-1))))
		}
	}
}

func TestEqualities(t *testing.T) {
	assert := asrt.New(t)

	tests := []struct {
		r1, r2     Rat
		gt, lt, eq bool
	}{
		{New(0), New(0), false, false, true},
		{New(0, 100), New(0, 10000), false, false, true},
		{New(100), New(100), false, false, true},
		{New(-100), New(-100), false, false, true},
		{New(-100, -1), New(100), false, false, true},
		{New(-1, 1), New(1, -1), false, false, true},
		{New(1, -1), New(-1, 1), false, false, true},
		{New(3, 7), New(3, 7), false, false, true},

		{New(0), New(3, 7), false, true, false},
		{New(0), New(100), false, true, false},
		{New(-1), New(3, 7), false, true, false},
		{New(-1), New(100), false, true, false},
		{New(1, 7), New(100), false, true, false},
		{New(1, 7), New(3, 7), false, true, false},
		{New(-3, 7), New(-1, 7), false, true, false},

		{New(3, 7), New(0), true, false, false},
		{New(100), New(0), true, false, false},
		{New(3, 7), New(-1), true, false, false},
		{New(100), New(-1), true, false, false},
		{New(100), New(1, 7), true, false, false},
		{New(3, 7), New(1, 7), true, false, false},
		{New(-1, 7), New(-3, 7), true, false, false},
	}

	for _, tc := range tests {
		assert.Equal(tc.gt, tc.r1.GT(tc.r2))
		assert.Equal(tc.lt, tc.r1.LT(tc.r2))
		assert.Equal(tc.eq, tc.r1.Equal(tc.r2))
	}

}

func TestArithmatic(t *testing.T) {
	assert := asrt.New(t)

	tests := []struct {
		r1, r2                         Rat
		resMul, resDiv, resAdd, resSub Rat
	}{
		// r1    r2      MUL     DIV     ADD     SUB
		{New(0), New(0), New(0), New(0), New(0), New(0)},
		{New(1), New(0), New(0), New(0), New(1), New(1)},
		{New(0), New(1), New(0), New(0), New(1), New(-1)},
		{New(0), New(-1), New(0), New(0), New(-1), New(1)},
		{New(-1), New(0), New(0), New(0), New(-1), New(-1)},

		{New(1), New(1), New(1), New(1), New(2), New(0)},
		{New(-1), New(-1), New(1), New(1), New(-2), New(0)},
		{New(1), New(-1), New(-1), New(-1), New(0), New(2)},
		{New(-1), New(1), New(-1), New(-1), New(0), New(-2)},

		{New(3), New(7), New(21), New(3, 7), New(10), New(-4)},
		{New(2), New(4), New(8), New(1, 2), New(6), New(-2)},
		{New(100), New(100), New(10000), New(1), New(200), New(0)},

		{New(3, 2), New(3, 2), New(9, 4), New(1), New(3), New(0)},
		{New(3, 7), New(7, 3), New(1), New(9, 49), New(58, 21), New(-40, 21)},
		{New(1, 21), New(11, 5), New(11, 105), New(5, 231), New(236, 105), New(-226, 105)},
		{New(-21), New(3, 7), New(-9), New(-49), New(-144, 7), New(-150, 7)},
		{New(100), New(1, 7), New(100, 7), New(700), New(701, 7), New(699, 7)},
	}

	for _, tc := range tests {
		assert.True(tc.resMul.Equal(tc.r1.Mul(tc.r2)), "r1 %v, r2 %v", tc.r1.GetRat(), tc.r2.GetRat())
		assert.True(tc.resAdd.Equal(tc.r1.Add(tc.r2)), "r1 %v, r2 %v", tc.r1.GetRat(), tc.r2.GetRat())
		assert.True(tc.resSub.Equal(tc.r1.Sub(tc.r2)), "r1 %v, r2 %v", tc.r1.GetRat(), tc.r2.GetRat())

		if tc.r2.Num() == 0 { // panic for divide by zero
			assert.Panics(func() { tc.r1.Quo(tc.r2) })
		} else {
			assert.True(tc.resDiv.Equal(tc.r1.Quo(tc.r2)), "r1 %v, r2 %v", tc.r1.GetRat(), tc.r2.GetRat())
		}
	}
}

func TestEvaluate(t *testing.T) {
	assert := asrt.New(t)

	tests := []struct {
		r1  Rat
		res int64
	}{
		{New(0), 0},
		{New(1), 1},
		{New(1, 4), 0},
		{New(1, 2), 0},
		{New(3, 4), 1},
		{New(5, 6), 1},
		{New(3, 2), 2},
		{New(5, 2), 2},
		{New(6, 11), 1},  // 0.545-> 1 even though 5 is first decimal and 1 not even
		{New(17, 11), 2}, // 1.545
		{New(5, 11), 0},
		{New(16, 11), 1},
		{New(113, 12), 9},
	}

	for _, tc := range tests {
		assert.Equal(tc.res, tc.r1.Evaluate(), "%v", tc.r1)
		assert.Equal(tc.res*-1, tc.r1.Mul(New(-1)).Evaluate(), "%v", tc.r1.Mul(New(-1)))
	}
}

func TestRound(t *testing.T) {
	assert, require := asrt.New(t), rqr.New(t)

	many3 := "333333333333333333333333333333333333333333333"
	many7 := "777777777777777777777777777777777777777777777"
	big3, worked := new(big.Int).SetString(many3, 10)
	require.True(worked)
	big7, worked := new(big.Int).SetString(many7, 10)
	require.True(worked)

	tests := []struct {
		r1, res    Rat
		precFactor int64
	}{
		{New(333, 777), New(429, 1000), 1000},
		{Rat{new(big.Rat).SetFrac(big3, big7)}, New(429, 1000), 1000},
		{Rat{new(big.Rat).SetFrac(big3, big7)}, New(4285714286, 10000000000), 10000000000},
		{New(1, 2), New(1, 2), 1000},
	}

	for _, tc := range tests {
		assert.Equal(tc.res, tc.r1.Round(tc.precFactor), "%v", tc.r1)
		negR1, negRes := tc.r1.Mul(New(-1)), tc.res.Mul(New(-1))
		assert.Equal(negRes, negR1.Round(tc.precFactor), "%v", negR1)
	}
}

func TestZeroSerializationJSON(t *testing.T) {
	assert := asrt.New(t)

	var r Rat
	err := json.Unmarshal([]byte("{\"numerator\":0,\"denominator\":1}"), &r)
	assert.Nil(err)
	err = json.Unmarshal([]byte("{\"numerator\":0,\"denominator\":0}"), &r)
	assert.NotNil(err)
	err = json.Unmarshal([]byte("{\"numerator\":1,\"denominator\":0}"), &r)
	assert.NotNil(err)
	err = json.Unmarshal([]byte("{}"), &r)
	assert.NotNil(err)
}

func TestSerializationJSON(t *testing.T) {
	assert, require := asrt.New(t), rqr.New(t)

	r := New(1, 3)

	rMarshal, err := json.Marshal(r)
	require.Nil(err)

	var rUnmarshal Rat
	err = json.Unmarshal(rMarshal, &rUnmarshal)
	require.Nil(err)

	assert.True(r.Equal(rUnmarshal), "original: %v, unmarshalled: %v", r, rUnmarshal)
}

func TestSerializationGoWire(t *testing.T) {
	assert, require := asrt.New(t), rqr.New(t)

	r := New(1, 3)

	rMarshal, err := cdc.MarshalJSON(r)
	require.Nil(err)

	var rUnmarshal Rat
	err = cdc.UnmarshalJSON(rMarshal, &rUnmarshal)
	require.Nil(err)

	assert.True(r.Equal(rUnmarshal), "original: %v, unmarshalled: %v", r, rUnmarshal)
}

type testEmbedStruct struct {
	Field1 string `json:"f1"`
	Field2 int    `json:"f2"`
	Field3 Rat    `json:"f3"`
}

func TestEmbeddedStructSerializationGoWire(t *testing.T) {
	assert, require := asrt.New(t), rqr.New(t)

	r := testEmbedStruct{"foo", 10, New(1, 3)}

	rMarshal, err := cdc.MarshalJSON(r)
	require.Nil(err)

	var rUnmarshal testEmbedStruct
	err = cdc.UnmarshalJSON(rMarshal, &rUnmarshal)
	require.Nil(err)

	assert.Equal(r.Field1, rUnmarshal.Field1)
	assert.Equal(r.Field2, rUnmarshal.Field2)
	assert.True(r.Field3.Equal(rUnmarshal.Field3), "original: %v, unmarshalled: %v", r, rUnmarshal)

}

type testEmbedInterface struct {
	Field1 string   `json:"f1"`
	Field2 int      `json:"f2"`
	Field3 Rational `json:"f3"`
}

func TestEmbeddedInterfaceSerializationGoWire(t *testing.T) {
	assert, require := asrt.New(t), rqr.New(t)

	r := testEmbedInterface{"foo", 10, New(1, 3)}

	rMarshal, err := cdc.MarshalJSON(r)
	require.Nil(err)

	var rUnmarshal testEmbedInterface
	err = cdc.UnmarshalJSON(rMarshal, &rUnmarshal)
	require.Nil(err)

	assert.Equal(r.Field1, rUnmarshal.Field1)
	assert.Equal(r.Field2, rUnmarshal.Field2)
	assert.True(r.Field3.Equal(rUnmarshal.Field3), "original: %v, unmarshalled: %v", r, rUnmarshal)
}
