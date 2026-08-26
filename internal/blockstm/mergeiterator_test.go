package blockstm

import (
	"testing"

	"github.com/test-go/testify/require"

	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
)

func TestCacheMergeIteratorMemoizedSource(t *testing.T) {
	parent := NewMemDB()
	for key, value := range map[string]string{
		"a": "parent-a",
		"c": "parent-c",
		"e": "parent-e",
	} {
		parent.Set([]byte(key), []byte(value))
	}

	cache := NewWriteSet(bytesIsZero, bytesValueLen)
	cache.OverlaySet([]byte("b"), []byte("cache-b"))
	cache.OverlaySet([]byte("c"), []byte("cache-c"))
	cache.OverlaySet([]byte("d"), nil)
	cache.OverlaySet([]byte("e"), nil)
	cache.OverlaySet([]byte("f"), []byte("cache-f"))

	for _, ascending := range []bool{true, false} {
		t.Run(map[bool]string{true: "ascending", false: "descending"}[ascending], func(t *testing.T) {
			var parentIter, cacheIter storetypes.GIterator[[]byte]
			if ascending {
				parentIter = parent.Iterator(nil, nil)
				cacheIter = cache.Iterator(nil, nil)
			} else {
				parentIter = parent.ReverseIterator(nil, nil)
				cacheIter = cache.ReverseIterator(nil, nil)
			}

			iter := NewCacheMergeIterator(parentIter, cacheIter, ascending, nil, bytesIsZero)
			var got []string
			for ; iter.Valid(); iter.Next() {
				got = append(got, string(iter.Key())+"="+string(iter.Value()))
			}
			require.NoError(t, iter.Close())

			expected := []string{"a=parent-a", "b=cache-b", "c=cache-c", "f=cache-f"}
			if !ascending {
				expected = []string{"f=cache-f", "c=cache-c", "b=cache-b", "a=parent-a"}
			}
			require.Equal(t, expected, got)
		})
	}
}
