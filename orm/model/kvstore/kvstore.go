package kvstore

import dbm "github.com/tendermint/tm-db"

type ReadStore interface {
	Get(key []byte) ([]byte, error)
	Has(key []byte) (bool, error)
	Iterator(start, end []byte) (Iterator, error)
	ReverseIterator(start, end []byte) (Iterator, error)
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

type Iterator = dbm.Iterator
