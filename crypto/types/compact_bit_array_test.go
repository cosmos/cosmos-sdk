package types

import (
	"encoding/json"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	tmrand "github.com/tendermint/tendermint/libs/rand"
)

func randCompactBitArray(bits int) (*CompactBitArray, []byte) {
	numBytes := (bits + 7) / 8
	src := tmrand.Bytes((bits + 7) / 8)
	bA := NewCompactBitArray(bits)

	for i := 0; i < numBytes-1; i++ {
		for j := uint8(0); j < 8; j++ {
			bA.SetIndex(i*8+int(j), src[i]&(uint8(1)<<(8-j)) > 0)
		}
	}
	// Set remaining bits
	for i := uint32(0); i < 8-bA.ExtraBitsStored; i++ {
		bA.SetIndex(numBytes*8+int(i), src[numBytes-1]&(uint8(1)<<(8-i)) > 0)
	}
	return bA, src
}

func TestNewBitArrayNeverCrashesOnNegatives(t *testing.T) {
	bitList := []int{-127, -128, -1 << 31}
	for _, bits := range bitList {
		bA := NewCompactBitArray(bits)
		require.Nil(t, bA)
	}
}

func TestBitArrayEqual(t *testing.T) {
	empty := new(CompactBitArray)
	big1, _ := randCompactBitArray(1000)
	big1Cpy := *big1
	big2, _ := randCompactBitArray(1000)
	big2.SetIndex(500, !big1.GetIndex(500)) // ensure they are different
	cases := []struct {
		name string
		b1   *CompactBitArray
		b2   *CompactBitArray
		eq   bool
	}{
		{name: "both nil are equal", b1: nil, b2: nil, eq: true},
		{name: "if one is nil then not equal", b1: nil, b2: empty, eq: false},
		{name: "nil and empty not equal", b1: empty, b2: nil, eq: false},
		{name: "empty and empty equal", b1: empty, b2: new(CompactBitArray), eq: true},
		{name: "same bits should be equal", b1: big1, b2: &big1Cpy, eq: true},
		{name: "different should not be equal", b1: big1, b2: big2, eq: false},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			eq := tc.b1.Equal(tc.b2)
			require.Equal(t, tc.eq, eq)
		})
	}
}

func TestJSONMarshalUnmarshal(t *testing.T) {

	bA1 := NewCompactBitArray(0)
	bA2 := NewCompactBitArray(1)

	bA3 := NewCompactBitArray(1)
	bA3.SetIndex(0, true)

	bA4 := NewCompactBitArray(5)
	bA4.SetIndex(0, true)
	bA4.SetIndex(1, true)

	bA5 := NewCompactBitArray(9)
	bA5.SetIndex(0, true)
	bA5.SetIndex(1, true)
	bA5.SetIndex(8, true)

	bA6 := NewCompactBitArray(16)
	bA6.SetIndex(0, true)
	bA6.SetIndex(1, true)
	bA6.SetIndex(8, false)
	bA6.SetIndex(15, true)

	testCases := []struct {
		bA           *CompactBitArray
		marshalledBA string
	}{
		{nil, `null`},
		{bA1, `null`},
		{bA2, `"_"`},
		{bA3, `"x"`},
		{bA4, `"xx___"`},
		{bA5, `"xx______x"`},
		{bA6, `"xx_____________x"`},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.bA.String(), func(t *testing.T) {
			bz, err := json.Marshal(tc.bA)
			require.NoError(t, err)

			assert.Equal(t, tc.marshalledBA, string(bz))

			var unmarshalledBA *CompactBitArray
			err = json.Unmarshal(bz, &unmarshalledBA)
			require.NoError(t, err)

			if tc.bA == nil {
				require.Nil(t, unmarshalledBA)
			} else {
				require.NotNil(t, unmarshalledBA)
				assert.EqualValues(t, tc.bA.Elems, unmarshalledBA.Elems)
				if assert.EqualValues(t, tc.bA.String(), unmarshalledBA.String()) {
					assert.EqualValues(t, tc.bA.Elems, unmarshalledBA.Elems)
				}
			}
		})
	}
}

