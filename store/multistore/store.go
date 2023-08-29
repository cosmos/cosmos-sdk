package multistore

import (
	"cosmossdk.io/store/v2"
	"cosmossdk.io/store/v2/commitment"
)

// TODO: Move this to Core package.
type MultiStore interface {
	WorkingHash() []byte
	Commit() error
}

type Store struct {
	ss store.VersionedDatabase
	sc map[string]*commitment.Database
}
