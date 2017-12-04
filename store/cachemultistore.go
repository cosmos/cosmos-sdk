package store

import dbm "github.com/tendermint/tmlibs/db"

//----------------------------------------
// cacheMultiStore

// cacheMultiStore holds many cache-wrapped stores.
// Implements MultiStore.
type cacheMultiStore struct {
	db           dbm.CacheDB
	curVersion   int64
	lastCommitID CommitID
	substores    map[string]CacheWriter
}

func newCacheMultiStoreFromRMS(rms *rootMultiStore) cacheMultiStore {
	cms := cacheMultiStore{
		db:           rms.db.CacheDB(),
		curVersion:   rms.curVersion,
		lastCommitID: rms.lastCommitID,
		substores:    make(map[string]CacheWriter, len(rms.substores)),
	}
	for name, substore := range rms.substores {
		cms.substores[name] = substore.CacheWrap().(CacheWriter)
	}
	return cms
}

func newCacheMultiStoreFromCMS(cms cacheMultiStore) cacheMultiStore {
	cms := cacheMultiStore{
		db:           cms.db.CacheDB(),
		curVersion:   cms.curVersion,
		lastCommitID: cms.lastCommitID,
		substores:    make(map[string]CacheWriter, len(cms.substores)),
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
	return cms.curVersion
}

// Implements CacheMultiStore
func (cms cacheMultiStore) Write() {
	cms.db.Write()
	for _, substore := range cms.substores {
		substore.Write()
	}
}

// Implements CacheMultiStore
func (cms cacheMultiStore) CacheMultiStore() CacheMultiStore {
	return newCacheMultiStoreFromCMS(cms)
}

// Implements CacheMultiStore
func (cms cacheMultiStore) GetCommitter(name string) Committer {
	return cms.store[name]
}

// Implements CacheMultiStore
func (cms cacheMultiStore) GetKVStore(name string) KVStore {
	return cms.store[name].(KVStore)
}

// Implements CacheMultiStore
func (cms cacheMultiStore) GetIterKVStore(name string) IterKVStore {
	return cms.store[name].(IterKVStore)
}
