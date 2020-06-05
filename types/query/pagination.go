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
	pageNum := req.PageNum
	limit := req.Limit

	if pageNum >= 1 {
		iterator := prefixStore.Iterator(req.PageKey, nil)
		defer iterator.Close()

		pageStart := pageNum * limit
		pageEnd := pageStart + limit
		var count uint64
		var nextPageKey []byte

		for ; iterator.Valid(); iterator.Next() {
			count++

			if count < pageStart {
				continue
			} else if count <= pageEnd {
				err := onResult(iterator.Key(), iterator.Value())
				if err != nil {
					return nil, err
				}
			} else if !req.CountTotal {
				nextPageKey = iterator.Key()
				break
			}
		}

		res := &PageResponse{NextPageKey: nextPageKey}
		if req.CountTotal {
			res.Total = count
		}

		return res, nil
	} else {
		iterator := prefixStore.Iterator(req.PageKey, nil)
		defer iterator.Close()

		var count uint64
		var nextPageKey []byte

		for ; iterator.Valid(); iterator.Next() {
			if count == limit {
				nextPageKey = iterator.Key()
				break
			}

			err := onResult(iterator.Key(), iterator.Value())
			if err != nil {
				return nil, err
			}

			count++
		}

		return &PageResponse{
			NextPageKey: nextPageKey,
		}, nil
	}
}
