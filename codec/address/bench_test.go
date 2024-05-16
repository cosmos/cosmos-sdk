package address

import (
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/stretchr/testify/assert"

	"cosmossdk.io/core/address"
)

func BenchmarkCodecWithCache(b *testing.B) {
	cdc := NewBech32Codec("cosmos")
	bytesToString(b, cdc)
}

func BenchmarkCodecWithoutCache(b *testing.B) {
	cdc := Bech32Codec{Bech32Prefix: "cosmos"}
	bytesToString(b, cdc)
}

func bytesToString(b *testing.B, cdc address.Codec) {
	b.Helper()
	addresses, err := generateAddresses(10)
	require.NoError(b, err)

	b.Helper()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := cdc.BytesToString(addresses[i%len(addresses)])
		assert.NoError(b, err)
	}
}
