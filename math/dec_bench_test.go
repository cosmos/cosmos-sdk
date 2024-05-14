package math

import (
	"testing"
)

func BenchmarkCompareLegacyDecAndNewDec(b *testing.B) {
	specs := map[string]struct {
		dividend, divisor string
	}{
		"small/ small": {
			dividend: "100", divisor: "5",
		},
		"big18/ small": {
			dividend: "999999999999999999", divisor: "10",
		},
		"big18/ big18": {
			dividend: "999999999999999999", divisor: "999999999999999999",
		},
		"small/ big18": {
			dividend: "100", divisor: "999999999999999999",
		},
		"big34/big34": {
			dividend: "9999999999999999999999999999999999", divisor: "1999999999999999999999999999999999",
		},
		"negative big34": {
			dividend: "-9999999999999999999999999999999999", divisor: "999999999999999999999999999",
		},
		"decimal small": {
			dividend: "0.0000000001", divisor: "10",
		},
	}
	for name, spec := range specs {
		b.Run(name, func(b *testing.B) {
			b.Run("LegacyDec", func(b *testing.B) {
				dv, ds := LegacyMustNewDecFromStr(spec.dividend), LegacyMustNewDecFromStr(spec.divisor)
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					_ = dv.Quo(ds)
				}
			})

			b.Run("NewDec", func(b *testing.B) {
				dv, ds := must(NewDecFromString(spec.dividend)), must(NewDecFromString(spec.divisor))
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					_, _ = dv.Quo(ds)
				}
			})
		})
	}
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