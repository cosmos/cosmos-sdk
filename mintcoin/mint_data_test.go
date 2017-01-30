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
	assert.False(s.IsBanker(addr1))
	assert.False(s.IsBanker(addr2))

	s.AddBanker(addr1)
	assert.True(s.IsBanker(addr1))
	assert.False(s.IsBanker(addr2))

	s.AddBanker(addr2)
	assert.True(s.IsBanker(addr1))
	assert.True(s.IsBanker(addr2))

	// make sure multiple adds don't lead to multiple entries
	s.AddBanker(addr1)
	s.AddBanker(addr1)
	s.RemoveBanker(addr1)
	assert.False(s.IsBanker(addr1))
	assert.True(s.IsBanker(addr2))
}
