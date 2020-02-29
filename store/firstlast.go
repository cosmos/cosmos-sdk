package store

import (
	"bytes"

	tmkv "github.com/tendermint/tendermint/libs/kv"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Gets the first item.
func First(st KVStore, start, end []byte) (kv tmkv.Pair, ok bool) {
	iter := st.Iterator(start, end)
	if !iter.Valid() {
		return kv, false
	}
	defer iter.Close()

	return tmkv.Pair{Key: iter.Key(), Value: iter.Value()}, true
}

// Gets the last item.  `end` is exclusive.
func Last(st KVStore, start, end []byte) (kv tmkv.Pair, ok bool) {
	iter := st.ReverseIterator(end, start)
	if !iter.Valid() {
		if v := st.Get(start); v != nil {
			return tmkv.Pair{Key: sdk.CopyBytes(start), Value: sdk.CopyBytes(v)}, true
		}
		return kv, false
	}
	defer iter.Close()

	if bytes.Equal(iter.Key(), end) {
		// Skip this one, end is exclusive.
		iter.Next()
		if !iter.Valid() {
			return kv, false
		}
	}

	return tmkv.Pair{Key: iter.Key(), Value: iter.Value()}, true
}
