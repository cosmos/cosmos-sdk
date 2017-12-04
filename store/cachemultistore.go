package store

import dbm "github.com/tendermint/tmlibs/db"

//----------------------------------------
// cacheMultiStore

// cacheMultiStore holds many cache-wrapped stores.
// Implements MultiStore.
type cacheMultiStore struct {
	db           dbm.DB
	version      int64
	lastCommitID CommitID
	substores    map[string]CacheWriter
}

func newCacheMultiStore(rs *rootMultiStore) cacheMultiStore {
	cms := cacheMultiStore{
		db:           dbm.CacheDB(),
		version:      rs.curVersion,
		lastCommitID: rs.lastCommitID,
		substores:    make(map[string]CacheWriter, len(rs.substores)),
	}
	for name, substore := range rs.substores {
		cms.substores[name] = substore.CacheWrap().(CacheWriter)
	}
	return cms
}

// Implements CacheMultiStore
func (cms cacheMultiStore) LastCommitID() CommitID {
	return cms.lastCommitID
}

// Implements CacheMultiStore
func (cms cacheMultiStore) CurrentVersion() int64 {
	return cms.version
}

// Implements CacheMultiStore
func (cms cacheMultiStore) Write() {
	cms.db.Write()
	for substore := range cms.substores {
		substore.Write()
	}
}

// Implements CacheMultiStore
func (rs cacheMultiStore) CacheMultiStore() CacheMultiStore {
	return newCacheMultiStore(rs)
}

// Implements CacheMultiStore
func (rs cacheMultiStore) GetCommitter(name string) Committer {
	return rs.store[name]
}

// Implements CacheMultiStore
func (rs cacheMultiStore) GetKVStore(name string) KVStore {
	return rs.store[name].(KVStore)
}

// Implements CacheMultiStore
func (rs cacheMultiStore) GetIterKVStore(name string) IterKVStore {
	return rs.store[name].(IterKVStore)
}
