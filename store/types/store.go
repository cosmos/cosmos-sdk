package types

import (
	"fmt"
	"io"

	abci "github.com/tendermint/tendermint/abci/types"
	tmstrings "github.com/tendermint/tendermint/libs/strings"
	dbm "github.com/tendermint/tm-db"

	snapshottypes "github.com/cosmos/cosmos-sdk/snapshots/types"
	"github.com/cosmos/cosmos-sdk/types/kv"
)

type Store interface {
	GetStoreType() StoreType
	CacheWrapper
}

// something that can persist to disk
type Committer interface {
	Commit() CommitID
	LastCommitID() CommitID

	SetPruning(PruningOptions)
	GetPruning() PruningOptions
}

// Stores of MultiStore must implement CommitStore.
type CommitStore interface {
	Committer
	Store
}

// Queryable allows a Store to expose internal state to the abci.Query
// interface. Multistore can route requests to the proper Store.
//
// This is an optional, but useful extension to any CommitStore
type Queryable interface {
	Query(abci.RequestQuery) abci.ResponseQuery
}

//----------------------------------------
// MultiStore

// StoreUpgrades defines a series of transformations to apply the multistore db upon load
type StoreUpgrades struct {
	Added   []string      `json:"added"`
	Renamed []StoreRename `json:"renamed"`
	Deleted []string      `json:"deleted"`
}

// UpgradeInfo defines height and name of the upgrade
// to ensure multistore upgrades happen only at matching height.
type UpgradeInfo struct {
	Name   string `json:"name"`
	Height int64  `json:"height"`
}

// StoreRename defines a name change of a sub-store.
// All data previously under a PrefixStore with OldKey will be copied
// to a PrefixStore with NewKey, then deleted from OldKey store.
type StoreRename struct {
	OldKey string `json:"old_key"`
	NewKey string `json:"new_key"`
}

// IsDeleted returns true if the given key should be added
func (s *StoreUpgrades) IsAdded(key string) bool {
	if s == nil {
		return false
	}
	return tmstrings.StringInSlice(key, s.Added)
}

// IsDeleted returns true if the given key should be deleted
func (s *StoreUpgrades) IsDeleted(key string) bool {
	if s == nil {
		return false
	}
	for _, d := range s.Deleted {
		if d == key {
			return true
		}
	}
	return false
}

// RenamedFrom returns the oldKey if it was renamed
// Returns "" if it was not renamed
func (s *StoreUpgrades) RenamedFrom(key string) string {
	if s == nil {
		return ""
	}
	for _, re := range s.Renamed {
		if re.NewKey == key {
			return re.OldKey
		}
	}
	return ""

}

type MultiStore interface {
	Store

	// Branches MultiStore into a cached storage object.
	// NOTE: Caller should probably not call .Write() on each, but
	// call CacheMultiStore.Write().
	CacheMultiStore() CacheMultiStore

	// CacheMultiStoreWithVersion branches the underlying MultiStore where
	// each stored is loaded at a specific version (height).
	CacheMultiStoreWithVersion(version int64) (CacheMultiStore, error)

	// Convenience for fetching substores.
	// If the store does not exist, panics.
	GetStore(StoreKey) Store
	GetKVStore(StoreKey) KVStore

	// TracingEnabled returns if tracing is enabled for the MultiStore.
	TracingEnabled() bool

	// SetTracer sets the tracer for the MultiStore that the underlying
	// stores will utilize to trace operations. The modified MultiStore is
	// returned.
	SetTracer(w io.Writer) MultiStore

	// SetTracingContext sets the tracing context for a MultiStore. It is
	// implied that the caller should update the context when necessary between
	// tracing operations. The modified MultiStore is returned.
	SetTracingContext(TraceContext) MultiStore
}

// From MultiStore.CacheMultiStore()....
type CacheMultiStore interface {
	MultiStore
	Write() // Writes operations to underlying KVStore
}

