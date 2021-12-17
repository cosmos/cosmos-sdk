package ormtable

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/cosmos/cosmos-sdk/types/query"

	"google.golang.org/protobuf/proto"
)

// PaginationRequest is a request to the Paginate function and extends the
// options in query.PageRequest.
type PaginationRequest struct {
	*query.PageRequest

	// Prefix is an optional prefix to create a prefix iterator against this
	// index. It cannot be used together with Start and End.
	Prefix []protoreflect.Value

	// Start is an optional start value to create a range iterator against this
	// index. It cannot be used together with Prefix.
	Start []protoreflect.Value

	// End is an optional end value to create a range iterator against this
	// index. It cannot be used together with Prefix.
	End []protoreflect.Value

	// Filter is an optional filter function that can be used to filter
	// the results in the underlying iterator and should return true to include
	// an item in the result.
	Filter func(message proto.Message) bool
}

// PaginationResponse is a response from the Paginate function and extends the
// options in query.PageResponse.
type PaginationResponse struct {
	*query.PageResponse

	// Items are the items in this page.
	Items []proto.Message

	// HaveMore indicates whether there are more pages.
	HaveMore bool

	// Cursors returns a cursor for each item and can be used to implement
	// GraphQL connection edges.
	Cursors []Cursor
}

// Paginate retrieves a "page" of data from the provided index and store.
func Paginate(
	index Index,
	ctx context.Context,
	request *PaginationRequest,
) (*PaginationResponse, error) {
	offset := int(request.Offset)
	if len(request.Key) != 0 && offset > 0 {
		return nil, fmt.Errorf("can only specify one of cursor or offset")
	}

	iteratorOpts := IteratorOptions{
		Reverse: request.Reverse,
		Cursor:  request.Key,
	}
	var it Iterator
	var err error
	if request.Start != nil || request.End != nil {
		if request.Prefix != nil {
			return nil, fmt.Errorf("can either use Start/End or Prefix, not both")
		}

		it, err = index.RangeIterator(ctx, request.Start, request.End, iteratorOpts)
	} else {
		it, err = index.PrefixIterator(ctx, request.Prefix, iteratorOpts)
	}
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
					PageResponse: &query.PageResponse{Total: uint64(i)},
				}, nil
			}
		}
	}

	haveMore := false
	cursors := make([]Cursor, 0, limit)
	items := make([]proto.Message, 0, limit)
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

		if request.Filter != nil && !request.Filter(message) {
			continue
		}

		i++
		cursors = append(cursors, it.Cursor())
		items = append(items, message)
	}

	pageRes := &query.PageResponse{}
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
		Items:        items,
	}, nil
}
