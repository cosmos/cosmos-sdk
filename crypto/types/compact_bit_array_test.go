package types

import (
	"encoding/json"
	"fmt"
	"math"
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
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			ba.NumTrueBitsBefore(90)
		}
	})
}

func TestNewCompactBitArrayCrashWithLimits(t *testing.T) {
	if testing.Short() {
		t.Skip("This test can be expensive in memory")
	}
	tests := []struct {
		in       int
		mustPass bool
	}{
		{int(^uint(0) >> 30), false},
		{int(^uint(0) >> 1), false},
		{int(^uint(0) >> 2), false},
		{int(math.MaxInt32), true},
		{int(math.MaxInt32) + 1, true},
		{int(math.MaxInt32) + 2, true},
		{int(math.MaxInt32) - 7, true},
		{int(math.MaxInt32) + 24, true},
		{int(math.MaxInt32) * 9, false}, // results in >=maxint after (bits+7)/8
		{1, true},
		{0, false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("%d", tt.in), func(t *testing.T) {
			got := NewCompactBitArray(tt.in)
			if g := got != nil; g != tt.mustPass {
				t.Fatalf("got!=nil=%t, want=%t", g, tt.mustPass)
			}
		})
	}
}
