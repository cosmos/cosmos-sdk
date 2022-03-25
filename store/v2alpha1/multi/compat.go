package multi

import (
	"fmt"
	"io"

	tmdb "github.com/tendermint/tm-db"

	v1 "github.com/cosmos/cosmos-sdk/store/types"
	v2 "github.com/cosmos/cosmos-sdk/store/v2"
)

var (
	_ v1.CommitMultiStore = (*compatStore)(nil)
	_ v1.Queryable        = (*compatStore)(nil)
	_ v1.CacheMultiStore  = (*compatCacheStore)(nil)
)

type compatStore struct {
	*Store
}

type compatCacheStore struct {
	*cacheStore
}

func WrapStoreAsV1CommitMultiStore(s v2.CommitMultiStore) (v1.CommitMultiStore, error) {
	impl, ok := s.(*Store)
	if !ok {
		return nil, fmt.Errorf("cannot wrap as v1.CommitMultiStore: %T", s)
	}
	return &compatStore{impl}, nil
}

func WrapCacheStoreAsV1CacheMultiStore(cs v2.CacheMultiStore) (v1.CacheMultiStore, error) {
	impl, ok := cs.(*cacheStore)
	if !ok {
		return nil, fmt.Errorf("cannot wrap as v1.CacheMultiStore: %T", cs)
	}
	return &compatCacheStore{impl}, nil
}

// commit store

func (st *compatStore) GetStoreType() v1.StoreType {
	return v1.StoreTypeMulti
}

func (st *compatStore) CacheWrap() v1.CacheWrap {
	return st.CacheMultiStore()
}

// TODO: v1 MultiStore ignores args, do we as well?
func (st *compatStore) CacheWrapWithTrace(io.Writer, v1.TraceContext) v1.CacheWrap {
	return st.CacheWrap()
}
func (st *compatStore) CacheWrapWithListeners(v1.StoreKey, []v1.WriteListener) v1.CacheWrap {
	return st.CacheWrap()
}

func (st *compatStore) CacheMultiStore() v1.CacheMultiStore {
	return &compatCacheStore{newCacheStore(st.Store)}
}
func (st *compatStore) CacheMultiStoreWithVersion(version int64) (v1.CacheMultiStore, error) {
	view, err := st.GetVersion(version)
	if err != nil {
		return nil, err
	}
	return &compatCacheStore{newCacheStore(view)}, nil
}

func (st *compatStore) GetStore(k v1.StoreKey) v1.Store {
	return st.GetKVStore(k)
}

func (st *compatStore) GetCommitStore(key v1.StoreKey) v1.CommitStore {
	panic("unsupported: GetCommitStore")
}
func (st *compatStore) GetCommitKVStore(key v1.StoreKey) v1.CommitKVStore {
	panic("unsupported: GetCommitKVStore")
}

func (st *compatStore) SetTracer(w io.Writer) v1.MultiStore {
	st.Store.SetTracer(w)
	return st
}
func (st *compatStore) SetTracingContext(tc v1.TraceContext) v1.MultiStore {
	st.Store.SetTracingContext(tc)
	return st
}

func (st *compatStore) MountStoreWithDB(key v1.StoreKey, typ v1.StoreType, db tmdb.DB) {
	panic("unsupported: MountStoreWithDB")
}

func (st *compatStore) LoadLatestVersion() error {
	return nil // this store is always at the latest version
}
func (st *compatStore) LoadLatestVersionAndUpgrade(upgrades *v1.StoreUpgrades) error {
	panic("unsupported: LoadLatestVersionAndUpgrade")
}
func (st *compatStore) LoadVersionAndUpgrade(ver int64, upgrades *v1.StoreUpgrades) error {
	panic("unsupported: LoadLatestVersionAndUpgrade")
}

func (st *compatStore) LoadVersion(ver int64) error {
	// TODO
	// cache a viewStore representing "current" version?
	panic("unsupported: LoadVersion")
}

func (st *compatStore) SetInterBlockCache(v1.MultiStorePersistentCache) {
	panic("unsupported: SetInterBlockCache")
}
func (st *compatStore) SetInitialVersion(version int64) error {
	if version < 0 {
		return fmt.Errorf("invalid version")
	}
	return st.Store.SetInitialVersion(uint64(version))
}
func (st *compatStore) SetIAVLCacheSize(size int) {
	panic("unsupported: SetIAVLCacheSize")
}

// cache store

func (cs *compatCacheStore) GetStoreType() v1.StoreType { return v1.StoreTypeMulti }
func (cs *compatCacheStore) CacheWrap() v1.CacheWrap {
	return cs.CacheMultiStore()
}
func (cs *compatCacheStore) CacheWrapWithTrace(w io.Writer, tc v1.TraceContext) v1.CacheWrap {
	return cs.CacheWrap()
}
func (cs *compatCacheStore) CacheWrapWithListeners(storeKey v1.StoreKey, listeners []v1.WriteListener) v1.CacheWrap {
	return cs.CacheWrap()
}
func (cs *compatCacheStore) CacheMultiStore() v1.CacheMultiStore {
	return &compatCacheStore{newCacheStore(cs.cacheStore)}
}
func (cs *compatCacheStore) CacheMultiStoreWithVersion(int64) (v1.CacheMultiStore, error) {
	return nil, fmt.Errorf("cannot branch cached multi-store with a version")
}

func (cs *compatCacheStore) GetStore(k v1.StoreKey) v1.Store { return cs.GetKVStore(k) }

func (cs *compatCacheStore) SetTracer(w io.Writer) v1.MultiStore {
	cs.cacheStore.SetTracer(w)
	return cs
}
func (cs *compatCacheStore) SetTracingContext(tc v1.TraceContext) v1.MultiStore {
	cs.cacheStore.SetTracingContext(tc)
	return cs
}
