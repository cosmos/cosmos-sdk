package store

import (
	"bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/kv"
)

// Gets the first item.
func First(st KVStore, start, end []byte) (kv.Pair, bool) {
	iter := st.Iterator(start, end)
	if !iter.Valid() {
		return kv.Pair{}, false
	}
	defer iter.Close()

	return kv.Pair{Key: iter.Key(), Value: iter.Value()}, true
}

// Gets the last item.  `end` is exclusive.
func Last(st KVStore, start, end []byte) (kv.Pair, bool) {
	iter := st.ReverseIterator(end, start)
	if !iter.Valid() {
		if v := st.Get(start); v != nil {
			return kv.Pair{Key: sdk.CopyBytes(start), Value: sdk.CopyBytes(v)}, true
		}
		return kv.Pair{}, false
	}
	defer iter.Close()

	if bytes.Equal(iter.Key(), end) {
		// Skip this one, end is exclusive.
		iter.Next()
		if !iter.Valid() {
			return kv.Pair{}, false
		}
	}

	return kv.Pair{Key: iter.Key(), Value: iter.Value()}, true
}
