package ormtable

import (
	"github.com/cosmos/cosmos-sdk/orm/internal/listinternal"

	queryv1beta1 "github.com/cosmos/cosmos-sdk/api/cosmos/base/query/v1beta1"
)

func paginate(it Iterator, options *listinternal.Options) Iterator {
	offset := int(options.Offset)
	limit := int(options.Limit)

	i := 0
	if offset != 0 {
		for ; i < offset; i++ {
			if !it.Next() {
				return &paginationIterator{
					Iterator: it,
					pageRes:  &queryv1beta1.PageResponse{Total: uint64(i)},
				}
			}
		}
	}

	done := limit + offset
	return &paginationIterator{
		Iterator:   it,
		pageRes:    nil,
		countTotal: options.CountTotal,
		i:          i,
		done:       done,
	}
}

type paginationIterator struct {
	Iterator
	pageRes    *queryv1beta1.PageResponse
	countTotal bool
	i          int
	done       int
}

func (it paginationIterator) Next() bool {
	if it.i >= it.done {
		it.pageRes = &queryv1beta1.PageResponse{}
		if it.Next() {
			it.pageRes.NextKey = it.Cursor()
			it.i++
		}
		if it.countTotal {
			for {
				it.i++
				if !it.Next() {
					it.pageRes.Total = uint64(it.i)
					return false
				}
			}
		}
		return false
	}

	it.i++
	return true
}

func (it paginationIterator) PageResponse() *queryv1beta1.PageResponse {
	return it.pageRes
}
