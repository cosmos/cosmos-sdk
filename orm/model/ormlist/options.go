// Package ormlist defines options for listing items from ORM indexes.
package ormlist

import (
	"github.com/cosmos/cosmos-sdk/orm/encoding/encodeutil"
	"github.com/cosmos/cosmos-sdk/orm/internal/listinternal"
)

// Option represents a list option.
type Option = listinternal.Option

// Start defines the values to use to start range iteration. It cannot be
// combined with Prefix.
//
// Values must correspond in type to the index's fields and the number of values
// provided cannot exceed the number of fields in the index, although fewer
// values can be provided.
//
// Range iteration can only be done for start and end values which are
// well-ordered, meaning that any unordered components must be equal. Ex.
// the bytes type is considered unordered, so a range iterator is created
// over an index with a bytes field, both start and end must have the same
// value for bytes.
func Start(values ...interface{}) Option {
	return listinternal.FuncOption(func(options *listinternal.Options) {
		options.Start = encodeutil.ValuesOf(values...)
	})
}

// End defines the values to use to end range iteration. It cannot be
// combined with Prefix.
//
// Values must correspond in type to the index's fields and the number of values
// provided cannot exceed the number of fields in the index, although fewer
// values can be provided.
//
// Range iteration can only be done for start and end values which are
// well-ordered, meaning that any unordered components must be equal. Ex.
// the bytes type is considered unordered, so a range iterator is created
// over an index with a bytes field, both start and end must have the same
// value for bytes.
func End(values ...interface{}) Option {
	return listinternal.FuncOption(func(options *listinternal.Options) {
		options.End = encodeutil.ValuesOf(values...)
	})
}

// Prefix defines values to use for prefix iteration. It cannot be used
// together with Start or End.
//
// Values must correspond in type to the index's fields and the number of values
// provided cannot exceed the number of fields in the index, although fewer
// values can be provided.
func Prefix(values ...interface{}) Option {
	return listinternal.FuncOption(func(options *listinternal.Options) {
		options.Prefix = encodeutil.ValuesOf(values...)
	})
}

// Reverse reverses the direction of iteration. If Reverse is
// provided twice, iteration will happen in the forward direction.
func Reverse() Option {
	return listinternal.FuncOption(func(options *listinternal.Options) {
		options.Reverse = !options.Reverse
	})
}

// Cursor specifies a cursor after which to restart iteration. Cursor values
// are returned by iterators and in pagination results.
func Cursor(cursor CursorT) Option {
	return listinternal.FuncOption(func(options *listinternal.Options) {
		options.Cursor = cursor
	})
}

// CursorT defines a cursor type.
type CursorT []byte
