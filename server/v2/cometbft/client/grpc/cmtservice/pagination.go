package cmtservice

import (
	math "math"

	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// DefaultPage is the default `page` number for queries.
// If the `page` number is not supplied, `DefaultPage` will be used.
const DefaultPage = 1

// DefaultLimit is the default `limit` for queries
// if the `limit` is not supplied, paginate will use `DefaultLimit`
const DefaultLimit = 100

// PaginationMaxLimit is the maximum limit the paginate function can handle
// which equals the maximum value that can be stored in uint64
var PaginationMaxLimit uint64 = math.MaxUint64

// ParsePagination validate PageRequest and returns page number & limit.
func ParsePagination(pageReq *PageRequest) (page, limit int, err error) {
	offset := 0
	limit = DefaultLimit

	if pageReq != nil {
		offset = int(pageReq.Offset)
		limit = int(pageReq.Limit)
	}
	if offset < 0 {
		return 1, 0, status.Error(codes.InvalidArgument, "offset must greater than 0")
	}

	if limit < 0 {
		return 1, 0, status.Error(codes.InvalidArgument, "limit must greater than 0")
	} else if limit == 0 {
		limit = DefaultLimit
	}

	page = offset/limit + 1

	return page, limit, nil
}
