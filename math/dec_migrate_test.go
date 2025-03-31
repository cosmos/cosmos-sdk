package math_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"
)

// TestDecFromLegacyDec verifies that converting a LegacyDec to a Dec via string round-trip works as expected.
func TestDecFromLegacyDec(t *testing.T) {
	// Define test cases: a list of valid decimal string representations.
	// Note: The legacy format always prints exactly 18 decimal places.
	testCases := []struct {
		name     string
		inputStr string
	}{
		{"Zero", "0"},
		{"One", "1"},
		{"NegativeOne", "-1"},
		{"IntegerWithNoDecimals", "123456789012345678"},
		{"SimpleDecimal", "123.456"},
		{"NegativeDecimal", "-9876.543210"},
		{"SmallestUnit", "0.000000000000000001"}, // 10^-18
		{"LargeNumber", "12345678901234567890.123456789012345678"},
		{"TrailingZeros", "100.000000000000000000"},
	}

	for _, tc := range testCases {
		// capture range variable
		t.Run(tc.name, func(t *testing.T) {
			// Create a LegacyDec from the test input string.
			legacyDec, err := math.LegacyNewDecFromStr(tc.inputStr)
			require.NoError(t, err)

			// Convert using our conversion function.
			dec, err := math.DecFromLegacyDec(legacyDec)
			require.NoError(t, err)

			// Convert directly from the input string for a canonical value.
			expectedDec, err := math.NewDecFromString(tc.inputStr)
			require.NoError(t, err)

			// Compare the two Dec values.
			require.True(t, dec.Equal(expectedDec))
		})
	}
}

// FuzzDecFromLegacyDec fuzzes the conversion function from LegacyDec to Dec.
func FuzzDecFromLegacyDec(f *testing.F) {
	// Seed the fuzzer with some valid input strings.
	seedInputs := []string{
		"0",
		"1",
		"-1",
		"123.456",
		"-9876.543210",
		"0.000000000000000001",
		"100.000000000000000000",
		"12345678901234567890.123456789012345678",
	}
	for _, s := range seedInputs {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, inputStr string) {
		// Attempt to create a LegacyDec from the fuzz input.
		legacyDec, err := math.LegacyNewDecFromStr(inputStr)
		if err != nil {
			// Ignore inputs that do not form a valid LegacyDec.
			return
		}

		// Convert using the conversion function.
		dec, err := math.DecFromLegacyDec(legacyDec)
		require.NoError(t, err)

		// Convert directly from the legacy string output.
		expectedDec, err := math.NewDecFromString(legacyDec.String())
		require.NoError(t, err)

		require.True(t, dec.Equal(expectedDec))
	})
}
