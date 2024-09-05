package pebbledb

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMVCCKey(t *testing.T) {
	for i := uint64(1); i < 1001; i++ {
		keyA := MVCCEncode([]byte("key001"), i)

		key, vBz, ok := SplitMVCCKey(keyA)

		version, err := decodeUint64Ascending(vBz)
		require.NoError(t, err)
		require.True(t, ok)
		require.Equal(t, i, version)
		require.Equal(t, []byte("key001"), key)
	}
}

func TestMVCCKeyCompare(t *testing.T) {
	testCases := []struct {
		keyA     []byte
		keyB     []byte
		expected int
	}{
		{
			// same key, same version
			keyA:     MVCCEncode([]byte("key001"), 1),
			keyB:     MVCCEncode([]byte("key001"), 1),
			expected: 0,
		},
		{
			// same key, different version
			keyA:     MVCCEncode([]byte("key001"), 1),
			keyB:     MVCCEncode([]byte("key001"), 2),
			expected: -1,
		},
		{
			// same key, different version (inverse)
			keyA:     MVCCEncode([]byte("key001"), 2),
			keyB:     MVCCEncode([]byte("key001"), 1),
			expected: 1,
		},
		{
			// different key, same version
			keyA:     MVCCEncode([]byte("key001"), 1),
			keyB:     MVCCEncode([]byte("key009"), 1),
			expected: -1,
		},
	}

	for _, tc := range testCases {
		require.Equalf(t, tc.expected, MVCCKeyCompare(tc.keyA, tc.keyB), "keyA: %s, keyB: %s", tc.keyA, tc.keyB)
	}
}
