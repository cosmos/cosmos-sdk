package blockstm

import (
	"bytes"
	"sort"
	"sync"
	"sync/atomic"
)

// minSnapshotKeysCap is the minimum capacity used when building a key snapshot.
const minSnapshotKeysCap = 1024

// mvIndexCache holds the key-snapshot cache shared by mvIndex implementations.
type mvIndexCache struct {
	// keySetVersion increments when a new key is added; it invalidates the snapshot cache.
	keySetVersion atomic.Uint64

	cacheMu      sync.RWMutex
	cacheVersion uint64
	// cacheAllKeysAsc is an immutable, ascending snapshot of all keys.
	cacheAllKeysAsc []Key
}

// snapshotAllKeysAsc returns an immutable, ascending snapshot of all keys.
//
// gather should return an unsorted list of all keys, using capHint as a suggested capacity.
func (c *mvIndexCache) snapshotAllKeysAsc(gather func(capHint int) []Key) []Key {
	ver := c.keySetVersion.Load()
	c.cacheMu.RLock()
	if c.cacheVersion == ver {
		keys := c.cacheAllKeysAsc
		c.cacheMu.RUnlock()
		return keys
	}
	c.cacheMu.RUnlock()

	c.cacheMu.Lock()
	defer c.cacheMu.Unlock()
	ver = c.keySetVersion.Load()
	if c.cacheVersion == ver {
		return c.cacheAllKeysAsc
	}

	capHint := len(c.cacheAllKeysAsc)
	if capHint < minSnapshotKeysCap {
		capHint = minSnapshotKeysCap
	}

	keys := gather(capHint)
	sort.Slice(keys, func(i, j int) bool {
		return bytes.Compare(keys[i], keys[j]) < 0
	})

	c.cacheAllKeysAsc = keys
	c.cacheVersion = ver
	return keys
}

func sliceKeys(keysAsc []Key, start, end []byte, ascending bool) []Key {
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
