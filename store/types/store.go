package types

import (
	"io"

	"cosmossdk.io/store/v2"
	"cosmossdk.io/store/v2/commitment"
	ics23 "github.com/cosmos/ics23/go"
)

// StoreType defines a type of KVStore.
type StoreType int

// MultiStore defines an abstraction layer containing a State Storage (SS) engine
// and one or more State Commitment (SC) engines.
//
// TODO:
// - Move relevant types to the 'core' package.
type MultiStore interface {
	GetSCStore(storeKey StoreKey) *commitment.Database
	MountSCStore(storeKey StoreKey, sc *commitment.Database) error
	GetProof(storeKey StoreKey, version uint64, key []byte) (*ics23.CommitmentProof, error)
	LoadVersion(version uint64) error
	LoadLatestVersion() error
	GetLatestVersion() (uint64, error)
	WorkingHash() []byte
	Commit() ([]byte, error)
	SetCommitHeader(h CommitHeader)

	// TODO:
	//
	// - Tracing
	// - Branching
	// - Queries
	//
	// Ref: https://github.com/cosmos/cosmos-sdk/issues/17314

	io.Closer
}

// KVStore defines the core storage primitive for modules to read and write state.
type KVStore interface {
	// GetStoreType returns the concrete store type.
	GetStoreType() StoreType

	// Get returns a value for a given key from the store.
	Get(key []byte) []byte

	// Has checks if a key exists.
	Has(key []byte) bool

	// Set sets a key/value entry to the store.
	Set(key, value []byte)

	// Delete deletes the key from the store.
	Delete(key []byte)

	CacheWrapper

	// Iterator creates a new Iterator over the domain [start, end). Note:
	//
	// - Start must be less than end
	// - The iterator must be closed by caller
	// - To iterate over entire domain, use store.Iterator(nil, nil)
	//
	// CONTRACT: No writes may happen within a domain while an iterator exists over
	// it, with the exception of a branched/cached KVStore.
	Iterator(start, end []byte) store.Iterator

	// ReverseIterator creates a new reverse Iterator over the domain [start, end).
	// It has the some properties and contracts as Iterator.
	ReverseIterator(start, end []byte) store.Iterator
}

// CacheKVStore defines an interface for a branched a KVStore. It extends KVStore
// by allowing dirty entries to be flushed to the underlying KVStore or discarded
// altogether. A CachedKVStore can itself be branched, allowing for nested branching
// where writes are flushed up the branched stack.
type CacheKVStore interface {
	KVStore

	// Write flushes writes to the underlying store.
	Write()

	// CacheWrap recursively wraps.
	CacheWrap() CacheKVStore

	// CacheWrapWithTrace recursively wraps with tracing enabled.
	CacheWrapWithTrace(w io.Writer, tc TraceContext) CacheKVStore
}

// CacheWrapper defines an interface for a branching a KVStore's state, allowing
// writes to be cached and flushed to the underlying store or discarded altogether.
// Reads should be performed against a "branched" state, allowing dirty entries
// to be cached and read from. If an entry is not found in the branched state, it
// will fallthrough to the underlying KVStore.
type CacheWrapper interface {
	CacheWrap() CacheKVStore

	// CacheWrapWithTrace branches a store with tracing enabled.
	CacheWrapWithTrace(w io.Writer, tc TraceContext) CacheKVStore
}