// CommitMultiStore is an interface for a MultiStore without cache capabilities.
type CommitMultiStore interface {
	Committer
	MultiStore
	snapshottypes.Snapshotter

	// Mount a store of type using the given db.
	// If db == nil, the new store will use the CommitMultiStore db.
	MountStoreWithDB(key StoreKey, typ StoreType, db dbm.DB)

	// Panics on a nil key.
	GetCommitStore(key StoreKey) CommitStore

	// Panics on a nil key.
	GetCommitKVStore(key StoreKey) CommitKVStore

	// Load the latest persisted version. Called once after all calls to
	// Mount*Store() are complete.
	LoadLatestVersion() error

	// LoadLatestVersionAndUpgrade will load the latest version, but also
	// rename/delete/create sub-store keys, before registering all the keys
	// in order to handle breaking formats in migrations
	LoadLatestVersionAndUpgrade(upgrades *StoreUpgrades) error

	// LoadVersionAndUpgrade will load the named version, but also
	// rename/delete/create sub-store keys, before registering all the keys
	// in order to handle breaking formats in migrations
	LoadVersionAndUpgrade(ver int64, upgrades *StoreUpgrades) error

	// Load a specific persisted version. When you load an old version, or when
	// the last commit attempt didn't complete, the next commit after loading
	// must be idempotent (return the same commit id). Otherwise the behavior is
	// undefined.
	LoadVersion(ver int64) error

	// Set an inter-block (persistent) cache that maintains a mapping from
	// StoreKeys to CommitKVStores.
	SetInterBlockCache(MultiStorePersistentCache)

	// SetInitialVersion sets the initial version of the IAVL tree. It is used when
	// starting a new chain at an arbitrary height.
	SetInitialVersion(version int64) error
}

//---------subsp-------------------------------
// KVStore

// KVStore is a simple interface to get/set data
type KVStore interface {
	Store

	// Get returns nil iff key doesn't exist. Panics on nil key.
	Get(key []byte) []byte

	// Has checks if a key exists. Panics on nil key.
	Has(key []byte) bool

	// Set sets the key. Panics on nil key or value.
	Set(key, value []byte)

	// Delete deletes the key. Panics on nil key.
	Delete(key []byte)

	// Iterator over a domain of keys in ascending order. End is exclusive.
	// Start must be less than end, or the Iterator is invalid.
	// Iterator must be closed by caller.
	// To iterate over entire domain, use store.Iterator(nil, nil)
	// CONTRACT: No writes may happen within a domain while an iterator exists over it.
	// Exceptionally allowed for cachekv.Store, safe to write in the modules.
	Iterator(start, end []byte) Iterator

	// Iterator over a domain of keys in descending order. End is exclusive.
	// Start must be less than end, or the Iterator is invalid.
	// Iterator must be closed by caller.
	// CONTRACT: No writes may happen within a domain while an iterator exists over it.
	// Exceptionally allowed for cachekv.Store, safe to write in the modules.
	ReverseIterator(start, end []byte) Iterator
}

// Iterator is an alias db's Iterator for convenience.
type Iterator = dbm.Iterator

// CacheKVStore branches a KVStore and provides read cache functionality.
// After calling .Write() on the CacheKVStore, all previously created
// CacheKVStores on the object expire.
type CacheKVStore interface {
	KVStore

	// Writes operations to underlying KVStore
	Write()
}

// CommitKVStore is an interface for MultiStore.
type CommitKVStore interface {
	Committer
	KVStore
}

//----------------------------------------
// CacheWrap

// CacheWrap is the most appropriate interface for store ephemeral branching and cache.
// For example, IAVLStore.CacheWrap() returns a CacheKVStore. CacheWrap should not return
// a Committer, since Commit ephemeral store make no sense. It can return KVStore,
// HeapStore, SpaceStore, etc.
type CacheWrap interface {
	// Write syncs with the underlying store.
	Write()

	// CacheWrap recursively wraps again.
	CacheWrap() CacheWrap

	// CacheWrapWithTrace recursively wraps again with tracing enabled.
	CacheWrapWithTrace(w io.Writer, tc TraceContext) CacheWrap
}

type CacheWrapper interface {
	// CacheWrap branches a store.
	CacheWrap() CacheWrap

	// CacheWrapWithTrace branches a store with tracing enabled.
	CacheWrapWithTrace(w io.Writer, tc TraceContext) CacheWrap
}

