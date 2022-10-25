package store

import (
	types "github.com/cosmos/cosmos-sdk/store/types"
)

// GetAndDecode gets and decodes key and returns it. Returns nil if key doesn't exist.
func GetAndDecode[T any](store types.KVStore, dec func([]byte) (T, error), key []byte) (T, error) {
	var res T
	bz := store.Get(key)
	if bz == nil {
		return res, nil
	}

	resp, err := dec(bz)
	if err != nil {
		return resp, err
	}

	return resp, nil
}

// GetAndDecodeWithBool gets and decodes key and returns it. Returns false if key doesn't exist.
func GetAndDecodeWithBool[T any](store types.KVStore, dec func([]byte) (T, bool), key []byte) (T, bool) {
	var res T
	bz := store.Get(key)
	if len(bz) == 0 {
		return res, false
	}

	resp, ok := dec(bz)
	return resp, ok
}
