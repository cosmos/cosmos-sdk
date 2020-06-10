package query

import (
	"github.com/cosmos/cosmos-sdk/store/types"
)

// Paginate does pagination of the results in the prefixStore based on the
// provided PageRequest. onResult should be used to do actual unmarshaling.
//
// Ex:
//		prefixStore := prefix.NewStore(store, someRequestParam)
//		var results []Result
//		pageRes, err := query.Paginate(accountStore, req.Page, func(key []byte, value []byte) error {
//			var result Result
//			err := Unmarshal(value, &balance)
//			results = append(results, result)
//			...
//		})
func Paginate(
	prefixStore types.KVStore,
	req *PageRequest,
	onResult func(key []byte, value []byte) error,
) (*PageResponse, error) {
	offset := req.Offset
	key := req.Key
	limit := req.Limit

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
	} else {
		iterator := prefixStore.Iterator(nil, nil)
		defer iterator.Close()

		end := offset + limit
		var count uint64
		var nextKey []byte

		for ; iterator.Valid(); iterator.Next() {
			count++

			if count < offset {
				continue
			} else if count <= end {
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
}
