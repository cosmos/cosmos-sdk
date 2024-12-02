package store

import (
	"io"

	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/store/v2/metrics"
	"cosmossdk.io/store/v2/proof"
)

// RootStore defines an abstraction layer containing a State Storage (SS) engine
// and one or more State Commitment (SC) engines.
type RootStore interface {
	Pruner
	Backend

	// StateLatest returns a read-only version of the RootStore at the latest
	// height, alongside the associated version.
	StateLatest() (uint64, corestore.ReaderMap, error)

	// StateAt is analogous to StateLatest() except it returns a read-only version
	// of the RootStore at the provided version. If such a version cannot be found,
	// an error must be returned.
	StateAt(version uint64) (corestore.ReaderMap, error)

	// Query performs a query on the RootStore for a given store key, version (height),
	// and key tuple. Queries should be routed to the underlying SS engine.
	Query(storeKey []byte, version uint64, key []byte, prove bool) (QueryResult, error)

	// LoadVersion loads the RootStore to the given version.
	LoadVersion(version uint64) error

	// LoadVersionForOverwriting loads the state at the given version.
	// Any versions greater than targetVersion will be deleted.
	LoadVersionForOverwriting(version uint64) error

	// LoadLatestVersion behaves identically to LoadVersion except it loads the
	// latest version implicitly.
	LoadLatestVersion() error

	// GetLatestVersion returns the latest version, i.e. height, committed.
	GetLatestVersion() (uint64, error)

	// SetInitialVersion sets the initial version on the RootStore.
	SetInitialVersion(v uint64) error

	// Commit should be responsible for taking the provided changeset and flushing
	// it to disk. Note, it will overwrite the changeset if WorkingHash() was called.
	// Commit() should ensure the changeset is committed to all SC and SS backends
	// and flushed to disk. It must return a hash of the merkle-ized committed state.
	Commit(cs *corestore.Changeset) ([]byte, error)

	// LastCommitID returns a CommitID pertaining to the last commitment.
	LastCommitID() (proof.CommitID, error)

	// SetMetrics sets the telemetry handler on the RootStore.
	SetMetrics(m metrics.Metrics)

	io.Closer
}

// Backend defines the interface for the RootStore backends.
type Backend interface {
	// GetStateCommitment returns the SC backend.
	GetStateCommitment() Committer
}

// UpgradeableStore defines the interface for upgrading store keys.
type UpgradeableStore interface {
	// LoadVersionAndUpgrade behaves identically to LoadVersion except it also
	// accepts a StoreUpgrades object that defines a series of transformations to
	// apply to store keys (if any).
	//
	// Note, handling StoreUpgrades is optional depending on the underlying store
	// implementation.
	LoadVersionAndUpgrade(version uint64, upgrades *corestore.StoreUpgrades) error
}

// Pruner defines the interface for pruning old versions of the store or database.
type Pruner interface {
	// Prune prunes the store to the provided version.
	Prune(version uint64) error
}

// PausablePruner extends the Pruner interface to include the API for pausing
// the pruning process.
type PausablePruner interface {
	Pruner

	// PausePruning pauses or resumes the pruning process to avoid the parallel writes
	// while committing the state.
	PausePruning(pause bool)
}

// QueryResult defines the response type to performing a query on a RootStore.
type QueryResult struct {
	Key      []byte
	Value    []byte
	Version  uint64
	ProofOps []proof.CommitmentOp
}
