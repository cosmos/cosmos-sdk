package codec

import (
	"bytes"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
	"sort"
	"testing"
)

func TestIntKeys(t *testing.T) {
	kc := NewInt64Key[int64]()
	rapid.Check(t, func(t *rapid.T) {
		slice := rapid.SliceOfN(rapid.Int64(), 50_000, 100_000).Draw(t, "random ints")
		sort.Slice(slice, func(i, j int) bool {
			return slice[i] < slice[j]
		})

		var current []byte
		for _, i := range slice {
			next := make([]byte, kc.Size(i))
			_, err := kc.Encode(next, i)
			require.NoError(t, err)
			cmp := bytes.Compare(current, next)
			require.True(t, cmp == 0 || cmp == -1)
		}
	})
}
