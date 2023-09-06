// Package ormlist defines options for listing items from ORM indexes.
package ormlist

import (
	"google.golang.org/protobuf/proto"

	queryv1beta1 "cosmossdk.io/api/cosmos/base/query/v1beta1"
	"cosmossdk.io/orm/internal/listinternal"
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
//
// Care should be taken when using Paginate together with Reverse and/or Cursor.
// In the case of combining Reverse and Paginate, if pageRequest.Reverse is
// true then iteration will proceed in the forward direction. This allows
// the default iteration direction for a query to be reverse with the option
// to switch this (to forward in this case) using PageRequest. If Cursor
// and Paginate are used together, whichever option is used first wins.
// If pageRequest is nil, this option will be a no-op so the caller does not
// need to do a nil check. This function defines no default limit, so if
// the caller does not define a limit, this will return all results. To
// specify a default limit use the DefaultLimit option.
func Paginate(pageRequest *queryv1beta1.PageRequest) Option {
	return listinternal.FuncOption(func(options *listinternal.Options) {
		if pageRequest == nil {
			return
		}

		if pageRequest.Reverse {
			// if the reverse is true we invert the direction of iteration,
			// meaning if iteration was already reversed we set it forward.
			options.Reverse = !options.Reverse
		}

		options.Cursor = pageRequest.Key
		options.Offset = pageRequest.Offset
		options.Limit = pageRequest.Limit
		options.CountTotal = pageRequest.CountTotal
	})
}

// DefaultLimit specifies a default limit for iteration. This option can be
// combined with Paginate to ensure that there is a default limit if none
// is specified in PageRequest.
func DefaultLimit(defaultLimit uint64) Option {
	return listinternal.FuncOption(func(options *listinternal.Options) {
		options.DefaultLimit = defaultLimit
	})
}

// CursorT defines a cursor type.
type CursorT []byte
