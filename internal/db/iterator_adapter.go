package db

import (
	dbm "github.com/cosmos/cosmos-sdk/db"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
)

var _ = (*storetypes.Iterator)(nil)

type dbAsStoreIter struct {
	dbm.Iterator
	valid bool
}

// DBToStoreIterator returns an iterator wrapping the given iterator so that it satisfies the
// (store/types).Iterator interface.
func DBToStoreIterator(source dbm.Iterator) *dbAsStoreIter {
	ret := &dbAsStoreIter{Iterator: source}
	ret.Next() // The DB iterator must be primed before it can access the first element, because Next also returns the validity status
	return ret
}

func (it *dbAsStoreIter) Next()       { it.valid = it.Iterator.Next() }
func (it *dbAsStoreIter) Valid() bool { return it.valid }
