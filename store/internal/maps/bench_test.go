package maps

import "testing"

func BenchmarkKVPairBytes(b *testing.B) {
	kvp := NewKVPair(make([]byte, 128), make([]byte, 1e6))
	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		b.SetBytes(int64(len(kvp.Bytes())))
	}
}
