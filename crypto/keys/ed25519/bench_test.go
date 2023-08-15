package ed25519

import (
	"io"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/internal/benchmarking"
	"github.com/cosmos/cosmos-sdk/crypto/types"
)

func BenchmarkKeyGeneration(b *testing.B) {
	b.ReportAllocs()
	benchmarkKeygenWrapper := func(reader io.Reader) types.PrivKey {
		return genPrivKey(reader)
	}k
	benchmarking.BenchmarkKeyGeneration(b, benchmarkKeygenWrapper)
}

func BenchmarkSigning(b *testing.B) {
	b.ReportAllocs()
	priv := GenPrivKey()
	benchmarking.BenchmarkSigning(b, priv)
}

func BenchmarkVerification(b *testing.B) {
	b.ReportAllocs()
	priv := GenPrivKey()
	benchmarking.BenchmarkVerification(b, priv)
}
