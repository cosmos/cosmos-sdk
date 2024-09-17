package store

import (
	"io"

	coreheader "cosmossdk.io/core/header"
	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/store/v2/metrics"
	"cosmossdk.io/store/v2/proof"
)

// RootStore defines an abstraction layer containing a State Storage (SS) engine
// and one or more State Commitment (SC) engines.
type RootStore interface {
	// StateLatest returns a read-only version of the RootStore at the latest
	// height, alongside the associated version.
	StateLatest() (uint64, corestore.ReaderMap, error)

	// StateAt is analogous to StateLatest() except it returns a read-only version
	// of the RootStore at the provided version. If such a version cannot be found,
	// an error must be returned.
	StateAt(version uint64) (corestore.ReaderMap, error)

	// GetStateStorage returns the SS backend.
	GetStateStorage() VersionedDatabase

	// GetStateCommitment returns the SC backend.
	GetStateCommitment() Committer

	// Query performs a query on the RootStore for a given store key, version (height),
	// and key tuple. Queries should be routed to the underlying SS engine.
	Query(storeKey []byte, version uint64, key []byte, prove bool) (QueryResult, error)

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

	// WorkingHash returns the current WIP commitment hash by applying the Changeset
	// to the SC backend. It is only used to get the hash of the intermediate state
	// before committing, the typical use case is for the genesis block.
	// NOTE: It also writes the changeset to the SS backend.
	WorkingHash(cs *corestore.Changeset) ([]byte, error)

	// Commit should be responsible for taking the provided changeset and flushing
	// it to disk. Note, it will overwrite the changeset if WorkingHash() was called.
	// Commit() should ensure the changeset is committed to all SC and SS backends
	// and flushed to disk. It must return a hash of the merkle-ized committed state.
	Commit(cs *corestore.Changeset) ([]byte, error)

	// LastCommitID returns a CommitID pertaining to the last commitment.
	LastCommitID() (proof.CommitID, error)

	// SetMetrics sets the telemetry handler on the RootStore.
	SetMetrics(m metrics.Metrics)

	Prune(version uint64) error

	io.Closer
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
