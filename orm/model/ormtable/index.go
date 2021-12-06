package ormtable

import (
	"github.com/cosmos/cosmos-sdk/orm/encoding/ormkv"
	"github.com/cosmos/cosmos-sdk/orm/model/kvstore"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type Index interface {
	PrefixIterator(store kvstore.IndexCommitmentReadStore, prefix []protoreflect.Value, options IteratorOptions) (Iterator, error)
	RangeIterator(store kvstore.IndexCommitmentReadStore, start, end []protoreflect.Value, options IteratorOptions) (Iterator, error)
	ReadValueFromIndexKey(store kvstore.IndexCommitmentReadStore, primaryKey []protoreflect.Value, value []byte, message proto.Message) error

	MessageType() protoreflect.MessageType
	GetFieldNames() []protoreflect.Name
	IsFullyOrdered() bool
	CompareKeys(key1, key2 []protoreflect.Value) int

	doNotImplement()
}

type concreteIndex interface {
	Index
	ormkv.IndexCodec
}

type UniqueIndex interface {
	Index

	Has(store kvstore.IndexCommitmentReadStore, keyValues []protoreflect.Value) (found bool, err error)
	Get(store kvstore.IndexCommitmentReadStore, keyValues []protoreflect.Value, message proto.Message) (found bool, err error)
}

type IteratorOptions struct {
	Reverse bool
	Cursor  []byte
}

type Indexer interface {
	Index

	OnCreate(store kvstore.Store, message protoreflect.Message) error
	OnUpdate(store kvstore.Store, new, existing protoreflect.Message) error
	OnDelete(store kvstore.Store, message protoreflect.Message) error

	doNotImplement()
}
