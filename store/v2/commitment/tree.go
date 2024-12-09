package commitment

import (
	"errors"
	"io"

	ics23 "github.com/cosmos/ics23/go"

	corestore "cosmossdk.io/core/store"
	snapshotstypes "cosmossdk.io/store/v2/snapshots/types"
)

// ErrorExportDone is returned by Exporter.Next() when all items have been exported.
var ErrorExportDone = errors.New("export is complete")

// Tree is the interface that wraps the basic Tree methods.
type Tree interface {
	Set(key, value []byte) error
	Remove(key []byte) error
	GetLatestVersion() (uint64, error)

	// Hash returns the hash of the current version of the tree
	Hash() []byte
	// Version returns the current version of the tree
	Version() uint64

	LoadVersion(version uint64) error
	LoadVersionForOverwriting(version uint64) error
	Commit() ([]byte, uint64, error)
	SetInitialVersion(version uint64) error
	GetProof(version uint64, key []byte) (*ics23.CommitmentProof, error)

	Prune(version uint64) error
	Export(version uint64) (Exporter, error)
	Import(version uint64) (Importer, error)

	io.Closer
}

// Reader is the optional interface that is only used to read data from the tree
// during the migration process.
type Reader interface {
	Get(version uint64, key []byte) ([]byte, error)
	Iterator(version uint64, start, end []byte, ascending bool) (corestore.Iterator, error)
}

// Exporter is the interface that wraps the basic Export methods.
type Exporter interface {
	Next() (*snapshotstypes.SnapshotIAVLItem, error)

	io.Closer
}

// Importer is the interface that wraps the basic Import methods.
type Importer interface {
	Add(*snapshotstypes.SnapshotIAVLItem) error
	Commit() error

	io.Closer
}
