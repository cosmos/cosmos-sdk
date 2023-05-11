package query

import (
	"fmt"

	"cosmossdk.io/store/types"
	proto "github.com/cosmos/gogoproto/proto"

	"github.com/cosmos/cosmos-sdk/codec"
)

// FilteredPaginate does pagination of all the results in the PrefixStore based on the
// provided PageRequest. onResult should be used to do actual unmarshaling and filter the results.
// If key is provided, the pagination uses the optimized querying.
// If offset is used, the pagination uses lazy filtering i.e., searches through all the records.
// The accumulate parameter represents if the response is valid based on the offset given.
// It will be false for the results (filtered) < offset  and true for `offset > accumulate <= end`.
// When accumulate is set to true the current result should be appended to the result set returned
// to the client.
func FilteredPaginate(
	prefixStore types.KVStore,
	pageRequest *PageRequest,
	onResult func(key, value []byte, accumulate bool) (bool, error),
) (*PageResponse, error) {
	// if the PageRequest is nil, use default PageRequest
	if pageRequest == nil {
		pageRequest = &PageRequest{}
	}

	key := pageRequest.Key
	limit := pageRequest.Limit
	countTotal := pageRequest.CountTotal

	if pageRequest.Offset > 0 && key != nil {
		return nil, fmt.Errorf("invalid request, either offset or key is expected, got both")
	}

	if limit == 0 {
		limit = DefaultLimit

		// count total results when the limit is zero/not supplied
		countTotal = true
	}

	if len(key) == 0 {
		key = nil
	}

	iterator := getIterator(prefixStore, key, pageRequest.Reverse)
	defer iterator.Close()

	if len(key) != 0 {
		accumulateFn := func(_ uint64) bool { return true }
		_, nextKey, err := iterateResults(iterator, limit, false, onResult, accumulateFn)
		if err != nil {
			return nil, err
		}

		return &PageResponse{
			NextKey: nextKey,
		}, nil
	}

	end := pageRequest.Offset + limit
	accumalateFn := func(numHits uint64) bool { return numHits >= pageRequest.Offset && numHits < end }
	numHits, nextKey, err := iterateResults(iterator, end+1, countTotal, onResult, accumalateFn)
	if err != nil {
		return nil, err
	}

	res := &PageResponse{NextKey: nextKey}
	if countTotal {
		res.Total = numHits
	}

	return res, nil
}

func iterateResults(iterator types.Iterator, limit uint64, countTotal bool, onResult func(key, value []byte, shouldAccumulate bool) (bool, error), accumulateFn func(numHits uint64) bool) (numHits uint64, nextKey []byte, err error) {
	for ; iterator.Valid(); iterator.Next() {
		if numHits == limit {
			if nextKey == nil {
				nextKey = iterator.Key()
			}

			if !countTotal {
				break
			}
		}

		if iterator.Error() != nil {
			return numHits, nil, iterator.Error()
		}

		hit, err := onResult(iterator.Key(), iterator.Value(), accumulateFn(numHits))
		if err != nil {
			return numHits, nil, err
		}

		if hit {
			numHits++
		}
	}

	return numHits, nextKey, nil
}

// GenericFilteredPaginate does pagination of all the results in the PrefixStore based on the
// provided PageRequest. `onResult` should be used to filter or transform the results.
// `c` is a constructor function that needs to return a new instance of the type T (this is to
// workaround some generic pitfalls in which we can't instantiate a T struct inside the function).
// If key is provided, the pagination uses the optimized querying.
// If offset is used, the pagination uses lazy filtering i.e., searches through all the records.
// The resulting slice (of type F) can be of a different type than the one being iterated through
// (type T), so it's possible to do any necessary transformation inside the onResult function.
func GenericFilteredPaginate[T, F proto.Message](
	cdc codec.BinaryCodec,
	prefixStore types.KVStore,
	pageRequest *PageRequest,
	onResult func(key []byte, value T) (F, error),
	constructor func() T,
) ([]F, *PageResponse, error) {
	// if the PageRequest is nil, use default PageRequest
	if pageRequest == nil {
		pageRequest = &PageRequest{}
	}

	offset := pageRequest.Offset
	key := pageRequest.Key
	limit := pageRequest.Limit
	countTotal := pageRequest.CountTotal
	reverse := pageRequest.Reverse
	results := []F{}

	if offset > 0 && key != nil {
		return results, nil, fmt.Errorf("invalid request, either offset or key is expected, got both")
	}

	if limit == 0 {
		limit = DefaultLimit

		// count total results when the limit is zero/not supplied
		countTotal = true
	}

	if len(key) != 0 {
		iterator := getIterator(prefixStore, key, reverse)
		defer iterator.Close()

		var (
			numHits uint64
			nextKey []byte
		)

		for ; iterator.Valid(); iterator.Next() {
			if numHits == limit {
				nextKey = iterator.Key()
				break
			}

			if iterator.Error() != nil {
				return nil, nil, iterator.Error()
			}

			protoMsg := constructor()

			err := cdc.Unmarshal(iterator.Value(), protoMsg)
			if err != nil {
				return nil, nil, err
			}

			val, err := onResult(iterator.Key(), protoMsg)
			if err != nil {
				return nil, nil, err
			}

			if proto.Size(val) != 0 {
				results = append(results, val)
				numHits++
			}
		}

		return results, &PageResponse{
			NextKey: nextKey,
		}, nil
	}

	iterator := getIterator(prefixStore, nil, reverse)
	defer iterator.Close()

	end := offset + limit

	var (
		numHits uint64
		nextKey []byte
	)

	for ; iterator.Valid(); iterator.Next() {
		if iterator.Error() != nil {
			return nil, nil, iterator.Error()
		}

		protoMsg := constructor()

		err := cdc.Unmarshal(iterator.Value(), protoMsg)
		if err != nil {
			return nil, nil, err
		}

		val, err := onResult(iterator.Key(), protoMsg)
		if err != nil {
			return nil, nil, err
		}

		if proto.Size(val) != 0 {
			// Previously this was the "accumulate" flag
			if numHits >= offset && numHits < end {
				results = append(results, val)
			}
			numHits++
		}

		if numHits == end+1 {
			if nextKey == nil {
				nextKey = iterator.Key()
			}

			if !countTotal {
				break
			}
		}
	}

	res := &PageResponse{NextKey: nextKey}
	if countTotal {
		res.Total = numHits
	}

	return results, res, nil
}
