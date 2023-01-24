package query

import (
	"context"
	"fmt"

	"cosmossdk.io/collections"
)

// Collection defines the minimum required API of a collection
// to work with pagination.
type Collection[K, V any] interface {
	// IterateRaw allows to iterate over a raw set of byte keys.
	IterateRaw(ctx context.Context, start, end []byte, order collections.Order) (collections.Iterator[K, V], error)
	// KeyCodec exposes the KeyCodec of a collection, required to encode a collection key from and to bytes
	// for pagination request and response.
	KeyCodec() collections.KeyCodec[K]
}

// CollectionPaginate follows the same behaviour as Paginate but works on a Collection.
func CollectionPaginate[K, V any, C Collection[K, V]](
	ctx context.Context,
	coll C,
	pageReq *PageRequest,
) ([]collections.KeyValue[K, V], *PageResponse, error) {
	return CollectionFilteredPaginate[K, V](ctx, coll, pageReq, nil)
}

// CollectionFilteredPaginate works in the same way as FilteredPaginate but for collection types.
// A nil predicateFunc means no filtering is applied and results are collected as is.
func CollectionFilteredPaginate[K, V any, C Collection[K, V]](
	ctx context.Context,
	coll C,
	pageReq *PageRequest,
	predicateFunc func(key K, value V) (include bool),
) ([]collections.KeyValue[K, V], *PageResponse, error) {
	if pageReq == nil {
		pageReq = &PageRequest{}
	}

	offset := pageReq.Offset
	key := pageReq.Key
	limit := pageReq.Limit
	countTotal := pageReq.CountTotal
	reverse := pageReq.Reverse

	if offset > 0 && key != nil {
		return nil, nil, fmt.Errorf("invalid request, either offset or key is expected, got both")
	}

	if limit == 0 {
		limit = DefaultLimit
		countTotal = true
	}

	if len(key) != 0 {
		return collFilteredPaginateByKey(ctx, coll, key, reverse, limit, predicateFunc)
	}

	return collFilteredPaginateNoKey(ctx, coll, reverse, offset, limit, countTotal, predicateFunc)
}

// collFilteredPaginateNoKey applies the provided pagination on the collection when the starting key is not set.
// If predicateFunc is nil no filtering is applied.
func collFilteredPaginateNoKey[K, V any, C Collection[K, V]](
	ctx context.Context,
	coll C,
	reverse bool,
	offset uint64,
	limit uint64,
	countTotal bool,
	predicateFunc func(K, V) bool,
) ([]collections.KeyValue[K, V], *PageResponse, error) {
	iterator, err := getCollIter[K, V](ctx, coll, nil, reverse)
	if err != nil {
		return nil, nil, err
	}
	defer iterator.Close()
	// we advance the iter equal to the provided offset
	if !advanceIter(iterator, offset) {
		return nil, nil, collections.ErrInvalidIterator
	}

	var (
		count   uint64
		nextKey []byte
		results []collections.KeyValue[K, V]
	)

	for ; iterator.Valid(); iterator.Next() {
		switch {
		// first case, we still haven't found all the results up to the limit
		case count < limit:
			kv, err := iterator.KeyValue()
			if err != nil {
				return nil, nil, err
			}
			// if no predicate function is specified then we just include the result
			if predicateFunc == nil {
				results = append(results, kv)
				count++
				// if predicate function is defined we check if the result matches the filtering criteria
			} else if predicateFunc(kv.Key, kv.Value) {
				results = append(results, kv)
				count++
			}
		// second case, we found all the objects specified within the limit
		case count == limit:
			key, err := iterator.Key()
			if err != nil {
				return nil, nil, err
			}
			nextKey, err = encodeCollKey[K, V](coll, key)
			if err != nil {
				return nil, nil, err
			}
			// if count total was not specified, we return the next key only
			if !countTotal {
				return results, &PageResponse{
					NextKey: nextKey,
				}, nil
			}
			// otherwise we fallthrough the third case
			fallthrough
		// this is the case in which we found all the required results
		// but we need to count how many possible results exist in total.
		// so we keep increasing the count until the iterator is fully consumed.
		case count > limit:
			count++
		}
	}
	return results, &PageResponse{
		NextKey: nextKey,
		Total:   count + offset,
	}, nil
}

func advanceIter[I interface {
	Next()
	Valid() bool
}](iter I, offset uint64,
) bool {
	for i := uint64(0); i < offset; i++ {
		if !iter.Valid() {
			return false
		}
		iter.Next()
	}
	return true
}

// collFilteredPaginateByKey paginates a collection when a starting key
// is provided in the PageRequest. Predicate is applied only if not nil.
func collFilteredPaginateByKey[K, V any, C Collection[K, V]](
	ctx context.Context,
	coll C,
	key []byte,
	reverse bool,
	limit uint64,
	predicateFunc func(K, V) bool,
) ([]collections.KeyValue[K, V], *PageResponse, error) {
	iterator, err := getCollIter[K, V](ctx, coll, key, reverse)
	if err != nil {
		return nil, nil, err
	}
	defer iterator.Close()

	var (
		count   uint64
		nextKey []byte
		results []collections.KeyValue[K, V]
	)

	for ; iterator.Valid(); iterator.Next() {
		// if we reached the specified limit
		// then we get the next key, and we exit the iteration.
		if count == limit {
			concreteKey, err := iterator.Key()
			if err != nil {
				return nil, nil, err
			}

			nextKey, err = encodeCollKey[K, V](coll, concreteKey)
			if err != nil {
				return nil, nil, err
			}
			break
		}

		kv, err := iterator.KeyValue()
		if err != nil {
			return nil, nil, err
		}
		// if no predicate is specified then we just append the result
		if predicateFunc == nil {
			results = append(results, kv)
			count++
			// if predicate is applied we execute the predicate function
			// and append only if predicateFunc yields true.
		} else if predicateFunc(kv.Key, kv.Value) {
			results = append(results, kv)
			count++
		}
	}

	return results, &PageResponse{
		NextKey: nextKey,
	}, nil
}

// todo maybe move to collections?
func encodeCollKey[K, V any](coll Collection[K, V], key K) ([]byte, error) {
	buffer := make([]byte, coll.KeyCodec().Size(key))
	_, err := coll.KeyCodec().Encode(buffer, key)
	return buffer, err
}

func getCollIter[K, V any](ctx context.Context, coll Collection[K, V], start []byte, reverse bool) (collections.Iterator[K, V], error) {
	if reverse {
		return coll.IterateRaw(ctx, nil, start, collections.OrderDescending)
	}
	return coll.IterateRaw(ctx, start, nil, collections.OrderAscending)
}
