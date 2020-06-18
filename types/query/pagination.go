package query

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/types"
)

// defaultLimit is the default `limit` for queries
// if the `limit` is not supplied, paginate will use `defaultLimit`
const defaultLimit = 100

// Paginate does pagination of all the results in the PrefixStore based on the
// provided PageRequest. onResult should be used to do actual unmarshaling.
//
// Ex:
//  func (q BaseKeeper) QuerySome(c context.Context, req *types.QuerySomeRequest)
// 			(*types.QuerySomeResponse, error) {
//		prefixStore := prefix.NewStore(store, someRequestParam)
//		var results []Result
//		pageRes, err := query.Paginate(prefixStore, req.Page, func(key []byte, value []byte) error {
//			var result Result
//			err := Unmarshal(value, &balance)
//			...
//			results = append(results, result)
//			...
//		})
//		...
//
//		return &types.QuerySomeResponse{Results: results, Res: pageRes}, nil
//  }
func Paginate(
	prefixStore types.KVStore,
	req *PageRequest,
	onResult func(key []byte, value []byte) error,
) (*PageResponse, error) {
	offset := req.Offset
	key := req.Key
	limit := req.Limit

	if offset > 0 && key != nil {
		return nil, fmt.Errorf("invalid request, either offset or key is expected, got both")
	}

	if limit == 0 {
		limit = defaultLimit

		// count total results when the limit is zero/not supplied
		req.CountTotal = true
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

			err := onResult(iterator.Key(), iterator.Value())
			if err != nil {
				return nil, err
			}

			count++
		}

		return &PageResponse{
			NextKey: nextKey,
		}, nil
	}

	iterator := prefixStore.Iterator(nil, nil)
	defer iterator.Close()

	end := offset + limit

	var count uint64
	var nextKey []byte

	for ; iterator.Valid(); iterator.Next() {
		count++

		if count <= offset {
			continue
		}
		if count <= end {
			err := onResult(iterator.Key(), iterator.Value())
			if err != nil {
				return nil, err
			}
		} else if !req.CountTotal {
			nextKey = iterator.Key()
			break
		}
	}

	res := &PageResponse{NextKey: nextKey}
	if req.CountTotal {
		res.Total = count
	}

	return res, nil
}
