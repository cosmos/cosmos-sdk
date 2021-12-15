package ormtable

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/cosmos/cosmos-sdk/orm/encoding/ormkv"
	"github.com/cosmos/cosmos-sdk/orm/model/kvstore"
)

// Index defines an index on a table. Index instances
// are stateless, with all state existing only in the store passed
// to index methods.
type Index interface {

	// PrefixIterator returns a prefix iterator for the provided prefix. Prefix
	// can contain 0 or more values that must correspond to the fields in the index.
	PrefixIterator(store kvstore.ReadBackend, prefix []protoreflect.Value, options IteratorOptions) (Iterator, error)

	// RangeIterator returns a range iterator between the provided start and end.
	// Start and end can contain 0 or more values that must correspond to the fields in the index.
	// Range iterators can only be contained for start and end values which are
	// well-ordered, meaning that any unordered components must be equal. Ex.
	// the bytes type is considered unordered, so a range iterator is created
	// over an index with a bytes field, both start and end must have the same
	// value for bytes.
	RangeIterator(store kvstore.ReadBackend, start, end []protoreflect.Value, options IteratorOptions) (Iterator, error)

	// MessageType returns the protobuf message type of the index.
	MessageType() protoreflect.MessageType

	// GetFieldNames returns the field names of the index.
	GetFieldNames() []protoreflect.Name

	// CompareKeys the two keys against the underlying IndexCodec, returning a
	// negative value if key1 is less than key2, 0 if they are equal, and a
	// positive value otherwise.
	CompareKeys(key1, key2 []protoreflect.Value) int

	// IsFullyOrdered returns true if all of the fields in the index are
	// considered "well-ordered" in terms of sorted iteration.
	IsFullyOrdered() bool

	doNotImplement()
}

// concreteIndex is used internally by table implementations.
type concreteIndex interface {
	Index
	ormkv.IndexCodec

	readValueFromIndexKey(store kvstore.ReadBackend, primaryKey []protoreflect.Value, value []byte, message proto.Message) error
}

// UniqueIndex defines an unique index on a table.
type UniqueIndex interface {
	Index

	// Has returns true if the keyValues are present in the store for this index.
	Has(store kvstore.ReadBackend, keyValues []protoreflect.Value) (found bool, err error)

	// Get retrieves the message if one exists in the store for the provided keyValues.
	Get(store kvstore.ReadBackend, keyValues []protoreflect.Value, message proto.Message) (found bool, err error)
}

// IteratorOptions are options for creating an iterator.
type IteratorOptions struct {

	// Reverse specifies whether the iterator should be a reverse iterator.
	Reverse bool

	// Cursor is an optional value that can be used to start iteration
	// from a cursor returned by Iterator.Cursor() which can be used to
	// support pagination.
	Cursor Cursor
}

type indexer interface {
	onInsert(store kvstore.Writer, message protoreflect.Message) error
	onUpdate(store kvstore.Writer, new, existing protoreflect.Message) error
	onDelete(store kvstore.Writer, message protoreflect.Message) error
}
