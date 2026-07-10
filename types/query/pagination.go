package query

import (
	"fmt"
	"math"

	db "github.com/cosmos/cosmos-db"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/store/v2/types"
)

// DefaultPage is the default `page` number for queries.
// If the `page` number is not supplied, `DefaultPage` will be used.
const DefaultPage = 1

// DefaultLimit is the default `limit` for queries if the `limit` is not supplied.
const DefaultLimit = 100

// MaxLimit is the maximum allowed `limit` for queries. Any caller-supplied
// value exceeding this is silently capped to MaxLimit to prevent DoS via
// unbounded store iteration.
const MaxLimit = 10_000

// ParsePagination validates PageRequest and returns page number & limit.
func ParsePagination(pageReq *PageRequest) (page, limit int, err error) {
	offset := 0
	limit = DefaultLimit

	if pageReq != nil {
		offset = int(pageReq.Offset)

		// Cap on the uint64 value before casting to int: pageReq.Limit values
		// above math.MaxInt64 (e.g. ^uint64(0)) would otherwise wrap negative
		// during the cast and be rejected below instead of capped to MaxLimit.
		if pageReq.Limit > MaxLimit {
			limit = MaxLimit
		} else {
			limit = int(pageReq.Limit)
		}
	}
	if offset < 0 {
		return 1, 0, status.Error(codes.InvalidArgument, "offset must greater than 0")
	}

	if limit < 0 {
		return 1, 0, status.Error(codes.InvalidArgument, "limit must greater than 0")
	} else if limit == 0 {
		limit = DefaultLimit
	} else if limit > MaxLimit {
		limit = MaxLimit
	}

	page = offset/limit + 1

	return page, limit, nil
}

// Paginate does pagination of all the results in the PrefixStore based on the
// provided PageRequest. onResult should be used to do actual unmarshaling.
func Paginate(
	prefixStore types.KVStore,
	pageRequest *PageRequest,
	onResult func(key, value []byte) error,
) (*PageResponse, error) {
	pageRequest = initPageRequestDefaults(pageRequest)

	if pageRequest.Offset > 0 && pageRequest.Key != nil {
		return nil, fmt.Errorf("invalid request, either offset or key is expected, got both")
	}

	iterator := getIterator(prefixStore, pageRequest.Key, pageRequest.Reverse)
	defer iterator.Close()

	var count uint64
	var nextKey []byte

	if len(pageRequest.Key) != 0 {
		for ; iterator.Valid(); iterator.Next() {
			if count == pageRequest.Limit {
				nextKey = iterator.Key()
				break
			}
			if iterator.Error() != nil {
				return nil, iterator.Error()
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

	end := pageRequest.Offset + pageRequest.Limit
	if end < pageRequest.Offset {
		// Saturate to MaxUint64 when offset+limit overflows. Without this,
		// a caller passing an absurdly large limit would wrap end back to a
		// small number and the loop would return zero results.
		end = math.MaxUint64
	}

	for ; iterator.Valid(); iterator.Next() {
		count++

		if count <= pageRequest.Offset {
			continue
		}
		if count <= end {
			err := onResult(iterator.Key(), iterator.Value())
			if err != nil {
				return nil, err
			}
		} else if count == end+1 {
			nextKey = iterator.Key()

			if !pageRequest.CountTotal {
				break
			}
		}
		if iterator.Error() != nil {
			return nil, iterator.Error()
		}
	}

	res := &PageResponse{NextKey: nextKey}
	if pageRequest.CountTotal {
		res.Total = count
	}

	return res, nil
}

func getIterator(prefixStore types.KVStore, start []byte, reverse bool) db.Iterator {
	if reverse {
		var end []byte
		if start != nil {
			itr := prefixStore.Iterator(start, nil)
			defer itr.Close()
			if itr.Valid() {
				itr.Next()
				end = itr.Key()
			}
		}
		return prefixStore.ReverseIterator(nil, end)
	}
	return prefixStore.Iterator(start, nil)
}

// initPageRequestDefaults initializes a PageRequest's defaults when those are not set.
func initPageRequestDefaults(pageRequest *PageRequest) *PageRequest {
	// if the PageRequest is nil, use default PageRequest
	if pageRequest == nil {
		pageRequest = &PageRequest{}
	}

	pageRequestCopy := *pageRequest
	if len(pageRequestCopy.Key) == 0 {
		pageRequestCopy.Key = nil
	}

	if pageRequestCopy.Limit == 0 {
		pageRequestCopy.Limit = DefaultLimit

		// count total results when the limit is zero/not supplied
		pageRequestCopy.CountTotal = true
	} else if pageRequestCopy.Limit > MaxLimit {
		pageRequestCopy.Limit = MaxLimit
	}

	return &pageRequestCopy
}
