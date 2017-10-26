package sdk

import (
	"github.com/tendermint/go-wire/data"
)

// KVStore is a simple interface to get/set data
type KVStore interface {
	Set(key, value []byte)
	Get(key []byte) (value []byte)
}

//----------------------------------------

// Model grabs together key and value to allow easier return values
type Model struct {
	Key   data.Bytes
	Value data.Bytes
}

// SimpleDB allows us to do some basic range queries on a db
type SimpleDB interface {
	KVStore

	Has(key []byte) (has bool)
	Remove(key []byte) (value []byte) // returns old value if there was one

	// Start is inclusive, End is exclusive...
	// Thus List ([]byte{12, 13}, []byte{12, 14}) will return anything with
	// the prefix []byte{12, 13}
	List(start, end []byte, limit int) []Model
	First(start, end []byte) Model
	Last(start, end []byte) Model

	// Checkpoint returns the same state, but where writes
	// are buffered and don't affect the parent
	Checkpoint() SimpleDB

	// Commit will take all changes from the checkpoint and write
	// them to the parent.
	// Returns an error if this is not a child of this one
	Commit(SimpleDB) error

	// Discard will remove reference to this
	Discard()
}
