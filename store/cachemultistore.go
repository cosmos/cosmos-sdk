package store

//----------------------------------------
// cacheMultiStore

// cacheMultiStore holds many cache-wrapped stores.
// Implements MultiStore.
type cacheMultiStore struct {
	db           CacheKVStore
	nextVersion  int64
	lastCommitID CommitID
	substores    map[string]CacheWrap
}

func newCacheMultiStoreFromRMS(rms *rootMultiStore) cacheMultiStore {
	cms := cacheMultiStore{
		db:           NewCacheKVStore(rms.db),
		nextVersion:  rms.nextVersion,
		lastCommitID: rms.lastCommitID,
		substores:    make(map[string]CacheWrap, len(rms.substores)),
	}
	for name, substore := range rms.substores {
		cms.substores[name] = substore.CacheWrap()
	}
	return cms
}

func newCacheMultiStoreFromCMS(cms cacheMultiStore) cacheMultiStore {
	cms2 := cacheMultiStore{
		db:           NewCacheKVStore(cms.db),
		nextVersion:  cms.nextVersion,
		lastCommitID: cms.lastCommitID,
		substores:    make(map[string]CacheWrap, len(cms.substores)),
	}
	for name, substore := range cms.substores {
		cms2.substores[name] = substore.CacheWrap()
	}
	return cms2
}

// Implements CacheMultiStore
func (cms cacheMultiStore) LastCommitID() CommitID {
	return cms.lastCommitID
}

// Implements CacheMultiStore
func (cms cacheMultiStore) NextVersion() int64 {
	return cms.nextVersion
}

// Implements CacheMultiStore
func (cms cacheMultiStore) Write() {
	cms.db.Write()
	for _, substore := range cms.substores {
		substore.Write()
	}
}

// Implements CacheMultiStore
func (cms cacheMultiStore) CacheWrap() CacheWrap {
	return cms.CacheMultiStore().(CacheWrap)
}

// Implements CacheMultiStore
func (cms cacheMultiStore) CacheMultiStore() CacheMultiStore {
	return newCacheMultiStoreFromCMS(cms)
}

// Implements CacheMultiStore
func (cms cacheMultiStore) GetStore(name string) interface{} {
	return cms.substores[name]
}

// Implements CacheMultiStore
func (cms cacheMultiStore) GetKVStore(name string) KVStore {
	return cms.substores[name].(KVStore)
}
