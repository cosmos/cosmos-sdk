package collections

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUint64Value(t *testing.T) {
	t.Run("invalid size", func(t *testing.T) {
		_, err := Uint64Value.Decode([]byte{0x1, 0x2})
		require.ErrorIs(t, err, ErrEncoding)
	})
}

func TestUInt64JSON(t *testing.T) {
	var x uint64 = 3076
	bz, err := uint64EncodeJSON(x)
	require.NoError(t, err)
	require.Equal(t, []byte(`"3076"`), bz)
	y, err := uint64DecodeJSON(bz)
	require.NoError(t, err)
	require.Equal(t, x, y)
}
