package address

import (
	"crypto/rand"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/hashicorp/golang-lru/simplelru"
	"github.com/stretchr/testify/assert"

	"github.com/cosmos/cosmos-sdk/internal/conv"
)

func generateAddresses(totalAddresses int) ([][]byte, error) {
	keys := make([][]byte, totalAddresses)
	addr := make([]byte, 32)
	for i := 0; i < totalAddresses; i++ {
		_, err := rand.Read(addr)
		if err != nil {
			return nil, err
		}
		keys[i] = addr
	}

	return keys, nil
}

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
			assert.True(t, ok)
			assert.Equal(t, cached.cache, tt.lru)

			addr, err := ac.StringToBytes(tt.address)
			assert.NoError(t, err)
			assert.Equal(t, tt.lru.Len(), 1)

			cachedAddr, ok := tt.lru.Get(tt.address)
			assert.True(t, ok)
			assert.Equal(t, addr, cachedAddr)

			accAddr, err := ac.BytesToString(addr)
			assert.NoError(t, err)
			assert.Equal(t, tt.lru.Len(), 2)

			cachedStrAddr, ok := tt.lru.Get(cached.codec.Bech32Prefix + conv.UnsafeBytesToStr(addr))
			assert.True(t, ok)
			assert.Equal(t, accAddr, cachedStrAddr)
		})
	}
}

func TestMultipleBech32Codec(t *testing.T) {
	cosmosAc, ok := NewBech32Codec("cosmos").(cachedBech32Codec)
	assert.True(t, ok)
	stakeAc := NewBech32Codec("stake").(cachedBech32Codec)
	assert.True(t, ok)
	assert.Equal(t, cosmosAc.cache, stakeAc.cache)

	addr := make([]byte, 32)
	_, err := rand.Read(addr)
	assert.NoError(t, err)

	cosmosAddr, err := cosmosAc.BytesToString(addr)
	assert.NoError(t, err)
	stakeAddr, err := stakeAc.BytesToString(addr)
	assert.NoError(t, err)
	assert.True(t, cosmosAddr != stakeAddr)

	cachedCosmosAddr, err := cosmosAc.BytesToString(addr)
	assert.NoError(t, err)
	assert.Equal(t, cosmosAddr, cachedCosmosAddr)

	cachedStakeAddr, err := stakeAc.BytesToString(addr)
	assert.NoError(t, err)
	assert.Equal(t, stakeAddr, cachedStakeAddr)
}

func TestBech32CodecRace(t *testing.T) {
	ac := NewBech32Codec("cosmos")
	myAddrBz := []byte{0x1, 0x2, 0x3, 0x4, 0x5}

	var (
		wgStart, wgDone sync.WaitGroup
		errCount        atomic.Uint32
	)
	const n = 3
	wgStart.Add(n)
	wgDone.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			wgStart.Done()
			wgStart.Wait() // wait for all routines started

			got, err := ac.BytesToString(myAddrBz)
			if err != nil || got != "cosmos1qypqxpq9dc9msf" {
				errCount.Add(1)
			}
			wgDone.Done()
		}()
	}
	wgDone.Wait() // wait for all routines completed
	assert.Equal(t, errCount.Load(), uint32(0))
}
