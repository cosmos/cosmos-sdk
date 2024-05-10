package store

import (
	"cosmossdk.io/core/store"
)

// Store defines the underlying storage engine of an app.
type Store interface {
	// StateLatest returns a readonly view over the latest
	// committed state of the store. Alongside the version
	// associated with it.
	StateLatest() (uint64, store.ReaderMap, error)

	// StateAt returns a readonly view over the provided
	// state. Must error when the version does not exist.
	StateAt(version uint64) (store.ReaderMap, error)
}
