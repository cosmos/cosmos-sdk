package ormtable

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/proto"

	queryv1beta1 "github.com/cosmos/cosmos-sdk/api/cosmos/base/query/v1beta1"
	"github.com/cosmos/cosmos-sdk/orm/model/ormlist"
)

// PaginationResponse is a response from the Paginate function and extends the
// options in query.PageResponse.
type PaginationResponse struct {
	*queryv1beta1.PageResponse

	// HaveMore indicates whether there are more pages.
	HaveMore bool

	// Cursors returns a cursor for each item and can be used to implement
	// GraphQL connection edges.
	Cursors []ormlist.CursorT
}

// Paginate retrieves a "page" of data from the provided index and context.
func Paginate(
	index Index,
	ctx context.Context,
	request *queryv1beta1.PageRequest,
	onItem func(proto.Message),
	options ...ormlist.Option,
) (*PaginationResponse, error) {
	offset := int(request.Offset)
	if len(request.Key) != 0 {
		if offset > 0 {
			return nil, fmt.Errorf("can only specify one of cursor or offset")
		}

		options = append(options, ormlist.Cursor(request.Key))
	}

	if request.Reverse {
		options = append(options, ormlist.Reverse())
	}

	it, err := index.Iterator(ctx, options...)
	if err != nil {
		return nil, err
	}
	defer it.Close()

	limit := int(request.Limit)
	if limit == 0 {
		return nil, fmt.Errorf("limit not specified")
	}

	i := 0
	if offset != 0 {
		for ; i < offset; i++ {
			if !it.Next() {
				return &PaginationResponse{
					PageResponse: &queryv1beta1.PageResponse{Total: uint64(i)},
				}, nil
			}
		}
	}

	haveMore := false
	cursors := make([]ormlist.CursorT, 0, limit)
	done := limit + offset
	for it.Next() {
		if i == done {
			haveMore = true
			if request.CountTotal {
				for {
					i++
					if !it.Next() {
						break
					}
				}
			}
			break
		}

		message, err := it.GetMessage()
		if err != nil {
			return nil, err
		}

		i++
		cursors = append(cursors, it.Cursor())
		onItem(message)
	}

	pageRes := &queryv1beta1.PageResponse{}
	if request.CountTotal {
		pageRes.Total = uint64(i)
	}
	n := len(cursors)
	if n != 0 {
		pageRes.NextKey = cursors[n-1]
	}
	return &PaginationResponse{
		PageResponse: pageRes,
		HaveMore:     haveMore,
		Cursors:      cursors,
	}, nil
}
