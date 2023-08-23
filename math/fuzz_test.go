package math

import (
	"testing"
)

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

func FuzzSmallIntSize(f *testing.F) {
	f.Add(int64(2<<53 - 1))
	f.Add(-int64(2<<53 - 1))
	f.Fuzz(func(t *testing.T, input int64) {
		i := NewInt(input)
		exp, _ := i.Marshal()
		if i.Size() != len(exp) {
			t.Fatalf("input %d: i.Size()=%d, len(input)=%d", input, i.Size(), len(exp))
		}
	})
}
