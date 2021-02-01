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

func TestComposed(t *testing.T) {
	assert := assert.New(t)
	a1 := addrMock{[]byte{11, 12}}
	a2 := addrMock{[]byte{21, 22}}

	typ := "multisig"
	ac, err := NewComposed(typ, []Addressable{a1, a2})
	assert.NoError(err)
	assert.Len(ac, Len)

	// check if optimizations work
	checkingKey := append([]byte{}, a1.AddressWithLen(t)...)
	checkingKey = append(checkingKey, a2.AddressWithLen(t)...)
	ac2 := Hash(typ, checkingKey)
	assert.Equal(ac, ac2, "NewComposed works correctly")

	// changing order of addresses shouldn't impact a composed address
	ac2, err = NewComposed(typ, []Addressable{a2, a1})
	assert.NoError(err)
	assert.Len(ac2, Len)
	assert.Equal(ac, ac2, "NewComposed is not sensitive for order")

	// changing a type should change composed address
	ac2, err = NewComposed(typ+"other", []Addressable{a2, a1})
	assert.NoError(err)
	assert.NotEqual(ac, ac2, "NewComposed must be sensitive to type")

	// changing order of addresses shouldn't impact a composed address
	ac2, err = NewComposed(typ, []Addressable{a1, addrMock{make([]byte, 300, 300)}})
	assert.Error(err)
	assert.Contains(err.Error(), "should be max 255 bytes, got 300")
}

type addrMock struct {
	Addr []byte
}

func (a addrMock) Address() []byte {
	return a.Addr
}

func (a addrMock) AddressWithLen(t *testing.T) []byte {
	addr, err := LengthPrefix(a.Addr)
	assert.NoError(t, err)
	return addr
}