func (cid CommitID) IsZero() bool {
	return cid.Version == 0 && len(cid.Hash) == 0
}

func (cid CommitID) String() string {
	return fmt.Sprintf("CommitID{%v:%X}", cid.Hash, cid.Version)
}

//----------------------------------------
// Store types

// kind of store
type StoreType int

const (
	StoreTypeMulti StoreType = iota
	StoreTypeDB
	StoreTypeIAVL
	StoreTypeTransient
	StoreTypeMemory
)

func (st StoreType) String() string {
	switch st {
	case StoreTypeMulti:
		return "StoreTypeMulti"

	case StoreTypeDB:
		return "StoreTypeDB"

	case StoreTypeIAVL:
		return "StoreTypeIAVL"

	case StoreTypeTransient:
		return "StoreTypeTransient"

	case StoreTypeMemory:
		return "StoreTypeMemory"
	}

	return "unknown store type"
}

//----------------------------------------
// Keys for accessing substores

// StoreKey is a key used to index stores in a MultiStore.
type StoreKey interface {
	Name() string
	String() string
}

// CapabilityKey represent the Cosmos SDK keys for object-capability
// generation in the IBC protocol as defined in https://github.com/cosmos/ics/tree/master/spec/ics-005-port-allocation#data-structures
type CapabilityKey StoreKey

// KVStoreKey is used for accessing substores.
// Only the pointer value should ever be used - it functions as a capabilities key.
type KVStoreKey struct {
	name string
}

// NewKVStoreKey returns a new pointer to a KVStoreKey.
// Use a pointer so keys don't collide.
func NewKVStoreKey(name string) *KVStoreKey {
	if name == "" {
		panic("empty key name not allowed")
	}
	return &KVStoreKey{
		name: name,
	}
}

func (key *KVStoreKey) Name() string {
	return key.name
}

func (key *KVStoreKey) String() string {
	return fmt.Sprintf("KVStoreKey{%p, %s}", key, key.name)
}

// TransientStoreKey is used for indexing transient stores in a MultiStore
type TransientStoreKey struct {
	name string
}

// Constructs new TransientStoreKey
// Must return a pointer according to the ocap principle
func NewTransientStoreKey(name string) *TransientStoreKey {
	return &TransientStoreKey{
		name: name,
	}
}

// Implements StoreKey
func (key *TransientStoreKey) Name() string {
	return key.name
}

// Implements StoreKey
func (key *TransientStoreKey) String() string {
	return fmt.Sprintf("TransientStoreKey{%p, %s}", key, key.name)
}

// MemoryStoreKey defines a typed key to be used with an in-memory KVStore.
type MemoryStoreKey struct {
	name string
}

func NewMemoryStoreKey(name string) *MemoryStoreKey {
	return &MemoryStoreKey{name: name}
}

// Name returns the name of the MemoryStoreKey.
func (key *MemoryStoreKey) Name() string {
	return key.name
}

// String returns a stringified representation of the MemoryStoreKey.
func (key *MemoryStoreKey) String() string {
	return fmt.Sprintf("MemoryStoreKey{%p, %s}", key, key.name)
}

//----------------------------------------

// key-value result for iterator queries
type KVPair kv.Pair

//----------------------------------------

// TraceContext contains TraceKVStore context data. It will be written with
// every trace operation.
type TraceContext map[string]interface{}

// MultiStorePersistentCache defines an interface which provides inter-block
// (persistent) caching capabilities for multiple CommitKVStores based on StoreKeys.
type MultiStorePersistentCache interface {
	// Wrap and return the provided CommitKVStore with an inter-block (persistent)
	// cache.
	GetStoreCache(key StoreKey, store CommitKVStore) CommitKVStore

	// Return the underlying CommitKVStore for a StoreKey.
	Unwrap(key StoreKey) CommitKVStore

	// Reset the entire set of internal caches.
	Reset()
}

// StoreWithInitialVersion is a store that can have an arbitrary initial
// version.
type StoreWithInitialVersion interface {
	// SetInitialVersion sets the initial version of the IAVL tree. It is used when
	// starting a new chain at an arbitrary height.
	SetInitialVersion(version int64)
}
