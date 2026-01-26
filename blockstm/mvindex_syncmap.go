//go:build syncmap

package blockstm

import (
	"bytes"
	"sync"

	"github.com/cosmos/cosmos-sdk/blockstm/tree"
)

// mvIndexSyncMapValue holds the stable key bytes and the per-key version tree.
type mvIndexSyncMapValue[V any] struct {
	key  []byte
	data *tree.BTree[secondaryDataItem[V]]
}

type mvIndexSyncMapBucket[V any] struct {
	mu      sync.RWMutex
	entries []mvIndexSyncMapValue[V]
}

// mvIndex is an index from key -> per-key version tree, implemented with sync.Map.
type mvIndex[V any] struct {
	m sync.Map // map[uint64]*mvIndexSyncMapBucket[V]
	mvIndexCache
}

func newMVIndex[V any]() *mvIndex[V] {
	return &mvIndex[V]{}
}

func (idx *mvIndex[V]) get(key []byte) *tree.BTree[secondaryDataItem[V]] {
	if idx == nil {
		return nil
	}
	h := hashKey64(key)
	v, ok := idx.m.Load(h)
	if !ok {
		return nil
	}
	b := v.(*mvIndexSyncMapBucket[V])
	b.mu.RLock()
	for _, entry := range b.entries {
		if bytes.Equal(entry.key, key) {
			data := entry.data
			b.mu.RUnlock()
			return data
		}
	}
	b.mu.RUnlock()
	return nil
}

func (idx *mvIndex[V]) getOrCreate(key []byte) *tree.BTree[secondaryDataItem[V]] {
	h := hashKey64(key)
	bAny, _ := idx.m.LoadOrStore(h, &mvIndexSyncMapBucket[V]{})
	b := bAny.(*mvIndexSyncMapBucket[V])

	b.mu.Lock()
	for _, entry := range b.entries {
		if bytes.Equal(entry.key, key) {
			data := entry.data
			b.mu.Unlock()
			return data
		}
	}

	kCopy := append([]byte(nil), key...)
	data := tree.NewBTree(secondaryLesser[V], InnerBTreeDegree)
	b.entries = append(b.entries, mvIndexSyncMapValue[V]{key: kCopy, data: data})
	b.mu.Unlock()

	idx.keySetVersion.Add(1)
	return data
}

func (idx *mvIndex[V]) snapshotKeys(start, end []byte, ascending bool) []Key {
	if idx == nil {
		return nil
	}
	return sliceKeys(idx.snapshotAllKeysAsc(), start, end, ascending)
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
	return idx.mvIndexCache.snapshotAllKeysAsc(func(capHint int) []Key {
		keys := make([]Key, 0, capHint)
		idx.m.Range(func(_, v any) bool {
			b := v.(*mvIndexSyncMapBucket[V])
			b.mu.RLock()
			for _, entry := range b.entries {
				keys = append(keys, entry.key)
			}
			b.mu.RUnlock()
			return true
		})
		return keys
	})
}
