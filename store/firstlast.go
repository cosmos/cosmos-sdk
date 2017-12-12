package store

import "bytes"

// Convenience for implemntation of IterKVCache.First using IterKVCache.Iterator
func iteratorFirst(st IterKVStore, start, end []byte) (kv KVPair, ok bool) {
	iter := st.Iterator(start, end)
	if !iter.Valid() {
		return kv, false
	}
	defer iter.Release()
	return KVPair{iter.Key(), iter.Value()}, true
}

// Convenience for implemntation of IterKVCache.Last using IterKVCache.ReverseIterator
func iteratorLast(st IterKVStore, start, end []byte) (kv KVPair, ok bool) {
	iter := st.ReverseIterator(end, start)
	if !iter.Valid() {
		if v := st.Get(start); v != nil {
			return KVPair{cp(start), cp(v)}, true
		} else {
			return kv, false
		}
	}
	defer iter.Release()

	if bytes.Equal(iter.Key(), end) {
		// Skip this one, end is exclusive.
		iter.Next()
		if !iter.Valid() {
			return kv, false
		}
	}

	return KVPair{iter.Key(), iter.Value()}, true
}
