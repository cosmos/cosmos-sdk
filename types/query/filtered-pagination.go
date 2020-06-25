package query

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/types"
)

// defaultLimit is the default `limit` for queries
// if the `limit` is not supplied, FilteredPaginate will use `defaultLimit`
const defaultLimit = 100

// FilteredPaginate does pagination of all the results in the PrefixStore based on the
// provided PageRequest. onResult should be used to do actual unmarshaling and filter the results.
// if key is provided, the pagination uses the optimized querying
// if offset is used, the pagination uses lazy filtering i.e., searches through all the records
func FilteredPaginate(
	prefixStore types.KVStore,
	req *PageRequest,
	onResult func(key []byte, value []byte) (bool, error),
) (*PageResponse, error) {
	offset := req.Offset
	key := req.Key
	limit := req.Limit
	countTotal := req.CountTotal

	if offset > 0 && key != nil {
		return nil, fmt.Errorf("invalid request, either offset or key is expected, got both")
	}

	if limit == 0 {
		limit = defaultLimit

		// count total results when the limit is zero/not supplied
		countTotal = true
	}

	if len(key) != 0 {
		iterator := prefixStore.Iterator(key, nil)
		defer iterator.Close()

		var count uint64
		var nextKey []byte

		for ; iterator.Valid(); iterator.Next() {
			if count == limit {
				nextKey = iterator.Key()
				break
			}

			hit, err := onResult(iterator.Key(), iterator.Value())
			if err != nil {
				return nil, err
			}

			if hit {
				count++
			}
		}

		return &PageResponse{
			NextKey: nextKey,
		}, nil
	}

	iterator := prefixStore.Iterator(nil, nil)
	defer iterator.Close()

	end := offset + limit

	var count uint64
	var numHits uint64
	var nextKey []byte

	for ; iterator.Valid(); iterator.Next() {
		count++

		if numHits <= end {
			hit, err := onResult(iterator.Key(), iterator.Value())
			if err != nil {
				return nil, err
			}

			if hit {
				numHits++
			}
		} else if !countTotal {
			nextKey = iterator.Key()
			break
		}
	}

	res := &PageResponse{NextKey: nextKey}
	if countTotal {
		res.Total = count
	}

	return res, nil
}
