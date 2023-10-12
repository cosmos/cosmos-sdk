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
