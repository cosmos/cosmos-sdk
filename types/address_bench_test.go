package types_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/types"
)

func BenchmarkBech32ifyPubKey(b *testing.B) {
	pk := make(ed25519.PubKey, ed25519.PubKeySize)
	rng := rand.New(rand.NewSource(time.Now().Unix()))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		rng.Read(pk)
		b.StartTimer()

		_, err := types.Bech32ifyPubKey(types.Bech32PubKeyTypeConsPub, pk)
		require.NoError(b, err)
	}
}

func BenchmarkGetPubKeyFromBech32(b *testing.B) {
	pk := make(ed25519.PubKey, ed25519.PubKeySize)
	rng := rand.New(rand.NewSource(time.Now().Unix()))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		rng.Read(pk)

		pkStr, err := types.Bech32ifyPubKey(types.Bech32PubKeyTypeConsPub, pk)
		require.NoError(b, err)

		b.StartTimer()
		pk2, err := types.GetPubKeyFromBech32(types.Bech32PubKeyTypeConsPub, pkStr)
		require.NoError(b, err)
		require.Equal(b, pk, pk2)
	}
}
