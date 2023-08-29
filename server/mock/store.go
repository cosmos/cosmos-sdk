package mock

import (
	"io"

	protoio "github.com/gogo/protobuf/io"
	dbm "github.com/tendermint/tm-db"

	pruningtypes "github.com/cosmos/cosmos-sdk/pruning/types"
	snapshottypes "github.com/cosmos/cosmos-sdk/snapshots/types"
	store "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.MultiStore = multiStore{}

type multiStore struct {
	kv map[sdk.StoreKey]kvStore
}

func (ms multiStore) CacheMultiStore() sdk.CacheMultiStore {
	panic("not implemented")
}

func (ms multiStore) CacheMultiStoreWithVersion(_ int64) (sdk.CacheMultiStore, error) {
	panic("not implemented")
}

func (ms multiStore) CacheWrap() sdk.CacheWrap {
	panic("not implemented")
}

func (ms multiStore) CacheWrapWithTrace(_ io.Writer, _ sdk.TraceContext) sdk.CacheWrap {
	panic("not implemented")
}

func (ms multiStore) CacheWrapWithListeners(_ store.StoreKey, _ []store.WriteListener) store.CacheWrap {
	panic("not implemented")
}

func (ms multiStore) TracingEnabled() bool {
	panic("not implemented")
}

func (ms multiStore) SetTracingContext(tc sdk.TraceContext) sdk.MultiStore {
	panic("not implemented")
}

func (ms multiStore) SetTracer(w io.Writer) sdk.MultiStore {
	panic("not implemented")
}

func (ms multiStore) AddListeners(key store.StoreKey, listeners []store.WriteListener) {
	panic("not implemented")
}

func (ms multiStore) ListeningEnabled(key store.StoreKey) bool {
	panic("not implemented")
}

func (ms multiStore) Commit() sdk.CommitID {
	panic("not implemented")
}

func (ms multiStore) LastCommitID() sdk.CommitID {
	panic("not implemented")
}

func (ms multiStore) SetPruning(opts pruningtypes.PruningOptions) {
	panic("not implemented")
}

func (ms multiStore) GetPruning() pruningtypes.PruningOptions {
	panic("not implemented")
}

func (ms multiStore) GetCommitKVStore(key sdk.StoreKey) sdk.CommitKVStore {
	panic("not implemented")
}

func (ms multiStore) GetCommitInfoFromDB(ver int64) (*store.CommitInfo, error) {
	panic("not implemented")
}

func (ms multiStore) GetCommitStore(key sdk.StoreKey) sdk.CommitStore {
	panic("not implemented")
}

func (ms multiStore) MountStoreWithDB(key store.StoreKey, typ store.StoreType, db dbm.DB) {
	ms.kv[key] = kvStore{store: make(map[string][]byte)}
}

func (ms multiStore) LoadLatestVersion() error {
	return nil
}

func (ms multiStore) LoadLatestVersionAndUpgrade(upgrades *store.StoreUpgrades) error {
	return nil
}

func (ms multiStore) LoadVersionAndUpgrade(ver int64, upgrades *store.StoreUpgrades) error {
	panic("not implemented")
}

func (ms multiStore) LoadVersion(ver int64) error {
	panic("not implemented")
}

func (ms multiStore) GetKVStore(key sdk.StoreKey) sdk.KVStore {
	return ms.kv[key]
}

func (ms multiStore) GetStore(key sdk.StoreKey) sdk.Store {
	panic("not implemented")
}

func (ms multiStore) GetStoreType() sdk.StoreType {
	panic("not implemented")
}

func (ms multiStore) PruneSnapshotHeight(height int64) {
	panic("not implemented")
}

func (ms multiStore) SetInterBlockCache(_ sdk.MultiStorePersistentCache) {
	panic("not implemented")
}
func (ms multiStore) SetIAVLCacheSize(size int) {
	panic("not implemented")
}

func (ms multiStore) SetInitialVersion(version int64) error {
	panic("not implemented")
}

func (ms multiStore) SetSnapshotInterval(snapshotInterval uint64) {
	panic("not implemented")
}

func (ms multiStore) Snapshot(height uint64, protoWriter protoio.Writer) error {
	panic("not implemented")
}

func (ms multiStore) Restore(
	height uint64, format uint32, protoReader protoio.Reader,
) (snapshottypes.SnapshotItem, error) {
	panic("not implemented")
}

func (ms multiStore) SetAppVersion(_ uint64) error {
	panic("not implemented")
}

func (ms multiStore) GetAppVersion() (uint64, error) {
	panic("not implemented")
}

var _ sdk.KVStore = kvStore{}

type kvStore struct {
	store map[string][]byte
}

func (kv kvStore) CacheWrap() sdk.CacheWrap {
	panic("not implemented")
}

func (kv kvStore) CacheWrapWithTrace(w io.Writer, tc sdk.TraceContext) sdk.CacheWrap {
	panic("not implemented")
}

func (kv kvStore) CacheWrapWithListeners(_ store.StoreKey, _ []store.WriteListener) store.CacheWrap {
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
	store.AssertValidKey(key)
	kv.store[string(key)] = value
}

func (kv kvStore) Delete(key []byte) {
	delete(kv.store, string(key))
}

func (kv kvStore) Prefix(prefix []byte) sdk.KVStore {
	panic("not implemented")
}

func (kv kvStore) Gas(meter sdk.GasMeter, config sdk.GasConfig) sdk.KVStore {
	panic("not implmeneted")
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

func NewCommitMultiStore() sdk.CommitMultiStore {
	return multiStore{kv: make(map[sdk.StoreKey]kvStore)}
}
