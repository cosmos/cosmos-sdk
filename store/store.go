package store

import (
	"io"

	coreheader "cosmossdk.io/core/header"
	"cosmossdk.io/store/v2/metrics"
)

// RootStore defines an abstraction layer containing a State Storage (SS) engine
// and one or more State Commitment (SC) engines.
type RootStore interface {
	// GetSCStore should return the SC backend.
	GetSCStore() Committer

	// Query performs a query on the RootStore for a given store key, version (height),
	// and key tuple. Queries should be routed to the underlying SS engine.
	Query(storeKey string, version uint64, key []byte, prove bool) (QueryResult, error)

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

// QueryResult defines the response type to performing a query on a RootStore.
type QueryResult struct {
	Key     []byte
	Value   []byte
	Version uint64
	Proof   CommitmentOp
}
