package ormtable

import (
	"fmt"

	"google.golang.org/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/orm/model/ormiterator"
)

func Paginate(
	getIterator func(IteratorOptions) (ormiterator.Iterator, error),
	request *PaginationRequest,
) (*PaginationResponse, error) {
	if len(request.Cursor) != 0 && request.Offset > 0 {
		return nil, fmt.Errorf("can only specify one of cursor or offset")
	}

	it, err := getIterator(IteratorOptions{
		Reverse: request.Reverse,
		Cursor:  request.Cursor,
	})
	if err != nil {
		return nil, err
	}
	defer it.Close()

	limit := request.Limit
	if limit == 0 {
		return nil, fmt.Errorf("limit not specified")
	}

	haveMore := false
	var nodes []proto.Message
	var cursors []ormiterator.Cursor
	i := 0
	for {
		have, err := it.Next()
		if err != nil {
			return nil, err
		}

		if !have {
			break
		}

		if i == limit {
			haveMore = true
			if request.CountTotal {
				for {
					have, err = it.Next()
					if err != nil {
						return nil, err
					}
					if !have {
						break
					}
					i++
				}
			}
			break
		}

		node, err := it.GetMessage()
		if request.Filter != nil && !request.Filter(node) {
			continue
		}

		i++
		cursors = append(cursors, it.Cursor())
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, node)
	}

	res := &PaginationResponse{
		Nodes:    nodes,
		Cursors:  cursors,
		HaveMore: haveMore,
	}
	if request.CountTotal {
		res.TotalCount = i
	}
	return res, nil
}

type PaginationRequest struct {
	Limit      int
	Offset     int
	Reverse    bool
	CountTotal bool
	Cursor     ormiterator.Cursor
	Filter     func(proto.Message) bool
}

type PaginationResponse struct {
	Nodes      []proto.Message
	Cursors    []ormiterator.Cursor
	HaveMore   bool
	TotalCount int
}
