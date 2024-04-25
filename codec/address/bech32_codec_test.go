package address

import (
	"encoding/binary"
	"testing"
	"time"

	"github.com/hashicorp/golang-lru/simplelru"
	"gotest.tools/v3/assert"

	"cosmossdk.io/core/address"

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

func TestBech32CodecRace(t *testing.T) {
	ac := NewBech32Codec("cosmos")

	workers := 4
	done := make(chan bool, workers)
	cancel := make(chan bool)

	for i := byte(1); i <= 2; i++ { // works which will loop in first 100 addresses
		go bytesToStringCaller(t, ac, i, 100, cancel, done)
	}

	for i := byte(1); i <= 2; i++ { // works which will generate 1e6 new addresses
		go bytesToStringCaller(t, ac, i, 1000000, cancel, done)
	}

	<-time.After(time.Millisecond * 30)
	close(cancel)

	// cleanup
	for i := 0; i < 4; i++ {
		<-done
	}
}

// generates AccAddress calling BytesToString
func bytesToStringCaller(t *testing.T, ac address.Codec, prefix byte, max uint32, cancel chan bool, done chan<- bool) {
	t.Helper()

	bz := make([]byte, 5) // prefix + 4 bytes for uint
	bz[0] = prefix
	for i := uint32(0); ; i++ {
		if i >= max {
			i = 0
		}
		select {
		case <-cancel:
			done <- true
			return
		default:
			binary.BigEndian.PutUint32(bz[1:], i)
			str, err := ac.BytesToString(bz)
			assert.NilError(t, err)
			assert.Assert(t, str != "")
		}

	}
}
