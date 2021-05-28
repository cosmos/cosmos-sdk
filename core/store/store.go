package store

import (
	"context"

	dbm "github.com/tendermint/tm-db"
)

type BasicKVStore interface {
	// Get returns nil iff key doesn't exist. Panics on nil key.
	Get(key []byte) []byte

	// Has checks if a key exists. Panics on nil key.
	Has(key []byte) bool

	// Set sets the key. Panics on nil key or value.
	Set(key, value []byte)

	// Delete deletes the key. Panics on nil key.
	Delete(key []byte)
}

type KVStore interface {
	BasicKVStore

	// Iterator over a domain of keys in ascending order. End is exclusive.
	// Start must be less than end, or the Iterator is invalid.
	// Iterator must be closed by caller.
	// To iterate over entire domain, use store.Iterator(nil, nil)
	// CONTRACT: No writes may happen within a domain while an iterator exists over it.
	// Exceptionally allowed for cachekv.Store, safe to write in the modules.
	Iterator(start, end []byte) Iterator

	// Iterator over a domain of keys in descending order. End is exclusive.
	// Start must be less than end, or the Iterator is invalid.
	// Iterator must be closed by caller.
	// CONTRACT: No writes may happen within a domain while an iterator exists over it.
	// Exceptionally allowed for cachekv.Store, safe to write in the modules.
	ReverseIterator(start, end []byte) Iterator
}

// Iterator is an alias db's Iterator for convenience.
type Iterator = dbm.Iterator

type StoreKey interface {
	Open(ctx context.Context) KVStore
}

type KVStoreKey interface {
	StoreKey
	kvStoreKey()
}

type LowLevelSCStoreKey interface {
	Open(ctx context.Context) BasicKVStore
	scStoreKey()
}

type LowLevelSSStoreKey interface {
	StoreKey
	ssStoreKey()
}
