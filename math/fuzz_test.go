package math

import (
	"testing"
)

func FuzzNewIntFromStringSize(f *testing.F) {
	if testing.Short() {
		f.Skip("running in -short mode")
	}

	// TODO add more test cases
	f.Add("0")
	f.Add("1")
	f.Add("123456789")
	f.Add("-123456789")
	f.Add("9999999999999999000")
	f.Add("-123456789123456789123456789")
	f.Add("1000000000000000000000000000")
	f.Add("-1000000000000000000000000000")
	f.Add("123456789123456789123456789")
	f.Add("-123456789123456789123456789")
	f.Add("0x123456789abcdef")
	f.Add("-0x123456789abcdef")

	f.Fuzz(func(t *testing.T, input string) {
		i, _ := NewIntFromString(input)
		if i.Size() != len(input) {
			t.Fatalf("input %s: i.Size()=%d, len(input)=%d", input, i.Size(), len(input))
		}
	})
}

func FuzzLegacyNewDecFromStr(f *testing.F) {
	if testing.Short() {
		f.Skip("running in -short mode")
	}

	f.Add("-123.456")
	f.Add("123.456789")
	f.Add("123456789")
	f.Add("0.12123456789")
	f.Add("-12123456789")

	f.Fuzz(func(t *testing.T, input string) {
		dec, err := LegacyNewDecFromStr(input)
		if err != nil && !dec.IsNil() {
			t.Fatalf("Inconsistency: dec.notNil=%v yet err=%v", dec, err)
		}
	})
}
