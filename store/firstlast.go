package store

import "bytes"

// Gets the first item.
func First(st KVStore, start, end []byte) (kv KVPair, ok bool) {
	iter := st.Iterator(start, end)
	if !iter.Valid() {
		return kv, false
	}
	defer iter.Release()

	return KVPair{iter.Key(), iter.Value()}, true
}

// Gets the last item.  `end` is exclusive.
func Last(st KVStore, start, end []byte) (kv KVPair, ok bool) {
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
