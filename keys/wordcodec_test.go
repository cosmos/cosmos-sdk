package keys

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLengthCalc(t *testing.T) {
	assert := assert.New(t)

	cases := []struct {
		bytes, words int
		flexible     bool
	}{
		{1, 1, false},
		{2, 2, false},
		// bytes pairs with same word count
		{3, 3, true},
		{4, 3, true},
		{5, 4, false},
		// bytes pairs with same word count
		{10, 8, true},
		{11, 8, true},
		{12, 9, false},
		{13, 10, false},
		{20, 15, false},
		// bytes pairs with same word count
		{21, 16, true},
		{32, 24, true},
	}

	for _, tc := range cases {
		wl := wordlenFromBytes(tc.bytes)
		assert.Equal(tc.words, wl, "%d", tc.bytes)

		bl, flex := bytelenFromWords(tc.words)
		assert.Equal(tc.flexible, flex, "%d", tc.words)
		if !flex {
			assert.Equal(tc.bytes, bl, "%d", tc.words)
		} else {
			// check if it is either tc.bytes or tc.bytes +1
			choices := []int{tc.bytes, tc.bytes + 1}
			assert.Contains(choices, bl, "%d", tc.words)
		}
	}
}

func TestEncodeDecode(t *testing.T) {
	assert, require := assert.New(t), require.New(t)

	codec, err := LoadCodec("english")
	require.Nil(err, "%+v", err)

	cases := [][]byte{
		{7, 8, 9},                         // TODO: 3 words -> 3 or 4 bytes
		{12, 54, 99, 11},                  // TODO: 3 words -> 3 or 4 bytes
		{0, 54, 99, 11},                   // TODO: 3 words -> 3 or 4 bytes, detect leading 0
		{1, 2, 3, 4, 5},                   // normal
		{0, 0, 0, 0, 122, 23, 82, 195},    // leading 0s (8 chars, unclear)
		{0, 0, 0, 0, 5, 22, 123, 55, 22},  // leading 0s (9 chars, clear)
		{22, 44, 55, 1, 13, 0, 0, 0, 0},   // trailing 0s (9 chars, clear)
		{0, 5, 253, 2, 0},                 // leading and trailing zeros
		{255, 196, 172, 234, 192, 255},    // big numbers
		{255, 196, 172, 1, 234, 192, 255}, // big numbers, two length choices
		// others?
	}

	for i, tc := range cases {
		w, err := codec.BytesToWords(tc)
		if assert.Nil(err, "%d: %v", i, err) {
			b, err := codec.WordsToBytes(w)
			if assert.Nil(err, "%d: %v", i, err) {
				assert.Equal(len(tc), len(b))
				assert.Equal(tc, b)
			}
		}
	}
}

func TestCheckInvalidLists(t *testing.T) {
	// assert, require := assert.New(t), require.New(t)

}

func TestCheckTypoDetection(t *testing.T) {

}
