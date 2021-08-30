package types

import (
	"io"

	snapshottypes "github.com/cosmos/cosmos-sdk/snapshots/types"
	v1 "github.com/cosmos/cosmos-sdk/store/types"
)

type StoreKey = v1.StoreKey
type CommitID = v1.CommitID
type StoreUpgrades = v1.StoreUpgrades
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

var (
	PruneDefault    = v1.PruneDefault
	PruneEverything = v1.PruneEverything
	PruneNothing    = v1.PruneNothing
)

type BasicRootStore interface {
	GetKVStore(StoreKey) KVStore
	CacheRootStore() CacheRootStore
}

type rootStoreTraceListen interface {
	TracingEnabled() bool
	SetTracer(w io.Writer)
	SetTraceContext(TraceContext)
	ListeningEnabled(key StoreKey) bool
	AddListeners(key StoreKey, listeners []WriteListener)
}

type RootStore interface {
	BasicRootStore
	rootStoreTraceListen

	GetVersion(uint64) (BasicRootStore, error)
}

type CommitRootStore interface {
	RootStore
	Committer
	snapshottypes.Snapshotter // or PortableStore
	SetInitialVersion(uint64) error
}

type CacheRootStore interface {
	BasicRootStore
	rootStoreTraceListen
	Write()
}

// provides inter-block (persistent) caching capabilities for a CommitRootStore
// TODO
type RootStorePersistentCache = v1.MultiStorePersistentCache

//----------------------------------------
// Store types

type StoreType = v1.StoreType

// Valid types
const StoreTypeMemory = v1.StoreTypeMemory
const StoreTypeTransient = v1.StoreTypeTransient
const StoreTypeDecoupled = v1.StoreTypeDecoupled
const StoreTypeSMT = v1.StoreTypeSMT
const StoreTypePersistent = StoreTypeDecoupled
