package mintcoin

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestState(t *testing.T) {
	assert := assert.New(t)
	addr1 := []byte("foobar")
	addr2 := []byte("biggie")

	s := MintState{}
	assert.False(s.IsIssuer(addr1))
	assert.False(s.IsIssuer(addr2))

	s.AddIssuer(addr1)
	assert.True(s.IsIssuer(addr1))
	assert.False(s.IsIssuer(addr2))

	s.AddIssuer(addr2)
	assert.True(s.IsIssuer(addr1))
	assert.True(s.IsIssuer(addr2))

	// make sure multiple adds don't lead to multiple entries
	s.AddIssuer(addr1)
	s.AddIssuer(addr1)
	s.RemoveIssuer(addr1)
	assert.False(s.IsIssuer(addr1))
	assert.True(s.IsIssuer(addr2))
}
