package codec

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestUntypedValueCodec is a test function for the UntypedValueCodec.
// It checks the encoding, decoding, JSON encoding/decoding, and stringifying behaviors.
func TestUntypedValueCodec(t *testing.T) {
	// Create a new UntypedValueCodec, which converts a key to a value codec using string keys.
	vc := NewUntypedValueCodec(ValueCodec[string](KeyToValueCodec(KeyCodec[string](NewStringKeyCodec[string]()))))

	// Sub-test for encoding and decoding values.
	t.Run("encode/decode", func(t *testing.T) {
		// Attempt to encode an integer, which should fail and return an ErrEncoding.
		_, err := vc.Encode(0)
		require.ErrorIs(t, err, ErrEncoding)

		// Encode the string "hello", which should succeed without error.
		b, err := vc.Encode("hello")
		require.NoError(t, err)

		// Decode the previously encoded value and ensure it matches "hello".
		value, err := vc.Decode(b)
		require.NoError(t, err)
		require.Equal(t, "hello", value)
	})

	// Sub-test for JSON encoding and decoding.
	t.Run("json encode/decode", func(t *testing.T) {
		// Attempt to JSON encode an integer, which should fail and return an ErrEncoding.
		_, err := vc.EncodeJSON(0)
		require.ErrorIs(t, err, ErrEncoding)

		// JSON encode the string "hello", which should succeed without error.
		b, err := vc.EncodeJSON("hello")
		require.NoError(t, err)

		// JSON decode the previously encoded value and ensure it matches "hello".
		value, err := vc.DecodeJSON(b)
		require.NoError(t, err)
		require.Equal(t, "hello", value)
	})

	// Sub-test for stringifying values.
	t.Run("stringify", func(t *testing.T) {
		// Attempt to stringify an integer, which should fail and return an ErrEncoding.
		_, err := vc.Stringify(0)
		require.ErrorIs(t, err, ErrEncoding)

		// Stringify the string "hello", which should succeed without error.
		s, err := vc.Stringify("hello")
		require.NoError(t, err)

		// Ensure the stringified value matches "hello".
		require.Equal(t, "hello", s)
	})
}
