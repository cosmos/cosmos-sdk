package mock

import (
	dbm "github.com/tendermint/tmlibs/db"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type multiStore struct {
	kv map[sdk.StoreKey]kvStore
}

func (ms multiStore) CacheMultiStore() sdk.CacheMultiStore {
	panic("not implemented")
}

func (ms multiStore) CacheWrap() sdk.CacheWrap {
	panic("not implemented")
}

func (ms multiStore) Commit() sdk.CommitID {
	panic("not implemented")
}

func (ms multiStore) LastCommitID() sdk.CommitID {
	panic("not implemented")
}

func (ms multiStore) GetCommitKVStore(key sdk.StoreKey) sdk.CommitKVStore {
	panic("not implemented")
}

func (ms multiStore) GetCommitStore(key sdk.StoreKey) sdk.CommitStore {
	panic("not implemented")
}

func (ms multiStore) MountStoreWithDB(key sdk.StoreKey, typ sdk.StoreType, db dbm.DB) {
	ms.kv[key] = kvStore{store: make(map[string][]byte)}
}

func (ms multiStore) LoadLatestVersion() error {
	return nil
}

func (ms multiStore) LoadVersion(ver int64) error {
	panic("not implemented")
}

func (ms multiStore) GetKVStore(key sdk.StoreKey) sdk.KVStore {
	return ms.kv[key]
}

func (ms multiStore) GetKVStoreWithGas(meter sdk.GasMeter, key sdk.StoreKey) sdk.KVStore {
	panic("not implemented")
}

func (ms multiStore) GetStore(key sdk.StoreKey) sdk.Store {
	panic("not implemented")
}

func (ms multiStore) GetStoreType() sdk.StoreType {
	panic("not implemented")
}

type kvStore struct {
	store map[string][]byte
}

func (kv kvStore) CacheWrap() sdk.CacheWrap {
	panic("not implemented")
}

func (kv kvStore) GetStoreType() sdk.StoreType {
	panic("not implemented")
}

func (kv kvStore) Get(key []byte) []byte {
	v, ok := kv.store[string(key)]
	if !ok {
		return nil
	}
	return v
}

func (kv kvStore) Has(key []byte) bool {
	_, ok := kv.store[string(key)]
	return ok
}

func (kv kvStore) Set(key, value []byte) {
	kv.store[string(key)] = value
}

func (kv kvStore) Delete(key []byte) {
	delete(kv.store, string(key))
}

func (kv kvStore) Prefix(prefix []byte) sdk.KVStore {
	panic("not implemented")
}

func (kv kvStore) Iterator(start, end []byte) sdk.Iterator {
	panic("not implemented")
}

func (kv kvStore) ReverseIterator(start, end []byte) sdk.Iterator {
	panic("not implemented")
}

func (kv kvStore) SubspaceIterator(prefix []byte) sdk.Iterator {
	panic("not implemented")
}

func (kv kvStore) ReverseSubspaceIterator(prefix []byte) sdk.Iterator {
	panic("not implemented")
}

func NewCommitMultiStore(db dbm.DB) sdk.CommitMultiStore {
	return multiStore{kv: make(map[sdk.StoreKey]kvStore)}
}
