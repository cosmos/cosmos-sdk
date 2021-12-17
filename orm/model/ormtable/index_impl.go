package ormtable

import (
	"context"

	"github.com/cosmos/cosmos-sdk/orm/model/kvstore"
	"github.com/cosmos/cosmos-sdk/orm/types/ormerrors"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/cosmos/cosmos-sdk/orm/encoding/ormkv"
)

// indexKeyIndex implements Index for a regular IndexKey.
type indexKeyIndex struct {
	*ormkv.IndexKeyCodec
	primaryKey     *primaryKeyIndex
	getReadBackend func(context.Context) (ReadBackend, error)
}

func (s indexKeyIndex) PrefixIterator(context context.Context, prefix []protoreflect.Value, options IteratorOptions) (Iterator, error) {
	backend, err := s.getReadBackend(context)
	if err != nil {
		return nil, err
	}

	prefixBz, err := s.EncodeKey(prefix)
	if err != nil {
		return nil, err
	}

	return prefixIterator(backend.IndexStoreReader(), backend, s, prefixBz, options)
}

func (s indexKeyIndex) RangeIterator(context context.Context, start, end []protoreflect.Value, options IteratorOptions) (Iterator, error) {
	backend, err := s.getReadBackend(context)
	if err != nil {
		return nil, err
	}

	err = s.CheckValidRangeIterationKeys(start, end)
	if err != nil {
		return nil, err
	}

	startBz, err := s.EncodeKey(start)
	if err != nil {
		return nil, err
	}

	endBz, err := s.EncodeKey(end)
	if err != nil {
		return nil, err
	}

	fullEndKey := len(s.GetFieldNames()) == len(end)

	return rangeIterator(backend.IndexStoreReader(), backend, s, startBz, endBz, fullEndKey, options)
}

var _ indexer = &indexKeyIndex{}
var _ Index = &indexKeyIndex{}

func (s indexKeyIndex) doNotImplement() {}

func (s indexKeyIndex) onInsert(store kvstore.Writer, message protoreflect.Message) error {
	k, v, err := s.EncodeKVFromMessage(message)
	if err != nil {
		return err
	}
	return store.Set(k, v)
}

func (s indexKeyIndex) onUpdate(store kvstore.Writer, new, existing protoreflect.Message) error {
	newValues := s.GetKeyValues(new)
	existingValues := s.GetKeyValues(existing)
	if s.CompareKeys(newValues, existingValues) == 0 {
		return nil
	}

	existingKey, err := s.EncodeKey(existingValues)
	if err != nil {
		return err
	}
	err = store.Delete(existingKey)
	if err != nil {
		return err
	}

	newKey, err := s.EncodeKey(newValues)
	if err != nil {
		return err
	}
	return store.Set(newKey, []byte{})
}

func (s indexKeyIndex) onDelete(store kvstore.Writer, message protoreflect.Message) error {
	_, key, err := s.EncodeKeyFromMessage(message)
	if err != nil {
		return err
	}
	return store.Delete(key)
}

func (s indexKeyIndex) readValueFromIndexKey(backend ReadBackend, primaryKey []protoreflect.Value, _ []byte, message proto.Message) error {
	found, err := s.primaryKey.get(backend, message, primaryKey)
	if err != nil {
		return err
	}

	if !found {
		return ormerrors.UnexpectedError.Wrapf("can't find primary key")
	}

	return nil
}
