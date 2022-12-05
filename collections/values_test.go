package collections

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestUint64Value(t *testing.T) {
	t.Run("bijective", func(t *testing.T) {
		assertValue(t, Uint64Value, 555)
	})

	t.Run("invalid size", func(t *testing.T) {
		_, err := Uint64Value.Decode([]byte{0x1, 0x2})
		require.ErrorIs(t, err, ErrEncoding)
	})
}
