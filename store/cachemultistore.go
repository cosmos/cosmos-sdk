package store

import dbm "github.com/tendermint/tmlibs/db"

//----------------------------------------
// cacheMultiStore

// cacheMultiStore holds many cache-wrapped stores.
// Implements MultiStore.
type cacheMultiStore struct {
	db           CacheKVStore
	curVersion   int64
	lastCommitID CommitID
	substores    map[string]CacheWrap
}

func newCacheMultiStoreFromRMS(rms *rootMultiStore) cacheMultiStore {
	cms := cacheMultiStore{
		db:           NewCacheKVStore(rms.db),
		curVersion:   rms.curVersion,
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
		db:           NewCacheKVStore(rms.db),
		curVersion:   cms.curVersion,
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
func (cms cacheMultiStore) CurrentVersion() int64 {
	return cms.curVersion
}

// Implements CacheMultiStore
func (cms cacheMultiStore) Write() error {
	cms.db.Write()
	for _, substore := range cms.substores {
		err := substore.Write()
		if err != nil {
			// NOTE: There is no way to recover from this because we've
			// (possibly) already written to the substore.  What we could do
			// instead is lock all the substores for write so that we know
			// Write() will succeed, but we don't expose a way to ensure
			// consisten versioning across all substores right now, so it
			// wouldn't be correct anwyays.
			//
			// Ergo, lets just require the user of CacheMultiStore to just be
			// aware and careful (e.g. nobody else should Write() to substores
			// except via cacheMultiStore.Write).
			//
			// Another way to deal with this is to wrap substores in something
			// that will override Write() so only cacheMultiStore can do  it.
			panic("Invalid CacheWrap write!")
		}
	}
	return nil
}

// Implements CacheMultiStore
func (cms cacheMultiStore) CacheWrap() CacheWrap {
	return cms.CacheMultiStore()
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
