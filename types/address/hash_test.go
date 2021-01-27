package address

import (
	"crypto/sha256"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHash(t *testing.T) {
	assert := assert.New(t)

	typ := "1"
	key := []byte{1}
	part1 := sha256.Sum256([]byte(typ))
	expected := sha256.Sum256(append(part1[:], key...))
	received := Hash(typ, key)
	assert.Equal(expected[:], received, "must create a correct address")

	received = Hash("other", key)
	assert.NotEqual(expected[:], received, "must create a correct address")

	assert.Len(received, Len, "must have correcte length")
}
