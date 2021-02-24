package address

import (
	"crypto/sha256"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestAddressSuite(t *testing.T) {
	suite.Run(t, new(AddressSuite))
}

type AddressSuite struct{ suite.Suite }

func (suite *AddressSuite) TestHash() {
	assert := suite.Assert()
	typ := "1"
	key := []byte{1}
	part1 := sha256.Sum256([]byte(typ))
	expected := sha256.Sum256(append(part1[:], key...))
	received := Hash(typ, key)
	assert.Equal(expected[:], received, "must create a correct address")

	received = Hash("other", key)
	assert.NotEqual(expected[:], received, "must create a correct address")
	assert.Len(received, Len, "must have correct length")
}

func (suite *AddressSuite) TestComposed() {
	assert := suite.Assert()
	a1 := addrMock{[]byte{11, 12}}
	a2 := addrMock{[]byte{21, 22}}

	typ := "multisig"
	ac, err := NewComposed(typ, []Addressable{a1, a2})
	assert.NoError(err)
	assert.Len(ac, Len)

	// check if optimizations work
	checkingKey := append([]byte{}, a1.AddressWithLen(suite.T())...)
	checkingKey = append(checkingKey, a2.AddressWithLen(suite.T())...)
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

func (suite *AddressSuite) TestModule() {
	assert := suite.Assert()
	var modName, key = "myModule", []byte{1, 2}
	addr := Module(modName, key)
	assert.Len(addr, Len, "must have address length")

	addr2 := Module("myModule2", key)
	assert.NotEqual(addr, addr2, "changing module name must change address")

	addr3 := Module(modName, []byte{1, 2, 3})
	assert.NotEqual(addr, addr3, "changing key must change address")
	assert.NotEqual(addr2, addr3, "changing key must change address")
}

func unsafeConvertABC() []byte {
	return unsafeStrToByteArray("abc")
}

func (suite *AddressSuite) TestUnsafeStrToBytes() {
	// we convert in other function to trigger GC
	for i := 0; i < 5; i++ {
		b := unsafeConvertABC()
		<-time.NewTimer(5 * time.Millisecond).C
		b2 := append(b, 'd')
		suite.Equal("abc", string(b))
		suite.Equal("abcd", string(b2))
	}
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
