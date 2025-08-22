package codec_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"cosmossdk.io/collections"
	"cosmossdk.io/collections/codec"
)

func TestStringKeyCodecNonTerminal(t *testing.T) {
	stringCodec := codec.NewStringKeyCodec[string]()

	t.Run("EncodeNonTerminal adds delimiter", func(t *testing.T) {
		key := "hello"
		buffer := make([]byte, stringCodec.SizeNonTerminal(key))
		
		written, err := stringCodec.EncodeNonTerminal(buffer, key)
		require.NoError(t, err)
		require.Equal(t, len(key)+1, written)
		
		// Check that delimiter is added at the end
		require.Equal(t, codec.StringDelimiter, buffer[len(key)])
		
		// Verify the string content is correct
		require.Equal(t, []byte("hello"), buffer[:len(key)])
	})

	t.Run("DecodeNonTerminal finds delimiter", func(t *testing.T) {
		// Create a buffer with string + delimiter
		key := "world"
		buffer := make([]byte, len(key)+1)
		copy(buffer, key)
		buffer[len(key)] = codec.StringDelimiter
		
		read, decoded, err := stringCodec.DecodeNonTerminal(buffer)
		require.NoError(t, err)
		require.Equal(t, len(key)+1, read)
		require.Equal(t, key, decoded)
	})

	t.Run("EncodeNonTerminal rejects strings with delimiter", func(t *testing.T) {
		// Create a string that contains the delimiter
		key := "hello\x00world" // \x00 is StringDelimiter
		buffer := make([]byte, stringCodec.SizeNonTerminal(key))
		
		_, err := stringCodec.EncodeNonTerminal(buffer, key)
		require.Error(t, err)
		require.Contains(t, err.Error(), "string delimiter")
	})

	t.Run("EncodeNonTerminal buffer too small", func(t *testing.T) {
		key := "hello"
		buffer := make([]byte, len(key)) // Buffer too small
		
		_, err := stringCodec.EncodeNonTerminal(buffer, key)
		require.Error(t, err)
		require.Contains(t, err.Error(), "buffer too small")
	})

	t.Run("DecodeNonTerminal no delimiter found", func(t *testing.T) {
		// Buffer without delimiter
		buffer := []byte("hello")
		
		_, _, err := stringCodec.DecodeNonTerminal(buffer)
		require.Error(t, err)
		require.Contains(t, err.Error(), "no instances of the string delimiter")
	})

	t.Run("Round trip encoding/decoding", func(t *testing.T) {
		key := "test_string"
		buffer := make([]byte, stringCodec.SizeNonTerminal(key))
		
		// Encode
		written, err := stringCodec.EncodeNonTerminal(buffer, key)
		require.NoError(t, err)
		require.Equal(t, len(key)+1, written)
		
		// Decode
		read, decoded, err := stringCodec.DecodeNonTerminal(buffer)
		require.NoError(t, err)
		require.Equal(t, written, read)
		require.Equal(t, key, decoded)
	})
}

func TestStringKeyCodecInPair(t *testing.T) {
	// Test that string codec works correctly in multipart keys
	pairCodec := collections.PairKeyCodec(collections.StringKey, collections.StringKey)
	
	t.Run("Pair with string keys", func(t *testing.T) {
		pair := collections.Join("first", "second")
		buffer := make([]byte, pairCodec.Size(pair))
		
		// Encode
		written, err := pairCodec.Encode(buffer, pair)
		require.NoError(t, err)
		require.Equal(t, len(buffer), written)
		
		// Decode
		read, decoded, err := pairCodec.Decode(buffer)
		require.NoError(t, err)
		require.Equal(t, written, read)
		require.Equal(t, pair, decoded)
	})
	
	t.Run("Pair with empty strings", func(t *testing.T) {
		pair := collections.Join("", "")
		buffer := make([]byte, pairCodec.Size(pair))
		
		// Encode
		written, err := pairCodec.Encode(buffer, pair)
		require.NoError(t, err)
		require.Equal(t, len(buffer), written)
		
		// Decode
		read, decoded, err := pairCodec.Decode(buffer)
		require.NoError(t, err)
		require.Equal(t, written, read)
		require.Equal(t, pair, decoded)
	})
} 
