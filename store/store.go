package store

import (
	"io"

	ics23 "github.com/cosmos/ics23/go"
)

// TODO: Move relevant types to the 'core' package.

// StoreType defines a type of KVStore.
type StoreType int

// Sentinel store types.
const (
	StoreTypeBranch StoreType = iota
	StoreTypeTrace
	StoreTypeMem
)

// RootStore defines an abstraction layer containing a State Storage (SS) engine
// and one or more State Commitment (SC) engines.
type RootStore interface {
	GetSCStore(storeKey string) Tree
	MountSCStore(storeKey string, sc Tree) error
	GetKVStore(storeKey string) KVStore
	GetBranchedKVStore(storeKey string) BranchedKVStore

	GetProof(storeKey string, version uint64, key []byte) (*ics23.CommitmentProof, error)

	Branch() BranchedRootStore

	SetTracingContext(tc TraceContext)
	SetTracer(w io.Writer)
	TracingEnabled() bool

	LoadVersion(version uint64) error
	LoadLatestVersion() error
	GetLatestVersion() (uint64, error)

	WorkingHash() ([]byte, error)
	SetCommitHeader(h CommitHeader)
	Commit() ([]byte, error)

	// TODO:
	//
	// - Queries
	//
	// Ref: https://github.com/cosmos/cosmos-sdk/issues/17314

	io.Closer
}

// BranchedRootStore defines an extension of the RootStore interface that allows
// for nested branching and flushing of writes. It extends RootStore by allowing
// a caller to call Branch() which should return a BranchedRootStore that has all
// internal relevant KV stores branched. A caller can then call Write() on the
// BranchedRootStore which will flush all changesets to the parent RootStore's
// internal KV stores.
type BranchedRootStore interface {
	RootStore

	Write()
}

// KVStore defines the core storage primitive for modules to read and write state.
type KVStore interface {
	GetStoreKey() string

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

	// GetChangeset returns the ChangeSet, if any, for the branched state. This
	// should contain all writes that are marked to be flushed and committed during
	// Commit().
	GetChangeset() *Changeset

	// Reset resets the store, which is implementation dependent.
	Reset() error

	// Iterator creates a new Iterator over the domain [start, end). Note:
	//
	// - Start must be less than end
	// - The iterator must be closed by caller
	// - To iterate over entire domain, use store.Iterator(nil, nil)
	//
	// CONTRACT: No writes may happen within a domain while an iterator exists over
	// it, with the exception of a branched/cached KVStore.
	Iterator(start, end []byte) Iterator

	// ReverseIterator creates a new reverse Iterator over the domain [start, end).
	// It has the some properties and contracts as Iterator.
	ReverseIterator(start, end []byte) Iterator
}

// BranchedKVStore defines an interface for a branched a KVStore. It extends KVStore
// by allowing dirty entries to be flushed to the underlying KVStore or discarded
// altogether. A BranchedKVStore can itself be branched, allowing for nested branching
// where writes are flushed up the branched stack.
type BranchedKVStore interface {
	KVStore

	// Write flushes writes to the underlying store.
	Write()

	// Branch recursively wraps.
	Branch() BranchedKVStore

	// BranchWithTrace recursively wraps with tracing enabled.
	BranchWithTrace(w io.Writer, tc TraceContext) BranchedKVStore
}