func TestCompactMarshalUnmarshal(t *testing.T) {
	bA1 := NewCompactBitArray(0)
	bA2 := NewCompactBitArray(1)

	bA3 := NewCompactBitArray(1)
	bA3.SetIndex(0, true)

	bA4 := NewCompactBitArray(5)
	bA4.SetIndex(0, true)
	bA4.SetIndex(1, true)

	bA5 := NewCompactBitArray(9)
	bA5.SetIndex(0, true)
	bA5.SetIndex(1, true)
	bA5.SetIndex(8, true)

	bA6 := NewCompactBitArray(16)
	bA6.SetIndex(0, true)
	bA6.SetIndex(1, true)
	bA6.SetIndex(8, false)
	bA6.SetIndex(15, true)

	testCases := []struct {
		bA           *CompactBitArray
		marshalledBA []byte
	}{
		{nil, []byte("null")},
		{bA1, []byte("null")},
		{bA2, []byte{byte(1), byte(0)}},
		{bA3, []byte{byte(1), byte(128)}},
		{bA4, []byte{byte(5), byte(192)}},
		{bA5, []byte{byte(9), byte(192), byte(128)}},
		{bA6, []byte{byte(16), byte(192), byte(1)}},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.bA.String(), func(t *testing.T) {
			bz := tc.bA.CompactMarshal()

			assert.Equal(t, tc.marshalledBA, bz)

			unmarshalledBA, err := CompactUnmarshal(bz)
			require.NoError(t, err)
			if tc.bA == nil {
				require.Nil(t, unmarshalledBA)
			} else {
				require.NotNil(t, unmarshalledBA)
				assert.EqualValues(t, tc.bA.Elems, unmarshalledBA.Elems)
				if assert.EqualValues(t, tc.bA.String(), unmarshalledBA.String()) {
					assert.EqualValues(t, tc.bA.Elems, unmarshalledBA.Elems)
				}
			}
		})
	}
}

func TestCompactBitArrayNumOfTrueBitsBefore(t *testing.T) {
	testCases := []struct {
		marshalledBA   string
		bAIndex        []int
		trueValueIndex []int
	}{
		{`"_____"`, []int{0, 1, 2, 3, 4}, []int{0, 0, 0, 0, 0}},
		{`"x"`, []int{0}, []int{0}},
		{`"_x"`, []int{1}, []int{0}},
		{`"x___xxxx"`, []int{0, 4, 5, 6, 7}, []int{0, 1, 2, 3, 4}},
		{`"__x_xx_x__x_x___"`, []int{2, 4, 5, 7, 10, 12}, []int{0, 1, 2, 3, 4, 5}},
		{`"______________xx"`, []int{14, 15}, []int{0, 1}},
	}
	for tcIndex, tc := range testCases {
		tc := tc
		tcIndex := tcIndex
		t.Run(tc.marshalledBA, func(t *testing.T) {
			var bA *CompactBitArray
			err := json.Unmarshal([]byte(tc.marshalledBA), &bA)
			require.NoError(t, err)

			for i := 0; i < len(tc.bAIndex); i++ {

				require.Equal(t, tc.trueValueIndex[i], bA.NumTrueBitsBefore(tc.bAIndex[i]), "tc %d, i %d", tcIndex, i)
			}
		})
	}
}

func TestCompactBitArrayGetSetIndex(t *testing.T) {
	r := rand.New(rand.NewSource(100))
	numTests := 10
	numBitsPerArr := 100
	for i := 0; i < numTests; i++ {
		bits := r.Intn(1000)
		bA, _ := randCompactBitArray(bits)

		for j := 0; j < numBitsPerArr; j++ {
			copy := bA.Copy()
			index := r.Intn(bits)
			val := (r.Int63() % 2) == 0
			bA.SetIndex(index, val)
			require.Equal(t, val, bA.GetIndex(index), "bA.SetIndex(%d, %v) failed on bit array: %s", index, val, copy)
		}
	}
}

func BenchmarkNumTrueBitsBefore(b *testing.B) {
	ba, _ := randCompactBitArray(100)

	b.Run("new", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ba.NumTrueBitsBefore(90)
		}
	})
}
