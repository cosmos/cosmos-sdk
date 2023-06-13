package db

import (
	storetypes "cosmossdk.io/store/types"

	dbm "github.com/cosmos/cosmos-db"
)

var _ = (*storetypes.Iterator)(nil)

type AsStoreIter struct {
	dbm.Iterator
}

// DBToStoreIterator returns an iterator wrapping the given iterator so that it satisfies the
// (store/types).Iterator interface.
func ToStoreIterator(source dbm.Iterator) *AsStoreIter {
	ret := &AsStoreIter{Iterator: source}
	ret.Next() // The DB iterator must be primed before it can access the first element, because Next also returns the validity status
	return ret
}

func (it *AsStoreIter) Next()       { it.Iterator.Next() }
func (it *AsStoreIter) Valid() bool { return it.Iterator.Valid() }
