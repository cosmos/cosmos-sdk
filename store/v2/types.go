package types

import (
	"io"

	snapshottypes "github.com/cosmos/cosmos-sdk/snapshots/types"
	v1 "github.com/cosmos/cosmos-sdk/store/types"
)

// Re-export original store types

type StoreKey = v1.StoreKey
type StoreType = v1.StoreType
type CommitID = v1.CommitID
type StoreUpgrades = v1.StoreUpgrades
type StoreRename = v1.StoreRename
type Iterator = v1.Iterator
type PruningOptions = v1.PruningOptions

type TraceContext = v1.TraceContext
type WriteListener = v1.WriteListener

type BasicKVStore = v1.BasicKVStore
type KVStore = v1.KVStore
type Committer = v1.Committer
type CommitKVStore = v1.CommitKVStore
type CacheKVStore = v1.CacheKVStore
type Queryable = v1.Queryable
type CacheWrap = v1.CacheWrap

type KVStoreKey = v1.KVStoreKey
type MemoryStoreKey = v1.MemoryStoreKey
type TransientStoreKey = v1.TransientStoreKey

var (
	PruneDefault    = v1.PruneDefault
	PruneEverything = v1.PruneEverything
	PruneNothing    = v1.PruneNothing
)

// BasicRootStore defines a minimal interface for accessing root state.
type BasicRootStore interface {
	// Returns a KVStore which has access only to the namespace of the StoreKey.
	// Panics if the key is not found in the schema.
	GetKVStore(StoreKey) KVStore
	// Returns a branched whose modifications are later merged back in.
	CacheRootStore() CacheRootStore
}

// mixin interface for trace and listen methods
type rootStoreTraceListen interface {
	TracingEnabled() bool
	SetTracer(w io.Writer)
	SetTraceContext(TraceContext)
	ListeningEnabled(key StoreKey) bool
	AddListeners(key StoreKey, listeners []WriteListener)
}

// CommitRootStore defines a complete interface for persistent root state, including
// (read-only) access to past versions, pruning, trace/listen, and state snapshots.
type CommitRootStore interface {
	BasicRootStore
	rootStoreTraceListen

	// Gets a read-only view of the store at a specific version.
	// Returns an error if the version is not found.
	GetVersion(int64) (BasicRootStore, error)
	// Closes the store and all backing transactions.
	Close() error

	// RootStore
	Committer
	snapshottypes.Snapshotter // todo: PortableStore?
	SetInitialVersion(uint64) error
}

// CacheRootStore defines a branch of the root state which can be written back to the source store.
type CacheRootStore interface {
	BasicRootStore
	rootStoreTraceListen
	Write()
}

// RootStorePersistentCache provides inter-block (persistent) caching capabilities for a CommitRootStore.
type RootStorePersistentCache = v1.MultiStorePersistentCache

// Re-export relevant store type values and utility functions

const StoreTypeMemory = v1.StoreTypeMemory
const StoreTypeTransient = v1.StoreTypeTransient
const StoreTypeDB = v1.StoreTypeDB
const StoreTypeSMT = v1.StoreTypeSMT
const StoreTypePersistent = v1.StoreTypePersistent

var NewKVStoreKey = v1.NewKVStoreKey
var PrefixEndBytes = v1.PrefixEndBytes
var KVStorePrefixIterator = v1.KVStorePrefixIterator
var KVStoreReversePrefixIterator = v1.KVStoreReversePrefixIterator
