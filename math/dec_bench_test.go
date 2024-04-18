package math

import (
	"testing"
)

func BenchmarkCompareLegacyDecAndNewDec(b *testing.B) {
	legacyB1 := LegacyNewDec(100)
	legacyB2 := LegacyNewDec(5)
	newB1 := NewDecFromInt64(100)
	newB2 := NewDecFromInt64(5)

	b.Run("LegacyDec", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = legacyB1.Quo(legacyB2)
		}
	})

	b.Run("NewDec", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = newB1.Quo(newB2)
		}
	})
}

func BenchmarkCompareLegacyDecAndNewDecQuoInteger(b *testing.B) {
	legacyB1 := LegacyNewDec(100)
	newB1 := NewDecFromInt64(100)

	b.Run("LegacyDec", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = legacyB1.Quo(LegacyNewDec(1))
		}
	})

	b.Run("NewDec", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = newB1.QuoInteger(NewDecFromInt64(1))
		}
	})
}

func BenchmarkCompareLegacyAddAndDecAdd(b *testing.B) {
	legacyB1 := LegacyNewDec(100)
	legacyB2 := LegacyNewDec(5)
	newB1 := NewDecFromInt64(100)
	newB2 := NewDecFromInt64(5)

	b.Run("LegacyDec", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = legacyB1.Add(legacyB2)
		}
	})

	b.Run("NewDec", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = newB1.Add(newB2)
		}
	})
}

func BenchmarkCompareLegacySubAndDecMul(b *testing.B) {
	legacyB1 := LegacyNewDec(100)
	legacyB2 := LegacyNewDec(5)
	newB1 := NewDecFromInt64(100)
	newB2 := NewDecFromInt64(5)

	b.Run("LegacyDec", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = legacyB1.Mul(legacyB2)
		}
	})

	b.Run("NewDec", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = newB1.Mul(newB2)
		}
	})
}

func BenchmarkCompareLegacySubAndDecSub(b *testing.B) {
	legacyB1 := LegacyNewDec(100)
	legacyB2 := LegacyNewDec(5)
	newB1 := NewDecFromInt64(100)
	newB2 := NewDecFromInt64(5)

	b.Run("LegacyDec", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = legacyB1.Sub(legacyB2)
		}
	})

	b.Run("NewDec", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = newB1.Sub(newB2)
		}
	})
}