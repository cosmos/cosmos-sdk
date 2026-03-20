package cachekv_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/store/v2/types"
	"github.com/stretchr/testify/require"
)

func collectIterator(iter types.Iterator) (keys, values [][]byte) {
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		keys = append(keys, iter.Key())
		values = append(values, iter.Value())
	}
	return
}

func TestIterator_EmptyStore(t *testing.T) {
	st, _ := newStoreWithParent()

	iter := st.Iterator(nil, nil)
	defer iter.Close()

	require.False(t, iter.Valid())
}

func TestIterator_ParentOnlyWhenClean(t *testing.T) {
	st, parent := newStoreWithParent()
	parent.Set([]byte("a"), []byte("1"))
	parent.Set([]byte("b"), []byte("2"))

	keys, values := collectIterator(st.Iterator(nil, nil))

	require.Equal(t, [][]byte{[]byte("a"), []byte("b")}, keys)
	require.Equal(t, [][]byte{[]byte("1"), []byte("2")}, values)
}

func TestIterator_MergesCacheAndParent(t *testing.T) {
	st, parent := newStoreWithParent()
	parent.Set([]byte("a"), []byte("1"))
	parent.Set([]byte("c"), []byte("3"))

	st.Set([]byte("b"), []byte("2"))

	keys, _ := collectIterator(st.Iterator(nil, nil))

	require.Equal(t, [][]byte{[]byte("a"), []byte("b"), []byte("c")}, keys)
}

func TestIterator_CacheOverridesParent(t *testing.T) {
	st, parent := newStoreWithParent()
	parent.Set([]byte("key"), []byte("parent"))

	st.Set([]byte("key"), []byte("cache"))

	keys, values := collectIterator(st.Iterator(nil, nil))

	require.Equal(t, [][]byte{[]byte("key")}, keys)
	require.Equal(t, [][]byte{[]byte("cache")}, values)
}

func TestIterator_DeletedKeysSkipped(t *testing.T) {
	st, parent := newStoreWithParent()
	parent.Set([]byte("a"), []byte("1"))
	parent.Set([]byte("b"), []byte("2"))
	parent.Set([]byte("c"), []byte("3"))

	st.Delete([]byte("b"))

	keys, _ := collectIterator(st.Iterator(nil, nil))

	require.Equal(t, [][]byte{[]byte("a"), []byte("c")}, keys)
}

func TestIterator_RespectsBounds(t *testing.T) {
	st, parent := newStoreWithParent()
	parent.Set([]byte("a"), []byte("1"))
	parent.Set([]byte("b"), []byte("2"))
	parent.Set([]byte("c"), []byte("3"))
	parent.Set([]byte("d"), []byte("4"))

	keys, _ := collectIterator(st.Iterator([]byte("b"), []byte("d")))

	require.Equal(t, [][]byte{[]byte("b"), []byte("c")}, keys)
}

func TestIterator_NilBounds(t *testing.T) {
	st, parent := newStoreWithParent()
	parent.Set([]byte("a"), []byte("1"))
	parent.Set([]byte("z"), []byte("26"))

	keys, _ := collectIterator(st.Iterator(nil, nil))

	require.Len(t, keys, 2)
	require.Equal(t, []byte("a"), keys[0])
	require.Equal(t, []byte("z"), keys[1])
}

func TestIterator_AscendingOrder(t *testing.T) {
	st, parent := newStoreWithParent()
	parent.Set([]byte("c"), []byte("3"))
	parent.Set([]byte("a"), []byte("1"))
	st.Set([]byte("b"), []byte("2"))

	keys, _ := collectIterator(st.Iterator(nil, nil))

	require.Equal(t, [][]byte{[]byte("a"), []byte("b"), []byte("c")}, keys)
}

func TestReverseIterator_DescendingOrder(t *testing.T) {
	st, parent := newStoreWithParent()
	parent.Set([]byte("a"), []byte("1"))
	parent.Set([]byte("c"), []byte("3"))
	st.Set([]byte("b"), []byte("2"))

	keys, _ := collectIterator(st.ReverseIterator(nil, nil))

	require.Equal(t, [][]byte{[]byte("c"), []byte("b"), []byte("a")}, keys)
}

func TestReverseIterator_DeletedKeysSkipped(t *testing.T) {
	st, parent := newStoreWithParent()
	parent.Set([]byte("a"), []byte("1"))
	parent.Set([]byte("b"), []byte("2"))
	parent.Set([]byte("c"), []byte("3"))

	st.Delete([]byte("b"))

	keys, _ := collectIterator(st.ReverseIterator(nil, nil))

	require.Equal(t, [][]byte{[]byte("c"), []byte("a")}, keys)
}

func TestReverseIterator_RespectsBounds(t *testing.T) {
	st, parent := newStoreWithParent()
	parent.Set([]byte("a"), []byte("1"))
	parent.Set([]byte("b"), []byte("2"))
	parent.Set([]byte("c"), []byte("3"))
	parent.Set([]byte("d"), []byte("4"))

	// end is exclusive
	keys, _ := collectIterator(st.ReverseIterator([]byte("b"), []byte("d")))

	require.Equal(t, [][]byte{[]byte("c"), []byte("b")}, keys)
}

func TestIterator_AfterWrite(t *testing.T) {
	st, parent := newStoreWithParent()
	st.Set([]byte("a"), []byte("1"))
	st.Write()

	// now the store is clean, should use parent iterator
	keys, _ := collectIterator(st.Iterator(nil, nil))

	require.Equal(t, [][]byte{[]byte("a")}, keys)
	require.Equal(t, []byte("1"), parent.Get([]byte("a")))
}
