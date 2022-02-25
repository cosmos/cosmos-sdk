package types

import (
	"io"

	snapshottypes "github.com/cosmos/cosmos-sdk/snapshots/types"
	v1 "github.com/cosmos/cosmos-sdk/store/types"
)

// Re-export relevant original store types
type (
	StoreKey       = v1.StoreKey
	StoreType      = v1.StoreType
	CommitID       = v1.CommitID
	StoreUpgrades  = v1.StoreUpgrades
	StoreRename    = v1.StoreRename
	Iterator       = v1.Iterator
	PruningOptions = v1.PruningOptions

	TraceContext  = v1.TraceContext
	WriteListener = v1.WriteListener

	BasicKVStore  = v1.BasicKVStore
	KVStore       = v1.KVStore
	Committer     = v1.Committer
	CommitKVStore = v1.CommitKVStore
	CacheKVStore  = v1.CacheKVStore
	Queryable     = v1.Queryable
	CacheWrap     = v1.CacheWrap

	KVStoreKey        = v1.KVStoreKey
	MemoryStoreKey    = v1.MemoryStoreKey
	TransientStoreKey = v1.TransientStoreKey

	KVPair      = v1.KVPair
	StoreKVPair = v1.StoreKVPair
)

// Re-export relevant constants, values and utility functions
const (
	StoreTypeMemory     = v1.StoreTypeMemory
	StoreTypeTransient  = v1.StoreTypeTransient
	StoreTypeDB         = v1.StoreTypeDB
	StoreTypeSMT        = v1.StoreTypeSMT
	StoreTypePersistent = v1.StoreTypePersistent
)

var (
	PruneDefault    = v1.PruneDefault
	PruneEverything = v1.PruneEverything
	PruneNothing    = v1.PruneNothing

	NewKVStoreKey                = v1.NewKVStoreKey
	PrefixEndBytes               = v1.PrefixEndBytes
	KVStorePrefixIterator        = v1.KVStorePrefixIterator
	KVStoreReversePrefixIterator = v1.KVStoreReversePrefixIterator

	NewStoreKVPairWriteListener = v1.NewStoreKVPairWriteListener

	AssertValidKey   = v1.AssertValidKey
	AssertValidValue = v1.AssertValidValue

	CommitmentOpDecoder = v1.CommitmentOpDecoder
	ProofOpFromMap      = v1.ProofOpFromMap

	ProofOpSMTCommitment          = v1.ProofOpSMTCommitment
	ProofOpSimpleMerkleCommitment = v1.ProofOpSimpleMerkleCommitment
	NewSmtCommitmentOp            = v1.NewSmtCommitmentOp
)

// MultiStore defines a minimal interface for accessing root state.
type MultiStore interface {
	// Returns a KVStore which has access only to the namespace of the StoreKey.
	// Panics if the key is not found in the schema.
	GetKVStore(StoreKey) KVStore
	// Returns a branched store whose modifications are later merged back in.
	CacheWrap() CacheMultiStore
}

// mixin interface for trace and listen methods
type rootStoreTraceListen interface {
	TracingEnabled() bool
	SetTracer(w io.Writer)
	SetTracingContext(TraceContext)
	ListeningEnabled(key StoreKey) bool
	AddListeners(key StoreKey, listeners []WriteListener)
}

// CommitMultiStore defines a complete interface for persistent root state, including
// (read-only) access to past versions, pruning, trace/listen, and state snapshots.
type CommitMultiStore interface {
	MultiStore
	rootStoreTraceListen
	Committer
	snapshottypes.Snapshotter

	// Gets a read-only view of the store at a specific version.
	// Returns an error if the version is not found.
	GetVersion(int64) (MultiStore, error)
	// Closes the store and all backing transactions.
	Close() error
	// Defines the minimum version number that can be saved by this store.
	SetInitialVersion(uint64) error
	// Gets all versions in the DB
	// https://github.com/cosmos/cosmos-sdk/pull/11124
	GetAllVersions() []int
}

// CacheMultiStore defines a branch of the root state which can be written back to the source store.
type CacheMultiStore interface {
	MultiStore
	rootStoreTraceListen

	// Write all cached changes back to the source store. Note: this overwrites any intervening changes.
	Write()
}

// MultiStorePersistentCache provides inter-block (persistent) caching capabilities for a CommitMultiStore.
// TODO: placeholder. Implement and redefine this
type MultiStorePersistentCache = v1.MultiStorePersistentCache
