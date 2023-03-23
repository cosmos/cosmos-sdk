package mock

import (
	"io"

	dbm "github.com/cosmos/cosmos-db"
	protoio "github.com/cosmos/gogoproto/io"

	"cosmossdk.io/store/metrics"
	pruningtypes "cosmossdk.io/store/pruning/types"
	snapshottypes "cosmossdk.io/store/snapshots/types"
	storetypes "cosmossdk.io/store/types"
)

var _ storetypes.MultiStore = multiStore{}

type multiStore struct {
	kv map[storetypes.StoreKey]kvStore
}

func (ms multiStore) CacheMultiStore() storetypes.CacheMultiStore {
	panic("not implemented")
}

func (ms multiStore) CacheMultiStoreWithVersion(_ int64) (storetypes.CacheMultiStore, error) {
	panic("not implemented")
}

func (ms multiStore) CacheWrap() storetypes.CacheWrap {
	panic("not implemented")
}

func (ms multiStore) CacheWrapWithTrace(_ io.Writer, _ storetypes.TraceContext) storetypes.CacheWrap {
	panic("not implemented")
}

func (ms multiStore) CacheWrapWithListeners(_ storetypes.StoreKey, _ []storetypes.MemoryListener) storetypes.CacheWrap {
	panic("not implemented")
}

func (ms multiStore) TracingEnabled() bool {
	panic("not implemented")
}

func (ms multiStore) SetTracingContext(_ storetypes.TraceContext) storetypes.MultiStore {
	panic("not implemented")
}

func (ms multiStore) SetTracer(_ io.Writer) storetypes.MultiStore {
	panic("not implemented")
}

func (ms multiStore) AddListeners(_ []storetypes.StoreKey) {
	panic("not implemented")
}

func (ms multiStore) SetMetrics(metrics.StoreMetrics) {
	panic("not implemented")
}

func (ms multiStore) ListeningEnabled(_ storetypes.StoreKey) bool {
	panic("not implemented")
}

func (ms multiStore) PopStateCache() []*storetypes.StoreKVPair {
	panic("not implemented")
}

func (ms multiStore) Commit() storetypes.CommitID {
	panic("not implemented")
}

func (ms multiStore) LastCommitID() storetypes.CommitID {
	panic("not implemented")
}

func (ms multiStore) SetPruning(_ pruningtypes.PruningOptions) {
	panic("not implemented")
}

func (ms multiStore) GetPruning() pruningtypes.PruningOptions {
	panic("not implemented")
}

func (ms multiStore) GetCommitKVStore(_ storetypes.StoreKey) storetypes.CommitKVStore {
	panic("not implemented")
}

func (ms multiStore) GetCommitStore(_ storetypes.StoreKey) storetypes.CommitStore {
	panic("not implemented")
}

func (ms multiStore) MountStoreWithDB(key storetypes.StoreKey, _ storetypes.StoreType, _ dbm.DB) {
	ms.kv[key] = kvStore{store: make(map[string][]byte)}
}

func (ms multiStore) LoadLatestVersion() error {
	return nil
}

func (ms multiStore) LoadLatestVersionAndUpgrade(_ *storetypes.StoreUpgrades) error {
	return nil
}

func (ms multiStore) LoadVersionAndUpgrade(_ int64, _ *storetypes.StoreUpgrades) error {
	panic("not implemented")
}

func (ms multiStore) LoadVersion(_ int64) error {
	panic("not implemented")
}

func (ms multiStore) GetKVStore(key storetypes.StoreKey) storetypes.KVStore {
	return ms.kv[key]
}

func (ms multiStore) GetStore(_ storetypes.StoreKey) storetypes.Store {
	panic("not implemented")
}

func (ms multiStore) GetStoreType() storetypes.StoreType {
	panic("not implemented")
}

func (ms multiStore) PruneSnapshotHeight(_ int64) {
	panic("not implemented")
}

func (ms multiStore) SetSnapshotInterval(_ uint64) {
	panic("not implemented")
}

func (ms multiStore) SetInterBlockCache(_ storetypes.MultiStorePersistentCache) {
	panic("not implemented")
}

func (ms multiStore) SetIAVLCacheSize(_ int) {
	panic("not implemented")
}

func (ms multiStore) SetIAVLDisableFastNode(_ bool) {
	panic("not implemented")
}

func (ms multiStore) SetLazyLoading(bool) {
	panic("not implemented")
}

func (ms multiStore) SetInitialVersion(_ int64) error {
	panic("not implemented")
}

func (ms multiStore) Snapshot(_ uint64, _ protoio.Writer) error {
	panic("not implemented")
}

func (ms multiStore) Restore(
	_ uint64, _ uint32, _ protoio.Reader,
) (snapshottypes.SnapshotItem, error) {
	panic("not implemented")
}

func (ms multiStore) RollbackToVersion(_ int64) error {
	panic("not implemented")
}

func (ms multiStore) LatestVersion() int64 {
	panic("not implemented")
}

var _ storetypes.KVStore = kvStore{}

type kvStore struct {
	store map[string][]byte
}

func (kv kvStore) CacheWrap() storetypes.CacheWrap {
	panic("not implemented")
}

func (kv kvStore) CacheWrapWithTrace(_ io.Writer, _ storetypes.TraceContext) storetypes.CacheWrap {
	panic("not implemented")
}

func (kv kvStore) CacheWrapWithListeners(_ storetypes.StoreKey, _ []storetypes.MemoryListener) storetypes.CacheWrap {
	panic("not implemented")
}

func (kv kvStore) GetStoreType() storetypes.StoreType {
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
	storetypes.AssertValidKey(key)
	kv.store[string(key)] = value
}

func (kv kvStore) Delete(key []byte) {
	delete(kv.store, string(key))
}

func (kv kvStore) Prefix(_ []byte) storetypes.KVStore {
	panic("not implemented")
}

func (kv kvStore) Gas(_ storetypes.GasMeter, _ storetypes.GasConfig) storetypes.KVStore {
	panic("not implmeneted")
}

func (kv kvStore) Iterator(_, _ []byte) storetypes.Iterator {
	panic("not implemented")
}

func (kv kvStore) ReverseIterator(_, _ []byte) storetypes.Iterator {
	panic("not implemented")
}

func (kv kvStore) SubspaceIterator(_ []byte) storetypes.Iterator {
	panic("not implemented")
}

func (kv kvStore) ReverseSubspaceIterator(_ []byte) storetypes.Iterator {
	panic("not implemented")
}

func NewCommitMultiStore() storetypes.CommitMultiStore {
	return multiStore{kv: make(map[storetypes.StoreKey]kvStore)}
}
