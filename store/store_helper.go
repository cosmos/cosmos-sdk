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
