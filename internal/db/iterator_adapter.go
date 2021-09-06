package db

import dbm "github.com/cosmos/cosmos-sdk/db"

type dbAsStoreIter struct {
	dbm.Iterator
	valid bool
}

func DBToStoreIterator(source dbm.Iterator) *dbAsStoreIter {
	ret := &dbAsStoreIter{Iterator: source}
	ret.Next()
	return ret
}

func (it *dbAsStoreIter) Next()       { it.valid = it.Iterator.Next() }
func (it *dbAsStoreIter) Valid() bool { return it.valid }
