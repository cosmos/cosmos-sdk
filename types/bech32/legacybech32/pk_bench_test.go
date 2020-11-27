package legacybech32

import (
	"math/rand"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/stretchr/testify/require"
)

func BenchmarkMarshalPubKey(b *testing.B) {
	pkBz := make([]byte, ed25519.PubKeySize)
	pk := &ed25519.PubKey{Key: pkBz}
	rng := rand.New(rand.NewSource(time.Now().Unix()))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		rng.Read(pk.Key)
		b.StartTimer()

		_, err := MarshalPubKey(ConsPub, pk)
		require.NoError(b, err)
	}
}

func BenchmarkGetPubKeyFromBech32(b *testing.B) {
	pkBz := make([]byte, ed25519.PubKeySize)
	pk := &ed25519.PubKey{Key: pkBz}
	rng := rand.New(rand.NewSource(time.Now().Unix()))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		rng.Read(pk.Key)

		pkStr, err := MarshalPubKey(ConsPub, pk)
		require.NoError(b, err)

		b.StartTimer()
		pk2, err := UnmarshalPubKey(ConsPub, pkStr)
		require.NoError(b, err)
		require.Equal(b, pk, pk2)
	}
}
