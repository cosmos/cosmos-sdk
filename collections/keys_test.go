package collections

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUint64Key(t *testing.T) {
	t.Run("bijective", func(t *testing.T) {
		checkKeyCodec(t, Uint64Key, 55)
	})

	t.Run("invalid key size", func(t *testing.T) {
		_, _, err := Uint64Key.DecodeKey([]byte{0x0, 0x1})
		require.ErrorIs(t, err, errDecodeKeySize)
	})
}
