package ed25519

import (
	"io"
	"testing"

	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/crypto/benchmark"
)

func BenchmarkKeyGeneration(b *testing.B) {
	benchmarkKeygenWrapper := func(reader io.Reader) crypto.PrivKey {
		return genPrivKey(reader)
	}
	benchmark.BenchmarkKeyGeneration(b, benchmarkKeygenWrapper)
}

func BenchmarkSigning(b *testing.B) {
	priv := GenPrivKey()
	benchmark.BenchmarkSigning(b, priv)
}

func BenchmarkVerification(b *testing.B) {
	priv := GenPrivKey()
	benchmark.BenchmarkVerification(b, priv)
}
