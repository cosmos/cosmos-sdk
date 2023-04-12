package ormtable

import (
	"math"

	"github.com/cosmos/cosmos-sdk/orm/internal/listinternal"

	queryv1beta1 "cosmossdk.io/api/cosmos/base/query/v1beta1"
)

func paginate(it Iterator, options *listinternal.Options) Iterator {
	offset := int(options.Offset)
	limit := int(options.Limit)
	if limit == 0 {
		limit = int(options.DefaultLimit)
	}

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

	var done int
	if limit != 0 {
		done = limit + offset
	} else {
		done = math.MaxInt
	}

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

func (it *paginationIterator) Next() bool {
	if it.i >= it.done {
		it.pageRes = &queryv1beta1.PageResponse{}
		cursor := it.Cursor()
		next := it.Iterator.Next()
		if next {
			it.pageRes.NextKey = cursor
			it.i++
		}
		if it.countTotal {
			// once it.Iterator.Next() returns false, another call to it will panic.
			// we check next here to ensure we do not call it again.
			if next {
				for {
					if !it.Iterator.Next() {
						it.pageRes.Total = uint64(it.i)
						return false
					}
					it.i++
				}
			} else {
				// when next is false, the iterator can no longer move forward,
				// so the index == total entries.
				it.pageRes.Total = uint64(it.i)
			}
		}
		return false
	}

	ok := it.Iterator.Next()
	if ok {
		it.i++
		return true
	}

	it.pageRes = &queryv1beta1.PageResponse{
		Total: uint64(it.i),
	}
	return false
}

func (it paginationIterator) PageResponse() *queryv1beta1.PageResponse {
	return it.pageRes
}
