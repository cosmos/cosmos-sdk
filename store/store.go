package store

import (
	"io"

	coreheader "cosmossdk.io/core/header"
	"cosmossdk.io/store/v2/metrics"
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
	// GetSCStore should return the SC backend.
	GetSCStore() Committer
	// GetKVStore returns the KVStore for the given store key. If an implementation
	// chooses to have a single SS backend, the store key may be ignored.
	GetKVStore(storeKey string) KVStore
	// GetBranchedKVStore returns the KVStore for the given store key. If an
	// implementation chooses to have a single SS backend, the store key may be
	// ignored.
	GetBranchedKVStore(storeKey string) BranchedKVStore

	// Query performs a query on the RootStore for a given store key, version (height),
	// and key tuple. Queries should be routed to the underlying SS engine.
	Query(storeKey string, version uint64, key []byte, prove bool) (QueryResult, error)

	// Branch should branch the entire RootStore, i.e. a copy of the original RootStore
	// except with all internal KV store(s) branched.
	Branch() BranchedRootStore

	// SetTracingContext sets the tracing context, i.e tracing metadata, on the
	// RootStore.
	SetTracingContext(tc TraceContext)
	// SetTracer sets the tracer on the RootStore, such that any calls to GetKVStore
	// or GetBranchedKVStore, will have tracing enabled.
	SetTracer(w io.Writer)
	// TracingEnabled returns true if tracing is enabled on the RootStore.
	TracingEnabled() bool

	// LoadVersion loads the RootStore to the given version.
	LoadVersion(version uint64) error
	// LoadLatestVersion behaves identically to LoadVersion except it loads the
	// latest version implicitly.
	LoadLatestVersion() error

	// GetLatestVersion returns the latest version, i.e. height, committed.
	GetLatestVersion() (uint64, error)

	// SetInitialVersion sets the initial version on the RootStore.
	SetInitialVersion(v uint64) error

	// SetCommitHeader sets the commit header for the next commit. This call and
	// implementation is optional. However, it must be supported in cases where
	// queries based on block time need to be supported.
	SetCommitHeader(h *coreheader.Info)

	// WorkingHash returns the current WIP commitment hash. Depending on the underlying
	// implementation, this may need to take the current changeset and write it to
	// the SC backend(s). In such cases, Commit() would return this hash and flush
	// writes to disk. This means that WorkingHash mutates the RootStore and must
	// be called prior to Commit().
	WorkingHash() ([]byte, error)
	// Commit should be responsible for taking the current changeset and flushing
	// it to disk. Note, depending on the implementation, the changeset, at this
	// point, may already be written to the SC backends. Commit() should ensure
	// the changeset is committed to all SC and SC backends and flushed to disk.
	// It must return a hash of the merkle-ized committed state. This hash should
	// be the same as the hash returned by WorkingHash() prior to calling Commit().
	Commit() ([]byte, error)

	// LastCommitID returns a CommitID pertaining to the last commitment.
	LastCommitID() (CommitID, error)

	// SetMetrics sets the telemetry handler on the RootStore.
	SetMetrics(m metrics.Metrics)

	io.Closer
}

// UpgradeableRootStore extends the RootStore interface to support loading versions
// with upgrades.
type UpgradeableRootStore interface {
	RootStore

	// LoadVersionAndUpgrade behaves identically to LoadVersion except it also
	// accepts a StoreUpgrades object that defines a series of transformations to
	// apply to store keys (if any).
	//
	// Note, handling StoreUpgrades is optional depending on the underlying RootStore
	// implementation.
	LoadVersionAndUpgrade(version uint64, upgrades *StoreUpgrades) error
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
	Reset(toVersion uint64) error

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

// QueryResult defines the response type to performing a query on a RootStore.
type QueryResult struct {
	Key     []byte
	Value   []byte
	Version uint64
	Proof   CommitmentOp
}
