package mock

import (
	"io"

	protoio "github.com/gogo/protobuf/io"

	snapshottypes "github.com/cosmos/cosmos-sdk/snapshots/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.CommitMultiStore = multiStore{}

type multiStore struct {
	kv map[storetypes.StoreKey]kvStore
}

var _ sdk.KVStore = kvStore{}

type kvStore struct {
	store map[string][]byte
}

type MultiStoreConfig = []storetypes.StoreKey

func NewCommitMultiStore(c MultiStoreConfig) sdk.CommitMultiStore {
	stores := make(map[storetypes.StoreKey]kvStore)
	for _, skey := range c {
		stores[skey] = kvStore{map[string][]byte{}}
	}
	return multiStore{kv: stores}
}

func (ms multiStore) CacheWrap() sdk.CacheMultiStore {
	panic("not implemented")
}

func (ms multiStore) TracingEnabled() bool {
	panic("not implemented")
}

func (ms multiStore) SetTracingContext(tc sdk.TraceContext) {
	panic("not implemented")
}

func (ms multiStore) SetTracer(w io.Writer) {
	panic("not implemented")
}

func (ms multiStore) AddListeners(key storetypes.StoreKey, listeners []storetypes.WriteListener) {
	panic("not implemented")
}

func (ms multiStore) ListeningEnabled(key storetypes.StoreKey) bool {
	panic("not implemented")
}

func (ms multiStore) Commit() storetypes.CommitID {
	panic("not implemented")
}

func (ms multiStore) LastCommitID() storetypes.CommitID {
	panic("not implemented")
}

func (ms multiStore) SetPruning(opts sdk.PruningOptions) {
	panic("not implemented")
}

func (ms multiStore) GetPruning() sdk.PruningOptions {
	panic("not implemented")
}

func (ms multiStore) GetKVStore(key storetypes.StoreKey) sdk.KVStore {
	return ms.kv[key]
}

func (ms multiStore) SetInitialVersion(version uint64) error {
	panic("not implemented")
}

func (ms multiStore) GetVersion(int64) (storetypes.MultiStore, error) {
	panic("not implemented")
}

func (ms multiStore) GetAllVersions() []int {
	panic("not implemented")
}

func (ms multiStore) Close() error {
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

func (kv kvStore) CacheWrap() storetypes.CacheWrap {
	panic("not implemented")
}

func (kv kvStore) CacheWrapWithTrace(w io.Writer, tc sdk.TraceContext) storetypes.CacheWrap {
	panic("not implemented")
}

func (kv kvStore) CacheWrapWithListeners(_ storetypes.StoreKey, _ []storetypes.WriteListener) storetypes.CacheWrap {
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

func (kv kvStore) Iterator(start, end []byte) sdk.Iterator {
	panic("not implemented")
}

func (kv kvStore) ReverseIterator(start, end []byte) sdk.Iterator {
	panic("not implemented")
}
