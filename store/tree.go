package store

import (
	ics23 "github.com/cosmos/ics23/go"
)

// Tree is an interface for a commitment layer to support multiple backends.
type Tree interface {
	WriteBatch(cs *ChangeSet) error
	WorkingHash() []byte
	GetLatestVersion() uint64
	LoadVersion(targetVersion uint64) error
	Commit() ([]byte, error)
	GetProof(version uint64, key []byte) (*ics23.CommitmentProof, error)
	Close() error
}
