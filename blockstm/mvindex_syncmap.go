//go:build syncmap

package blockstm

import (
	"bytes"
	"sort"
	"sync"
	"sync/atomic"

	"github.com/cosmos/cosmos-sdk/blockstm/tree"
)

// minSnapshotKeysCap is the minimum capacity used when building a key snapshot.
const minSnapshotKeysCap = 1024

// mvIndexSyncMapValue holds the stable key bytes and the per-key version tree.
type mvIndexSyncMapValue[V any] struct {
	key  []byte
	data *tree.BTree[secondaryDataItem[V]]
}

// mvIndex is an index from key -> per-key version tree, implemented with sync.Map.
// Keys are converted to string for map lookup.
type mvIndex[V any] struct {
	m sync.Map // map[string]*mvIndexSyncMapValue[V]

	// keySetVersion increments when a new key is added; it invalidates the snapshot cache.
	keySetVersion atomic.Uint64

	cacheMu      sync.RWMutex
	cacheVersion uint64
	// cacheAllKeysAsc is an immutable, ascending snapshot of all keys.
	cacheAllKeysAsc []Key
}

func newMVIndex[V any]() *mvIndex[V] {
	return &mvIndex[V]{}
}

func (idx *mvIndex[V]) get(key []byte) *tree.BTree[secondaryDataItem[V]] {
	if idx == nil {
		return nil
	}
	v, ok := idx.m.Load(string(key))
	if !ok {
		return nil
	}
	return v.(*mvIndexSyncMapValue[V]).data
}

func (idx *mvIndex[V]) getOrCreate(key []byte) *tree.BTree[secondaryDataItem[V]] {
	sKey := string(key)
	if v, ok := idx.m.Load(sKey); ok {
		return v.(*mvIndexSyncMapValue[V]).data
	}

	kCopy := append([]byte(nil), key...)
	data := tree.NewBTree(secondaryLesser[V], InnerBTreeDegree)
	val := &mvIndexSyncMapValue[V]{key: kCopy, data: data}

	actual, loaded := idx.m.LoadOrStore(sKey, val)
	if loaded {
		return actual.(*mvIndexSyncMapValue[V]).data
	}
	idx.keySetVersion.Add(1)
	return data
}

func (idx *mvIndex[V]) snapshotKeys(start, end []byte, ascending bool) []Key {
	if idx == nil {
		return nil
	}

	keysAsc := idx.snapshotAllKeysAsc()
	if len(keysAsc) == 0 {
		return nil
	}

	lo := 0
	if start != nil {
		lo = sort.Search(len(keysAsc), func(i int) bool { return bytes.Compare(keysAsc[i], start) >= 0 })
	}
	hi := len(keysAsc)
	if end != nil {
		hi = sort.Search(len(keysAsc), func(i int) bool { return bytes.Compare(keysAsc[i], end) >= 0 })
	}
	if lo > hi {
		return nil
	}

	view := keysAsc[lo:hi]
	if ascending {
		return view
	}
	return reverseCopyKeys(view)
}

func (idx *mvIndex[V]) snapshotAllKeys(ascending bool) []Key {
	if idx == nil {
		return nil
	}
	keysAsc := idx.snapshotAllKeysAsc()
	if ascending {
		return keysAsc
	}
	return reverseCopyKeys(keysAsc)
}

func (idx *mvIndex[V]) snapshotAllKeysAsc() []Key {
	ver := idx.keySetVersion.Load()
	idx.cacheMu.RLock()
	if idx.cacheVersion == ver {
		keys := idx.cacheAllKeysAsc
		idx.cacheMu.RUnlock()
		return keys
	}
	idx.cacheMu.RUnlock()

	idx.cacheMu.Lock()
	defer idx.cacheMu.Unlock()
	ver = idx.keySetVersion.Load()
	if idx.cacheVersion == ver {
		return idx.cacheAllKeysAsc
	}

	capHint := len(idx.cacheAllKeysAsc)
	if capHint < minSnapshotKeysCap {
		capHint = minSnapshotKeysCap
	}

	keys := make([]Key, 0, capHint)
	idx.m.Range(func(_, v any) bool {
		keys = append(keys, v.(*mvIndexSyncMapValue[V]).key)
		return true
	})

	sort.Slice(keys, func(i, j int) bool {
		return bytes.Compare(keys[i], keys[j]) < 0
	})

	idx.cacheAllKeysAsc = keys
	idx.cacheVersion = ver
	return keys
}

func reverseCopyKeys(keys []Key) []Key {
	out := make([]Key, len(keys))
	for i, k := range keys {
		out[len(keys)-1-i] = k
	}
	return out
}
