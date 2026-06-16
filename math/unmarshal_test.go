package math_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"
)

// TestUnmarshalEmptyIsZeroNotNil ensures that unmarshaling empty bytes (e.g. a
// zero-length protobuf custom-type field, which is decodable wire data) leaves a
// usable zero value rather than a receiver whose internal big.Int is nil. A nil
// internal value makes later methods such as IsZero / IsNegative panic.
func TestUnmarshalEmptyIsZeroNotNil(t *testing.T) {
	t.Run("LegacyDec", func(t *testing.T) {
		var d math.LegacyDec
		require.NoError(t, d.Unmarshal([]byte{}))
		require.False(t, d.IsNil())
		require.True(t, d.IsZero())      // panicked before the fix
		require.False(t, d.IsNegative()) // panicked before the fix
	})

	t.Run("Int", func(t *testing.T) {
		var i math.Int
		require.NoError(t, i.Unmarshal([]byte{}))
		require.False(t, i.IsNil())
		require.True(t, i.IsZero()) // panicked before the fix
	})

	t.Run("Uint", func(t *testing.T) {
		var u math.Uint
		require.NoError(t, u.Unmarshal([]byte{}))
		require.False(t, u.IsNil())
		require.True(t, u.IsZero()) // panicked before the fix
	})
}
