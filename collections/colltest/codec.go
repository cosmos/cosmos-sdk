package colltest

import (
	"cosmossdk.io/collections"
	"github.com/stretchr/testify/require"
	"testing"
)

// TestKeyCodec asserts the correct behaviour of a KeyCodec over the type T.
func TestKeyCodec[T any](t *testing.T, keyCodec collections.KeyCodec[T], key T) {
	buffer := make([]byte, keyCodec.Size(key))
	written, err := keyCodec.Encode(buffer, key)
	require.NoError(t, err)
	require.Equal(t, len(buffer), written)
	read, decodedKey, err := keyCodec.Decode(buffer)
	require.NoError(t, err)
	require.Equal(t, len(buffer), read, "encoded key and read bytes must have same size")
	require.Equal(t, key, decodedKey, "encoding and decoding produces different keys")
	// test if terminality is correctly applied
	pairCodec := collections.PairKeyCodec(keyCodec, collections.StringKey)
	pairKey := collections.Join(key, "TEST")
	buffer = make([]byte, pairCodec.Size(pairKey))
	written, err = pairCodec.Encode(buffer, pairKey)
	require.NoError(t, err)
	read, decodedPairKey, err := pairCodec.Decode(buffer)
	require.NoError(t, err)
	require.Equal(t, len(buffer), read, "encoded non terminal key and pair key read bytes must have same size")
	require.Equal(t, pairKey, decodedPairKey, "encoding and decoding produces different keys with non terminal encoding")

	// check JSON
	keyJSON, err := keyCodec.EncodeJSON(key)
	require.NoError(t, err)
	decoded, err := keyCodec.DecodeJSON(keyJSON)
	require.NoError(t, err)
	require.Equal(t, key, decoded, "json encoding and decoding did not produce the same results")
}

// TestValueCodec asserts the correct behaviour of a ValueCodec over the type T.
func TestValueCodec[T any](t *testing.T, encoder collections.ValueCodec[T], value T) {
	encodedValue, err := encoder.Encode(value)
	require.NoError(t, err)
	decodedValue, err := encoder.Decode(encodedValue)
	require.NoError(t, err)
	require.Equal(t, value, decodedValue, "encoding and decoding produces different values")

	encodedJSONValue, err := encoder.EncodeJSON(value)
	require.NoError(t, err)
	decodedJSONValue, err := encoder.DecodeJSON(encodedJSONValue)
	require.NoError(t, err)
	require.Equal(t, value, decodedJSONValue, "encoding and decoding in json format produces different values")

	require.NotEmpty(t, encoder.ValueType())

	_ = encoder.Stringify(value)
}
