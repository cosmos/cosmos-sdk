package query

import (
	"github.com/cosmos/cosmos-sdk/store/types"
)

func Paginate(
	prefixStore types.KVStore,
	req *PageRequest,
	onResult func(key []byte, value []byte) error,
) (*PageResponse, error) {
	if req.PageNum >= 1 {
		iterator := prefixStore.Iterator(req.PageKey, nil)
		defer iterator.Close()

		pageStart := req.PageNum * req.Limit
		pageEnd := pageStart + req.Limit
		var count uint64
		var nextPageKey []byte

		for ; iterator.Valid(); iterator.Next() {
			count++
			if count < pageStart {
				continue
			} else if count <= pageEnd {
				err := onResult(iterator.Key(), iterator.Value())
				if err != nil {
					return PageResponse{}, err
				}
			} else if !req.CountTotal {
				nextPageKey = iterator.Key()
				break
			}
		}

		res := PageResponse{NextPageKey: nextPageKey}
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
			if count == req.Limit {
				nextPageKey = iterator.Key()
				break
			}
			err := onResult(iterator.Key(), iterator.Value())
			if err != nil {
				return PageResponse{}, err
			}
			count++
		}
		return PageResponse{
			NextPageKey: nextPageKey,
		}, nil
	}
}
