package commitment

import (
	"errors"
	"io"

	ics23 "github.com/cosmos/ics23/go"

	snapshotstypes "cosmossdk.io/store/v2/snapshots/types"
)

// ErrorExportDone is returned by Exporter.Next() when all items have been exported.
var ErrorExportDone = errors.New("export is complete")

// Tree is the interface that wraps the basic Tree methods.
type Tree interface {
	Set(key, value []byte) error
	Remove(key []byte) error
	GetLatestVersion() uint64
	WorkingHash() []byte
	LoadVersion(version uint64) error
	Commit() ([]byte, error)
	SetInitialVersion(version uint64) error
	GetProof(version uint64, key []byte) (*ics23.CommitmentProof, error)
	Prune(version uint64) error
	Export(version uint64) (Exporter, error)
	Import(version uint64) (Importer, error)

	io.Closer
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
