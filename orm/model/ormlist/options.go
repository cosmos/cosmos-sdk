// Package ormlist defines options for listing items from ORM indexes.
package ormlist

import (
	"google.golang.org/protobuf/proto"

	queryv1beta1 "github.com/cosmos/cosmos-sdk/api/cosmos/base/query/v1beta1"

	"github.com/cosmos/cosmos-sdk/orm/internal/listinternal"
)

// Option represents a list option.
type Option = listinternal.Option

// Reverse reverses the direction of iteration. If Reverse is
// provided twice, iteration will happen in the forward direction.
func Reverse() Option {
	return listinternal.FuncOption(func(options *listinternal.Options) {
		options.Reverse = !options.Reverse
	})
}

// Filter returns an option which applies a filter function to each item
// and skips over it when the filter function returns false.
func Filter(filterFn func(message proto.Message) bool) Option {
	return listinternal.FuncOption(func(options *listinternal.Options) {
		options.Filter = filterFn
	})
}

// Cursor specifies a cursor after which to restart iteration. Cursor values
// are returned by iterators and in pagination results.
func Cursor(cursor CursorT) Option {
	return listinternal.FuncOption(func(options *listinternal.Options) {
		options.Cursor = cursor
	})
}

// Paginate paginates iterator output based on the provided page request.
// The Iterator.PageRequest value on the returned iterator will be non-nil
// after Iterator.Next() returns false when this option is provided.
// Care should be taken when using Paginate together with Reverse and/or Cursor
// and generally this should be avoided.
func Paginate(pageRequest *queryv1beta1.PageRequest) Option {
	return listinternal.FuncOption(func(options *listinternal.Options) {
		if pageRequest.Reverse {
			options.Reverse = !options.Reverse
		}

		options.Cursor = pageRequest.Key
		options.Offset = pageRequest.Offset
		options.Limit = pageRequest.Limit
		options.CountTotal = pageRequest.CountTotal
	})
}

// CursorT defines a cursor type.
type CursorT []byte
