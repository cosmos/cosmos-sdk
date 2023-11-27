package commitment

import (
	"io"

	ics23 "github.com/cosmos/ics23/go"
)

// Tree is the interface that wraps the basic Tree methods.
type Tree interface {
	Set(key, value []byte) error
	Remove(key []byte) error
	GetLatestVersion() uint64
	WorkingHash() []byte
	LoadVersion(version uint64) error
	Commit() ([]byte, error)
	GetProof(version uint64, key []byte) (*ics23.CommitmentProof, error)
	Prune(version uint64) error

	io.Closer
}
