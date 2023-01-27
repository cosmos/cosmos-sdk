package collections

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUint64Key(t *testing.T) {
	t.Run("invalid key size", func(t *testing.T) {
		_, _, err := Uint64Key.Decode([]byte{0x0, 0x1})
		require.ErrorIs(t, err, errDecodeKeySize)
	})
}

func TestStringKey(t *testing.T) {
}
