package tmhash_test

import (
	"crypto/sha256"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tendermint/go-crypto/tmhash"
)

func TestHash(t *testing.T) {
	testVector := []byte("abc")
	hasher := tmhash.New()
	hasher.Write(testVector)
	bz := hasher.Sum(nil)

	hasher = sha256.New()
	hasher.Write(testVector)
	bz2 := hasher.Sum(nil)
	bz2 = bz2[:20]

	assert.Equal(t, bz, bz2)
}
