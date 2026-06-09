package mldsa65_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/internal/benchmarking"
	"github.com/cosmos/cosmos-sdk/crypto/keys/mldsa65"
)

func BenchmarkSigning(b *testing.B) {
	b.ReportAllocs()
	priv, err := mldsa65.GenPrivKey()
	if err != nil {
		b.Fatal(err)
	}
	benchmarking.BenchmarkSigning(b, &priv)
}

func BenchmarkVerification(b *testing.B) {
	b.ReportAllocs()
	priv, err := mldsa65.GenPrivKey()
	if err != nil {
		b.Fatal(err)
	}
	benchmarking.BenchmarkVerification(b, &priv)
}
