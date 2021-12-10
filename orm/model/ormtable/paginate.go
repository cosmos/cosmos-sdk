package ormtable

import (
	"fmt"

	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/cosmos/cosmos-sdk/orm/model/kvstore"

	"github.com/cosmos/cosmos-sdk/types/query"

	"google.golang.org/protobuf/proto"
)

type PaginationRequest struct {
	*query.PageRequest

	Prefix []protoreflect.Value
	Start  []protoreflect.Value
	End    []protoreflect.Value
	Filter func(message proto.Message) bool
}

func Paginate(
	index Index,
	store kvstore.IndexCommitmentReadStore,
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

		it, err = index.RangeIterator(store, request.Start, request.End, iteratorOpts)
	} else {
		it, err = index.PrefixIterator(store, request.Prefix, iteratorOpts)
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
				for it.Next() {
					i++
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
		pageRes.Total = uint64(i + 1)
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

// PaginationResponse contains extra pagination data that may be useful
// for services like GraphQL.
type PaginationResponse struct {
	*query.PageResponse
	Items    []proto.Message
	HaveMore bool
	Cursors  []Cursor
}
