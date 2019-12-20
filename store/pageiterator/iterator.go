package pageiterator

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/types"
)

// KVStorePrefixIteratorPaginated returns iterator over items in the selected page.
// Items iterated and skipped in ascending order.
func KVStorePrefixIteratorPaginated(kvs types.KVStore, prefix []byte, page, limit uint) types.Iterator {
	return KVStorePaginatedIterator(types.KVStorePrefixIterator(kvs, prefix), page, limit)
}

// KVStoreReversePrefixIteratorPaginated returns iterator over items in the selected page.
// Items iterated and skipped in descending order.
func KVStoreReversePrefixIteratorPaginated(kvs types.KVStore, prefix []byte, page, limit uint) types.Iterator {
	return KVStorePaginatedIterator(types.KVStoreReversePrefixIterator(kvs, prefix), page, limit)
}

// KVStorePaginatedIterator returns wrapper for provided iterator.
// Wrapper will enforce iteration
func KVStorePaginatedIterator(iterator types.Iterator, page, limit uint) types.Iterator {
	pi := &PaginatedIterator{
		Iterator: iterator,
		page:     page,
		limit:    limit,
	}
	pi.skip()
	return pi
}

// PaginatedIterator is a wrapper around Iterator that iterates over values starting for given page and limit.
type PaginatedIterator struct {
	types.Iterator

	page, limit uint // provided during initialization
	iterated    uint // incremented in a call to Next

}

func (pi *PaginatedIterator) skip() {
	for i := (pi.page - 1) * pi.limit; i > 0 && pi.Iterator.Valid(); i-- {
		pi.Iterator.Next()
	}
}

// Next will panic after limit is reached.
func (pi *PaginatedIterator) Next() {
	if pi.iterated == pi.limit {
		panic(fmt.Sprintf("PaginatedIterator reached limit %d", pi.limit))
	}
	pi.Iterator.Next()
	pi.iterated++
}

// Valid if below limit and underlying iterator is valid.
func (pi *PaginatedIterator) Valid() bool {
	if pi.iterated >= pi.limit {
		return false
	}
	return pi.Iterator.Valid()
}
