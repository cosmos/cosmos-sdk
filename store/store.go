package store

import (
	"io"

	ics23 "github.com/cosmos/ics23/go"
)

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
	// GetSCStore should return the SC backend for the given store key. A RootStore
	// implementation may choose to ignore the store key in cases where only a single
	// SC backend is used.
	GetSCStore(storeKey string) Tree
	// MountSCStore should mount the given SC backend for the given store key. For
	// implementations that utilize a single SC backend, this method may be optional
	// or a no-op.
	MountSCStore(storeKey string, sc Tree) error
	// GetKVStore returns the KVStore for the given store key. If an implementation
	// chooses to have a single SS backend, the store key may be ignored.
	GetKVStore(storeKey string) KVStore
	// GetBranchedKVStore returns the KVStore for the given store key. If an
	// implementation chooses to have a single SS backend, the store key may be
	// ignored.
	GetBranchedKVStore(storeKey string) BranchedKVStore

	// GetProof returns a proof for the given key, version (height), and store key
	// tuple. See the CommitmentProof type for the concrete supported proof types.
	GetProof(storeKey string, version uint64, key []byte) (*ics23.CommitmentProof, error)

	// Query(*RequestQuery) (*ResponseQuery, error)

	// Branch should branch the entire RootStore, i.e. a copy of the original RootStore
	// except with all internal KV store(s) branched.
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
