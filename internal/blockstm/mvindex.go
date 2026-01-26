package blockstm

import (
	"bytes"
	"sort"
	"sync"
	"sync/atomic"

	tree2 "github.com/cosmos/cosmos-sdk/internal/blockstm/tree"
)

const mvIndexShards = 64

// minSnapshotKeysCap is the minimum capacity used when building a key snapshot.
const minSnapshotKeysCap = 1024

type mvIndexEntry[V any] struct {
	key  []byte
	data *tree2.BTree[secondaryDataItem[V]]
}

type mvIndexShard[V any] struct {
	mu sync.RWMutex
	m  map[uint64][]mvIndexEntry[V]
}

// mvIndex is a sharded index from key -> per-key version tree.
// Keys are copied on first insert to keep them stable.
type mvIndex[V any] struct {
	shards [mvIndexShards]mvIndexShard[V]

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
	return idx
}

func (idx *mvIndex[V]) get(key []byte) *tree2.BTree[secondaryDataItem[V]] {
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

func (idx *mvIndex[V]) getOrCreate(key []byte) *tree2.BTree[secondaryDataItem[V]] {
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

	kCopy := append([]byte(nil), key...)
	data := tree2.NewBTree(secondaryLesser[V], InnerBTreeDegree)
	sh.m[h] = append(entries, mvIndexEntry[V]{key: kCopy, data: data})
	sh.mu.Unlock()
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
	for i := range idx.shards {
		sh := &idx.shards[i]
		sh.mu.RLock()
		for _, bucket := range sh.m {
			for _, entry := range bucket {
				keys = append(keys, entry.key)
			}
		}
		sh.mu.RUnlock()
	}

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
