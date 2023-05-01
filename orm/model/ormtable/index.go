package ormtable

import (
	"context"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"cosmossdk.io/orm/encoding/ormkv"
	"cosmossdk.io/orm/model/ormlist"
	"cosmossdk.io/orm/types/kv"
)

// Index defines an index on a table. Index instances
// are stateless, with all state existing only in the store passed
// to index methods.
type Index interface {
	// List does iteration over the index with the provided prefix key and options.
	// Prefix key values must correspond in type to the index's fields and the
	// number of values provided cannot exceed the number of fields in the index,
	// although fewer values can be provided.
	List(ctx context.Context, prefixKey []interface{}, options ...ormlist.Option) (Iterator, error)

	// ListRange does range iteration over the index with the provided from and to
	// values and options.
	//
	// From and to values must correspond in type to the index's fields and the number of values
	// provided cannot exceed the number of fields in the index, although fewer
	// values can be provided.
	//
	// Range iteration can only be done for from and to values which are
	// well-ordered, meaning that any unordered components must be equal. Ex.
	// the bytes type is considered unordered, so a range iterator is created
	// over an index with a bytes field, both start and end must have the same
	// value for bytes.
	//
	// Range iteration is inclusive at both ends.
	ListRange(ctx context.Context, from, to []interface{}, options ...ormlist.Option) (Iterator, error)

	// DeleteBy deletes any entries which match the provided prefix key.
	DeleteBy(context context.Context, prefixKey ...interface{}) error

	// DeleteRange deletes any entries between the provided range keys.
	DeleteRange(context context.Context, from, to []interface{}) error

	// MessageType returns the protobuf message type of the index.
	MessageType() protoreflect.MessageType

	// Fields returns the canonical field names of the index.
	Fields() string

	doNotImplement()
}

// concreteIndex is used internally by table implementations.
type concreteIndex interface {
	Index
	ormkv.IndexCodec

	readValueFromIndexKey(context ReadBackend, primaryKey []protoreflect.Value, value []byte, message proto.Message) error
}

// UniqueIndex defines an unique index on a table.
type UniqueIndex interface {
	Index

	// Has returns true if the key values are present in the store for this index.
	Has(context context.Context, keyValues ...interface{}) (found bool, err error)

	// Get retrieves the message if one exists for the provided key values.
	Get(context context.Context, message proto.Message, keyValues ...interface{}) (found bool, err error)
}

type indexer interface {
	onInsert(store kv.Store, message protoreflect.Message) error
	onUpdate(store kv.Store, new, existing protoreflect.Message) error
	onDelete(store kv.Store, message protoreflect.Message) error
}
