package blockstm

import (
	"bytes"
	"sync"
	"sync/atomic"

	tree2 "github.com/cosmos/cosmos-sdk/internal/blockstm/tree"
)

const mvIndexShards = 64

// minSnapshotKeysCap is the minimum capacity used when building a key snapshot.
const minSnapshotKeysCap = 1024

type mvIndexEntry[V any] struct {
	key  []byte
	data *tree2.SmallBTree[secondaryDataItem[V]]
}

// mvIndexKeyEntry is stored in the ordered key-set so iterators can retrieve
// the per-key version tree without an additional hash lookup.
//
// Note: Tree is immutable w.r.t. the pointer itself (created once per key),
// but the underlying *tree.BTree is still mutated by writers.
type mvIndexKeyEntry[V any] struct {
	Key  Key
	Tree *tree2.SmallBTree[secondaryDataItem[V]]
}

type mvIndexShard[V any] struct {
	mu sync.RWMutex
	m  map[uint64][]mvIndexEntry[V]
}

// mvIndex is a sharded index from key -> per-key version tree.
type mvIndex[V any] struct {
	shards [mvIndexShards]mvIndexShard[V]
	// keys is an ordered set of per-key metadata.
	//
	// Storing mvIndexKeyEntry by value avoids an extra allocation per key.
	// Iterators can fetch the per-key version tree without an additional hash lookup.
	keys *tree2.COWBTree[mvIndexKeyEntry[V]]

	// keySetVersion increments when a new key is added; it invalidates the snapshot cache.
	keySetVersion atomic.Uint64

	cacheMu      sync.RWMutex
	cacheVersion uint64
	// cacheAllKeysAsc is an immutable, ascending snapshot of all keys.
	cacheAllKeysAsc []Key
}

func newMVIndex[V any]() *mvIndex[V] {
	idx := &mvIndex[V]{}
	for i := range idx.shards {
		idx.shards[i].m = make(map[uint64][]mvIndexEntry[V])
	}
	idx.keys = tree2.NewCOWBTree(func(a, b mvIndexKeyEntry[V]) bool {
		return bytes.Compare(a.Key, b.Key) < 0
	}, OuterBTreeDegree)
	return idx
}

func (idx *mvIndex[V]) get(key []byte) *tree2.SmallBTree[secondaryDataItem[V]] {
	if idx == nil {
		return nil
	}
	h := hashKey64(key)
	sh := &idx.shards[h&uint64(mvIndexShards-1)]
	sh.mu.RLock()
	entries := sh.m[h]
	for _, entry := range entries {
		if bytes.Equal(entry.key, key) {
			data := entry.data
			sh.mu.RUnlock()
			return data
		}
	}
	sh.mu.RUnlock()
	return nil
}

func (idx *mvIndex[V]) getOrCreate(key []byte) *tree2.SmallBTree[secondaryDataItem[V]] {
	h := hashKey64(key)
	sh := &idx.shards[h&uint64(mvIndexShards-1)]
	sh.mu.Lock()
	entries := sh.m[h]
	for _, entry := range entries {
		if bytes.Equal(entry.key, key) {
			data := entry.data
			sh.mu.Unlock()
			return data
		}
	}

	// Avoid an extra allocation by reusing the provided key slice.
	// Callers must treat keys as immutable once written.
	kCopy := key
	data := tree2.NewSmallBTree(secondaryLesser[V], InnerBTreeDegree)
	sh.m[h] = append(entries, mvIndexEntry[V]{key: kCopy, data: data})
	sh.mu.Unlock()
	idx.keys.Set(mvIndexKeyEntry[V]{Key: kCopy, Tree: data})
	idx.keySetVersion.Add(1)
	return data
}

func (idx *mvIndex[V]) cursorKeys(start, end []byte, ascending bool) keyCursor[V] {
	if idx == nil {
		return noopKeyCursor[V]{}
	}
	// The key-set iterator runs on an immutable snapshot and does not block writers.
	return newBTreeKeyCursor(idx.keys, start, end, ascending)
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
	idx.keys.Scan(func(item mvIndexKeyEntry[V]) bool {
		keys = append(keys, item.Key)
		return true
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

// hashKey64 computes a 64-bit FNV-1a hash.
func hashKey64(b []byte) uint64 {
	const (
		fnv1a64OffsetBasis = 14695981039346656037
		fnv1a64Prime       = 1099511628211
	)
	h := uint64(fnv1a64OffsetBasis)
	for _, c := range b {
		h ^= uint64(c)
		h *= fnv1a64Prime
	}
	return h
}
