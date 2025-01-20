package v2

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSeededRandSource(t *testing.T) {
	const (
		seed1              int64 = 1
		firstValFromSeed1  int64 = 0x4d65822107fcfd52
		secondValFromSeed1 int64 = 0x78629a0f5f3f164f
	)
	src := NewSeededRandSource(seed1)
	for _, v := range []int64{firstValFromSeed1, secondValFromSeed1} {
		assert.Equal(t, v, src.Int63())
	}
	assert.Equal(t, seed1, src.GetSeed())
}

func TestByteSource(t *testing.T) {
	const (
		seed1              = 1
		firstValFromSeed1  = 0x4d65822107fcfd52
		secondValFromSeed1 = 0x78629a0f5f3f164f
	)
	specs := map[string]struct {
		fuzzSeed []byte
		exp      []uint64
	}{
		"fallback fuzz takes over": {
			fuzzSeed: []byte{},
			exp:      []uint64{firstValFromSeed1, secondValFromSeed1},
		},
		"fuzzSeeds served first": {
			fuzzSeed: []byte{
				1, 2, 3, 4, 5, 6, 7, 8,
				9, 10, 11, 12, 13, 14, 15, 16,
				17, 18, // incomplete uin64, should be ignored
			},
			exp: []uint64{0x102030405060708, 0x90a0b0c0d0e0f10, firstValFromSeed1},
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			byteSource := NewByteSource(spec.fuzzSeed, seed1)
			for _, v := range spec.exp {
				assert.Equal(t, v, byteSource.Uint64())
			}
		})
	}
}
