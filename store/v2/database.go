package store

import (
	"io"

	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/store/v2/proof"
)

// VersionedDatabase defines an API for a versioned database that allows reads,
// writes, iteration and commitment over a series of versions.
type VersionedDatabase interface {
	Has(storeKey []byte, version uint64, key []byte) (bool, error)
	Get(storeKey []byte, version uint64, key []byte) ([]byte, error)
	GetLatestVersion() (uint64, error)
	SetLatestVersion(version uint64) error

	Iterator(storeKey []byte, version uint64, start, end []byte) (corestore.Iterator, error)
	ReverseIterator(storeKey []byte, version uint64, start, end []byte) (corestore.Iterator, error)

	ApplyChangeset(version uint64, cs *corestore.Changeset) error

	// Close releases associated resources. It should NOT be idempotent. It must
	// only be called once and any call after may panic.
	io.Closer
}

// UpgradableDatabase defines an API for a versioned database that allows pruning
// deleted storeKeys
type UpgradableDatabase interface {
	// PruneStoreKeys prunes all data associated with the given storeKeys whenever
	// the given version is pruned.
	PruneStoreKeys(storeKeys []string, version uint64) error
}

// Committer defines an API for committing state.
type Committer interface {
	// WriteChangeset writes the changeset to the commitment state.
	WriteChangeset(cs *corestore.Changeset) error

	// WorkingCommitInfo returns the CommitInfo for the working tree.
	WorkingCommitInfo(version uint64) *proof.CommitInfo

	// GetLatestVersion returns the latest version.
	GetLatestVersion() (uint64, error)

	// LoadVersion loads the tree at the given version.
	LoadVersion(targetVersion uint64) error

	// Commit commits the working tree to the database.
	Commit(version uint64) (*proof.CommitInfo, error)

	// GetProof returns the proof of existence or non-existence for the given key.
	GetProof(storeKey []byte, version uint64, key []byte) ([]proof.CommitmentOp, error)

	// Get returns the value for the given key at the given version.
	//
	// NOTE: This method only exists to support migration from IAVL v0/v1 to v2.
	// Once migration is complete, this method should be removed and/or not used.
	Get(storeKey []byte, version uint64, key []byte) ([]byte, error)

	// SetInitialVersion sets the initial version of the committer.
	SetInitialVersion(version uint64) error

	// GetCommitInfo returns the CommitInfo for the given version.
	GetCommitInfo(version uint64) (*proof.CommitInfo, error)

	// Close releases associated resources. It should NOT be idempotent. It must
	// only be called once and any call after may panic.
	io.Closer
}
