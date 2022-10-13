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
	return resp, err
}

// GetAndDecodeWithBool gets and decodes key and returns it. Returns false if key doesn't exist.
func GetAndDecodeWithBool[T any](store types.KVStore, dec func([]byte) (T, bool), key []byte) (T, bool) {
	var res T
	bz := store.Get(key)
	if len(bz) == 0 {
		return res, false
	}
	resp, boolval := dec(bz)
	return resp, boolval
}

// StoreAPI is a wrapper around the store's KVStore to provide more safe key management and better ease-of-use.
type StoreAPI struct {
	types.KVStore
}

// NewStoreAPI returns a new KVStore.
func NewStoreAPI(store types.KVStore) StoreAPI {
	return StoreAPI{
		KVStore: store,
	}
}

// Set stores the value under the given key.
func (store StoreAPI) Set(key []byte, value []byte) {
	store.KVStore.Set(key, value)
}

// Get returns the raw bytes stored under the given key. Returns nil when key does not exist.
func (store StoreAPI) Get(key []byte) []byte {
	return store.KVStore.Get(key)
}

// Delete deletes the value stored under the given key, if it exists.
func (store StoreAPI) Delete(key []byte) {
	store.KVStore.Delete(key)
}
