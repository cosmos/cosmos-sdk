//go:build !syncmap

package blockstm

import (
	"bytes"
	"sync"

	"github.com/cosmos/cosmos-sdk/blockstm/tree"
)

const mvIndexShards = 64

type mvIndexEntry[V any] struct {
	key  []byte
	data *tree.BTree[secondaryDataItem[V]]
}

type mvIndexShard[V any] struct {
	mu sync.RWMutex
	m  map[uint64][]mvIndexEntry[V]
}

// mvIndex is a sharded index from key -> per-key version tree.
// Keys are copied on first insert to keep them stable.
type mvIndex[V any] struct {
	shards [mvIndexShards]mvIndexShard[V]
	mvIndexCache
}

func newMVIndex[V any]() *mvIndex[V] {
	idx := &mvIndex[V]{}
	for i := range idx.shards {
		idx.shards[i].m = make(map[uint64][]mvIndexEntry[V])
	}
	return idx
}

func (idx *mvIndex[V]) get(key []byte) *tree.BTree[secondaryDataItem[V]] {
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

func (idx *mvIndex[V]) getOrCreate(key []byte) *tree.BTree[secondaryDataItem[V]] {
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
	data := tree.NewBTree(secondaryLesser[V], InnerBTreeDegree)
	sh.m[h] = append(entries, mvIndexEntry[V]{key: kCopy, data: data})
	sh.mu.Unlock()
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
		return keys
	})
}
