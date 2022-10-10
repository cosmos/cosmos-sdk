package store

import (
	types "github.com/cosmos/cosmos-sdk/store/types"
)

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

func GetAndDecodeWithBool[T any](store types.KVStore, dec func([]byte) (T, bool), key []byte) (T, bool) {
	var res T
	bz := store.Get(key)
	if len(bz) == 0 {
		return res, false
	}
	resp, boolval := dec(bz)
	return resp, boolval
}

type StoreAPI struct {
	types.KVStore
}

func NewStoreAPI(store types.KVStore) StoreAPI {
	return StoreAPI{
		KVStore: store,
	}
}

func (store StoreAPI) Set(key []byte, value []byte) {
	store.KVStore.Set(key, value)
}

func (store StoreAPI) Get(key []byte) []byte {
	return store.KVStore.Get(key)
}

func (store StoreAPI) Delete(key []byte) {
	store.KVStore.Delete(key)
}
