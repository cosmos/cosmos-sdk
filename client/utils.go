package client

import (
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/spf13/pflag"
)

// Paginate returns the correct starting and ending index for a paginated query,
// given that client provides a desired page and limit of objects and the handler
// provides the total number of objects. The start page is assumed to be 1-indexed.
// If the start page is invalid, non-positive values are returned signaling the
// request is invalid; it returns non-positive values if limit is non-positive and
// defLimit is negative.
func Paginate(numObjs, page, limit, defLimit int) (start, end int) {
	if page <= 0 {
		// invalid start page
		return -1, -1
	}

	// fallback to default limit if supplied limit is invalid
	if limit <= 0 {
		if defLimit < 0 {
			// invalid default limit
			return -1, -1
		}
		limit = defLimit
	}

	start = (page - 1) * limit
	end = limit + start

	if end >= numObjs {
		end = numObjs
	}

	if start >= numObjs {
		// page is out of bounds
		return -1, -1
	}

	return start, end
}

// ReadPageRequest reads and builds the necessary page request flags for pagination.
func ReadPageRequest(flagSet *pflag.FlagSet) (*query.PageRequest, error) {
	pageKey, err := flagSet.GetString(flags.FlagPageKey)
	if err != nil {
		return nil, err
	}

	offset, err := flagSet.GetUint64(flags.FlagOffset)
	if err != nil {
		return nil, err
	}

	limit, err := flagSet.GetUint64(flags.FlagLimit)
	if err != nil {
		return nil, err
	}

	countTotal, err := flagSet.GetBool(flags.FlagCountTotal)
	if err != nil {
		return nil, err
	}

	return &query.PageRequest{
		Key:        []byte(pageKey),
		Offset:     offset,
		Limit:      limit,
		CountTotal: countTotal,
	}, nil
}
