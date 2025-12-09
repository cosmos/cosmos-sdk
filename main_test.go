package main

import (
	"testing"

	sdkmath "cosmossdk.io/math"
)

// Benchmark the stress test
func BenchmarkStressTest(b *testing.B) {
	pool := MockOsmosisPool{BTCBalance: sdkmath.ZeroInt()}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tx := generateMockBTCTx(i)
		simulateBitcoinInteroperability(&pool, tx)
	}
}
