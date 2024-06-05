package math

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func BenchmarkCompareLegacyDecAndNewDecQuotient(b *testing.B) {
	specs := map[string]struct {
		dividend, divisor string
	}{
		"small/ small": {
			dividend: "100", divisor: "5",
		},
		"big18/ small": {
			dividend: "999999999999999999", divisor: "10",
		},
		"self18/ self18": {
			dividend: "999999999999999999", divisor: "999999999999999999",
		},
		"big18/ big18": {
			dividend: "888888888888888888", divisor: "444444444444444444",
		},
		"decimal18b/ decimal18c": {
			dividend: "8.88888888888888888", divisor: "4.1234567890123",
		},
		"small/ big18": {
			dividend: "100", divisor: "999999999999999999",
		},
		"big34/ big34": {
			dividend: "9999999999999999999999999999999999", divisor: "1999999999999999999999999999999999",
		},
		"negative big34": {
			dividend: "-9999999999999999999999999999999999", divisor: "999999999999999999999999999",
		},
		"decimal small": {
			dividend: "0.0000000001", divisor: "10",
		},
		"decimal small/decimal small ": {
			dividend: "0.0000000001", divisor: "0.0001",
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

func BenchmarkCompareLegacyDecAndNewDecSum(b *testing.B) {
	specs := map[string]struct {
		summands []string
	}{
		"1+2": {
			summands: []string{"1", "2"},
		},
		"small numbers": {
			summands: []string{"123", "0.2", "3.1415", "15"},
		},
		"medium numbers": {
			summands: []string{"1234.567899", "9991345552.2340134"},
		},
		"big18": {
			summands: []string{"123456789012345678", "123456789012345678", "123456789012345678", "123456789012345678", "123456789012345678", "123456789012345678"},
		},

		"growing numbers": {
			summands: []string{"1", "100", "1000", "100000", "10000000", "10000000000", "10000000000000", "100000000000000000"},
		},
		"decimals": {
			summands: []string{"0.1", "0.01", "0.001", "0.000001", "0.00000001", "0.00000000001", "0.00000000000001", "0.000000000000000001"},
		},
	}
	for name, spec := range specs {
		b.Run(name, func(b *testing.B) {
			b.Run("LegacyDec", func(b *testing.B) {
				summands := make([]LegacyDec, len(spec.summands))
				for i, s := range spec.summands {
					summands[i] = LegacyMustNewDecFromStr(s)
				}
				sum := LegacyNewDec(0)
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					for _, s := range summands {
						sum = sum.Add(s)
					}
				}
			})

			b.Run("NewDec", func(b *testing.B) {
				summands := make([]Dec, len(spec.summands))
				for i, s := range spec.summands {
					summands[i] = must(NewDecFromString(s))
				}
				sum := NewDecFromInt64(0)
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					for _, s := range summands {
						sum, _ = sum.Add(s)
					}
				}
			})
		})
	}
}

func BenchmarkCompareLegacyDecAndNewDecSub(b *testing.B) {
	specs := map[string]struct {
		minuend     string
		subtrahends []string
	}{
		"100 - 1 - 2": {
			minuend:     "100",
			subtrahends: []string{"1", "2"},
		},
		"small numbers": {
			minuend:     "152.4013",
			subtrahends: []string{"123", "0.2", "3.1415", "15"},
		},
		"10000000 - big18 numbers": {
			minuend:     "10000000",
			subtrahends: []string{"123456789012345678", "123456789012345678", "123456789012345678", "123456789012345678", "123456789012345678", "123456789012345678"},
		},
		"10000000 - growing numbers": {
			minuend:     "10000000",
			subtrahends: []string{"1", "100", "1000", "100000", "10000000", "10000000000", "10000000000000", "100000000000000000"},
		},
		"10000000 shrinking decimals": {
			minuend:     "10000000",
			subtrahends: []string{"0.1", "0.01", "0.001", "0.000001", "0.00000001", "0.00000000001", "0.00000000000001", "0.000000000000000001"},
		},
	}
	for name, spec := range specs {
		b.Run(name, func(b *testing.B) {
			b.Run("LegacyDec", func(b *testing.B) {
				summands := make([]LegacyDec, len(spec.subtrahends))
				for i, s := range spec.subtrahends {
					summands[i] = LegacyMustNewDecFromStr(s)
				}
				diff := LegacyMustNewDecFromStr(spec.minuend)
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					for _, s := range summands {
						diff = diff.Sub(s)
					}
				}
			})

			b.Run("NewDec", func(b *testing.B) {
				summands := make([]Dec, len(spec.subtrahends))
				for i, s := range spec.subtrahends {
					summands[i] = must(NewDecFromString(s))
				}
				diff := must(NewDecFromString(spec.minuend))
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					for _, s := range summands {
						diff, _ = diff.Sub(s)
					}
				}
			})
		})
	}
}

func BenchmarkCompareLegacyDecAndNewDecMul(b *testing.B) {
	specs := map[string]struct {
		multiplier, multiplicant string
	}{
		"small/ small": {
			multiplier: "100", multiplicant: "5",
		},
		"big18/ small": {
			multiplier: "999999999999999999", multiplicant: "10",
		},
		"self18/ self18": {
			multiplier: "999999999999999999", multiplicant: "999999999999999999",
		},
		"big18/ big18": {
			multiplier: "888888888888888888", multiplicant: "444444444444444444",
		},
		"decimal18b/ decimal18c": {
			multiplier: "8.88888888888888888", multiplicant: "4.1234567890123",
		},
		"small/ big18": {
			multiplier: "100", multiplicant: "999999999999999999",
		},
		"big34/ big34": {
			multiplier: "9999999999999999999999999999999999", multiplicant: "1999999999999999999999999999999999",
		},
		"negative big34": {
			multiplier: "-9999999999999999999999999999999999", multiplicant: "999999999999999999999999999",
		},
		"decimal small": {
			multiplier: "0.0000000001", multiplicant: "10",
		},
		"decimal small/decimal small ": {
			multiplier: "0.0000000001", multiplicant: "0.0001",
		},
	}
	for name, spec := range specs {
		b.Run(name, func(b *testing.B) {
			b.Run("LegacyDec", func(b *testing.B) {
				dv, ds := LegacyMustNewDecFromStr(spec.multiplier), LegacyMustNewDecFromStr(spec.multiplicant)
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					_ = dv.Mul(ds)
				}
			})

			b.Run("NewDec", func(b *testing.B) {
				dv, ds := must(NewDecFromString(spec.multiplier)), must(NewDecFromString(spec.multiplicant))
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					_, _ = dv.Mul(ds)
				}
			})
		})
	}
}

func BenchmarkCompareLegacyDecAndNewDecMarshalUnmarshal(b *testing.B) {
	specs := map[string]struct {
		src string
	}{
		"small": {
			src: "1",
		},
		"big18": {
			src: "999999999999999999",
		},
		"negative big34": {
			src: "9999999999999999999999999999999999",
		},
		"decimal": {
			src: "12345.678901234341",
		},
	}
	for name, spec := range specs {
		b.Run(name, func(b *testing.B) {
			b.Run("LegacyDec", func(b *testing.B) {
				src := LegacyMustNewDecFromStr(spec.src)
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					bz, err := src.Marshal()
					require.NoError(b, err)
					var d LegacyDec
					require.NoError(b, d.Unmarshal(bz))
				}
			})

			b.Run("NewDec", func(b *testing.B) {
				src := must(NewDecFromString(spec.src))
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					bz, err := src.Marshal()
					require.NoError(b, err)
					var d Dec
					require.NoError(b, d.Unmarshal(bz))
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
