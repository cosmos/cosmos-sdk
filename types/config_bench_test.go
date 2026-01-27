package types

import "testing"

func BenchmarkGetConfig_DefaultScope(b *testing.B) {
	// Make sure we use the default (cached) scope, not env override.
	b.Setenv(EnvConfigScope, "")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = GetConfig()
	}
}

func BenchmarkGetConfig_EnvScope(b *testing.B) {
	// Benchmark behavior when env override is set.
	b.Setenv(EnvConfigScope, "bench-scope")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = GetConfig()
	}
}
