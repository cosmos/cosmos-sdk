package store

import (
	"io"

	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/store/v2/proof"
)

type VersionedReader interface {
	Has(storeKey []byte, version uint64, key []byte) (bool, error)
	Get(storeKey []byte, version uint64, key []byte) ([]byte, error)

	GetLatestVersion() (uint64, error)
	VersionExists(v uint64) (bool, error)

	Iterator(storeKey []byte, version uint64, start, end []byte) (corestore.Iterator, error)
	ReverseIterator(storeKey []byte, version uint64, start, end []byte) (corestore.Iterator, error)
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
	UpgradeableStore
	VersionedReader
	// WriteChangeset writes the changeset to the commitment state.
	WriteChangeset(cs *corestore.Changeset) error

	// GetLatestVersion returns the latest version.
	GetLatestVersion() (uint64, error)

	// LoadVersion loads the tree at the given version.
	LoadVersion(targetVersion uint64) error

	// LoadVersionForOverwriting loads the tree at the given version.
	// Any versions greater than targetVersion will be deleted.
	LoadVersionForOverwriting(targetVersion uint64) error

	// Commit commits the working tree to the database.
	Commit(version uint64) (*proof.CommitInfo, error)

	// GetProof returns the proof of existence or non-existence for the given key.
	GetProof(storeKey []byte, version uint64, key []byte) ([]proof.CommitmentOp, error)

	// SetInitialVersion sets the initial version of the committer.
	SetInitialVersion(version uint64) error

	// GetCommitInfo returns the CommitInfo for the given version.
	GetCommitInfo(version uint64) (*proof.CommitInfo, error)

	Get(storeKey []byte, version uint64, key []byte) ([]byte, error)

	// Closer releases associated resources. It should NOT be idempotent. It must
	// only be called once and any call after may panic.
	io.Closer
}
