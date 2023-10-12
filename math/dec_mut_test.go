package math

import "testing"

var sink any

func BenchmarkLegacyDecMut(b *testing.B) {
	b.ReportAllocs()

	d := LegacyMustNewDecFromStr("123456789.123456789")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink = d.Mut()
	}
	if sink == nil {
		b.Fatal("Benchmark was not run")
	}
	sink = nil
}

// before the conversions:
// BenchmarkLegacyDec_NegMut-10    	496984528	         2.431 ns/op	       0 B/op	       0 allocs/op
// after conversions:
// BenchmarkLegacyDec_NegMut-10    	482561031	         2.471 ns/op	       0 B/op	       0 allocs/op
func BenchmarkLegacyDec_NegMut(b *testing.B) {
	b.ReportAllocs()

	d := LegacyMustNewDecFromStr("123456789.123456789")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink = d.NegMut()
	}

	if sink == nil {
		b.Fatal("Benchmark was not run")
	}
	sink = nil
}

func BenchmarkLegacyDecMut_Neg(b *testing.B) {
	b.ReportAllocs()

	d := LegacyMustNewDecFromStr("123456789.123456789").Mut()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		sink = d.Neg()
	}
	if sink == nil {
		b.Fatal("Benchmark was not run")
	}
	sink = nil
}

func BenchmarkLegacyDec_PowerMut(b *testing.B) {
	b.ReportAllocs()

	d := LegacyMustNewDecFromStr("0.00000001")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		sink = d.PowerMut(500)
	}

	if sink == nil {
		b.Fatal("Benchmark was not run")
	}
	sink = nil
}

func BenchmarkLegacyDecMut_Power(b *testing.B) {
	b.ReportAllocs()

	d := LegacyMustNewDecFromStr("0.00000001").Mut()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		sink = d.Power(500)
	}

	if sink == nil {
		b.Fatal("Benchmark was not run")
	}
	sink = nil
}
