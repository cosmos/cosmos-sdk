package kv

import sdkstore "github.com/cosmos/cosmos-sdk/store"

type ReadStore interface {
	Get(key []byte) ([]byte, error)
	Has(key []byte) (bool, error)
	Iterator(start, end []byte) Iterator
	ReverseIterator(start, end []byte) Iterator
}

type IndexCommitmentReadStore interface {
	ReadCommitmentStore() ReadStore
	ReadIndexStore() ReadStore
}

type Store interface {
	ReadStore
	Set(key, value []byte) error
	Delete(key []byte) error
}

type IndexCommitmentStore interface {
	IndexCommitmentReadStore
	CommitmentStore() Store
	IndexStore() Store
}

type Iterator = sdkstore.Iterator
