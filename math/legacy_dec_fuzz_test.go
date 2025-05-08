package math

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
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
		require.NoError(t, err)
		require.True(t, !dec.IsNil())
	})
}

func FuzzLegacyDecMarshalUnmarshalJSON(f *testing.F) {
	// Seed with some valid decimal strings.
	seeds := []string{
		"0", "123", "-123", "123.456", "-123.456", "1.23E4", "1.23e4", "1.23456789E-10",
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		// Try to create a LegacyDec from the input.
		dec, err := LegacyNewDecFromStr(input)
		if err != nil {
			// Skip inputs that cannot be parsed.
			t.Skip()
		}

		// Marshal to JSON.
		jsonData, err := dec.MarshalJSON()
		require.NoError(t, err)

		// Unmarshal back.
		var decoded LegacyDec
		err = decoded.UnmarshalJSON(jsonData)
		require.NoError(t, err)

		// Check that the round-trip value is equal.
		require.True(t, dec.Equal(decoded), fmt.Sprintf("JSON round-trip mismatch for input %q: original %q, decoded %q", input, dec.String(), decoded.String()))
	})
}

func FuzzLegacyDecMarshalUnmarshal(f *testing.F) {
	// Seed with some valid decimal strings.
	seeds := []string{
		"0", "123", "-123", "123.456", "-123.456", "1.23E4", "1.23e4", "1.23456789E-10",
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		// Parse the input into a LegacyDec.
		dec, err := LegacyNewDecFromStr(input)
		if err != nil {
			// Skip invalid inputs.
			t.Skip()
		}

		// Marshal using the custom binary (gogo proto) encoding.
		bz, err := dec.Marshal()
		require.NoError(t, err)

		// Unmarshal back.
		var decoded LegacyDec
		err = decoded.Unmarshal(bz)
		require.NoError(t, err)

		// Check that the round-trip value is equal.
		require.True(t, dec.Equal(decoded), fmt.Sprintf("JSON round-trip mismatch for input %q: original %q, decoded %q", input, dec.String(), decoded.String()))
	})
}
