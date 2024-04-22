package address

import (
	"testing"

	"github.com/hashicorp/golang-lru/simplelru"
	"gotest.tools/v3/assert"

	"github.com/cosmos/cosmos-sdk/internal/conv"
)

func TestNewBech32Codec(t *testing.T) {
	tests := []struct {
		name    string
		prefix  string
		lru     *simplelru.LRU
		address string
	}{
		{
			name:    "create accounts cached bech32 codec",
			prefix:  "cosmos",
			lru:     accAddrCache,
			address: "cosmos1p8s0p6gqc6c9gt77lgr2qqujz49huhu6a80smx",
		},
		{
			name:    "create validator cached bech32 codec",
			prefix:  "cosmosvaloper",
			lru:     valAddrCache,
			address: "cosmosvaloper1sjllsnramtg3ewxqwwrwjxfgc4n4ef9u2lcnj0",
		},
		{
			name:    "create consensus cached bech32 codec",
			prefix:  "cosmosvalcons",
			lru:     consAddrCache,
			address: "cosmosvalcons1ntk8eualewuprz0gamh8hnvcem2nrcdsgz563h",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.lru.Len(), 0)
			ac := NewBech32Codec(tt.prefix)
			cached, ok := ac.(cachedBech32Codec)
			assert.Assert(t, ok)
			assert.Equal(t, cached.cache, tt.lru)

			addr, err := ac.StringToBytes(tt.address)
			assert.NilError(t, err)
			assert.Equal(t, tt.lru.Len(), 1)

			cachedAddr, ok := tt.lru.Get(tt.address)
			assert.Assert(t, ok)
			assert.DeepEqual(t, addr, cachedAddr)

			accAddr, err := ac.BytesToString(addr)
			assert.NilError(t, err)
			assert.Equal(t, tt.lru.Len(), 2)

			cachedStrAddr, ok := tt.lru.Get(conv.UnsafeBytesToStr(addr))
			assert.Assert(t, ok)
			assert.DeepEqual(t, accAddr, cachedStrAddr)
		})
	}
}
