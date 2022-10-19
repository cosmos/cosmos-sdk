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
	
	resp, ok := dec(bz)
	return resp, ok
}

// KVStoreWrapper is a wrapper around the store's KVStore to provide more safe key management and better ease-of-use.
type KVStoreWrapper struct {
	types.KVStore
}

// NewKVStoreWrapper returns a new KVStore.
func NewKVStoreWrapper(store types.KVStore) KVStoreWrapper {
	return KVStoreWrapper{
		KVStore: store,
	}
}

// Set stores the value under the given key.
func (store KVStoreWrapper) Set(key []byte, value []byte) {
	store.KVStore.Set(key, value)
}

// Get returns the raw bytes stored under the given key. Returns nil when key does not exist.
func (store KVStoreWrapper) Get(key []byte) []byte {
	return store.KVStore.Get(key)
}

// Delete deletes the value stored under the given key, if it exists.
func (store KVStoreWrapper) Delete(key []byte) {
	store.KVStore.Delete(key)
}
