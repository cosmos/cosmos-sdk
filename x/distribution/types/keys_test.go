package types

import (
	"testing"

	"cosmossdk.io/collections/colltest"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

func TestLEUint64Key(t *testing.T) {
	t.Run("conformance", rapid.MakeCheck(func(r *rapid.T) {
		colltest.TestKeyCodec(t, LEUint64Key, rapid.Uint64().Draw(r, "uint64"))
	}))

	t.Run("buffer too small", func(t *testing.T) {
		_, _, err := LEUint64Key.Decode([]byte{0})
		require.ErrorContains(t, err, "invalid buffer size")
	})

}
