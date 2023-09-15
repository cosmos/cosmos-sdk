package math

import (
	"math/big"
	"strconv"
	"strings"
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

var (
	max5Percent, _ = LegacyNewDecFromStr("5") // 5% max tolerance of difference in values
	decDiv100, _   = LegacyNewDecFromStr("0.01")
)

func FuzzLegacyDecApproxRoot(f *testing.F) {
	if testing.Short() {
		f.Skip("running in -short mode")
	}

	// 1. Add the corpus: <LEGACY_DEC>,<ROOT>
	f.Add("-1000000000.5,5")
	f.Add("1000000000.5,5")
	f.Add("128,8")
	f.Add("-128,8")
	f.Add("100000,2")
	f.Add("-100000,2")

	// 2. Now fuzz it.
	f.Fuzz(func(t *testing.T, input string) {
		splits := strings.Split(input, ",")
		if len(splits) < 2 {
			// Invalid input, just skip over it.
			return
		}

		decStr, powStr := splits[0], splits[1]
		nthRoot, err := strconv.ParseUint(powStr, 10, 64)
		if err != nil {
			// Invalid input, nothing to do here.
			return
		}
		dec, err := LegacyNewDecFromStr(decStr)
		if err != nil {
			// Invalid input, nothing to do here.
			return
		}

		// Ensure that we aren't passing in a power larger than the value itself.
		nthRootAsDec, err := LegacyNewDecFromStr(powStr)
		if err != nil {
			// Invalid input, nothing to do here.
			return
		}
		if nthRootAsDec.GTE(dec) {
			// nthRoot cannot be greater than or equal to the value itself, return.
			return
		}

		gotApproxSqrt, err := dec.ApproxRoot(nthRoot)
		if err != nil {
			if strings.Contains(err.Error(), "out of bounds") {
				return
			}
			t.Fatalf("\nGiven: %s, nthRoot: %d\nerr: %v", dec, nthRoot, err)
		}

		// For more focused usage and easy parity checks, we are just doing only
		// square root comparisons, hence any nthRoot != 2 can end the journey here!
		if nthRoot != 2 {
			return
		}

		// Firstly ensure that gotApproxSqrt * gotApproxSqrt is
		// super duper close to the value of dec.
		squared := gotApproxSqrt.Mul(gotApproxSqrt)
		if !squared.Equal(dec) {
			diff := squared.Sub(dec).Abs().Mul(decDiv100).Quo(dec)
			if diff.GTE(max5Percent) {
				t.Fatalf("Discrepancy:\n(%s)^2 != %s\n\tGot: %s\nDiscrepancy %%: %s", gotApproxSqrt, dec, squared, diff)
			}
		}

		// By this point we are dealing with square root.
		// Now roundtrip to ensure that the difference between the
		// expected value and that approximation isn't off by 5%.
		stdlibFloat, ok := new(big.Float).SetString(decStr)
		if !ok {
			return
		}
		origWasNegative := stdlibFloat.Sign() == -1
		if origWasNegative {
			// Make it an absolute value to avoid panics
			// due to passing in negative values into .Sqrt.
			stdlibFloat = new(big.Float).Abs(stdlibFloat)
		}

		stdlibSqrt := new(big.Float).Sqrt(stdlibFloat)
		if origWasNegative {
			// Invert the sign to maintain parity with cosmossdk.io/math.LegacyDec.ApproxRoot
			// which returns a negative value even for square roots.
			stdlibSqrt = new(big.Float).Neg(stdlibSqrt)
		}

		stdlibSqrtAsDec, err := LegacyNewDecFromStr(stdlibSqrt.String())
		if err != nil {
			return
		}

		diff := stdlibSqrtAsDec.Sub(gotApproxSqrt).Abs().Mul(decDiv100).Quo(gotApproxSqrt)
		if diff.IsNegative() {
			diff = diff.Neg()
		}
		if diff.GT(max5Percent) {
			t.Fatalf("\nGiven: sqrt(%s)\nPrecision loss as the difference %s > %s\n"+
				"Stdlib sqrt: %+60s\ncosmossdk.io/math.*Dec.ApproxSqrt: %+60s",
				dec, diff, max5Percent, stdlibSqrtAsDec, gotApproxSqrt)
		}
	})
}
