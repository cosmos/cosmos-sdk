package ormtable

import (
	"github.com/cosmos/cosmos-sdk/orm/model/kvstore"
	"github.com/cosmos/cosmos-sdk/orm/types/ormerrors"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/cosmos/cosmos-sdk/orm/encoding/ormkv"
)

type IndexKeyIndex struct {
	*ormkv.IndexKeyCodec
	primaryKey *PrimaryKeyIndex
}

func NewIndexKeyIndex(indexKeyCodec *ormkv.IndexKeyCodec, primaryKey *PrimaryKeyIndex) *IndexKeyIndex {
	return &IndexKeyIndex{IndexKeyCodec: indexKeyCodec, primaryKey: primaryKey}
}

func (s IndexKeyIndex) PrefixIterator(store kvstore.IndexCommitmentReadStore, prefix []protoreflect.Value, options IteratorOptions) (Iterator, error) {
	prefixBz, err := s.EncodeKey(prefix)
	if err != nil {
		return nil, err
	}

	return prefixIterator(store.ReadIndexStore(), store, s, prefixBz, options)
}

func (s IndexKeyIndex) RangeIterator(store kvstore.IndexCommitmentReadStore, start, end []protoreflect.Value, options IteratorOptions) (Iterator, error) {
	err := s.CheckValidRangeIterationKeys(start, end)
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

	return rangeIterator(store.ReadIndexStore(), store, s, startBz, endBz, options)
}

var _ Indexer = &IndexKeyIndex{}
var _ Index = &IndexKeyIndex{}

var sentinelValue = []byte{0}

func (s IndexKeyIndex) doNotImplement() {}

func (s IndexKeyIndex) OnInsert(store kvstore.Store, message protoreflect.Message) error {
	k, v, err := s.EncodeKVFromMessage(message)
	if err != nil {
		return err
	}
	return store.Set(k, v)
}

func (s IndexKeyIndex) OnUpdate(store kvstore.Store, new, existing protoreflect.Message) error {
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
	return store.Set(newKey, sentinelValue)
}

func (s IndexKeyIndex) OnDelete(store kvstore.Store, message protoreflect.Message) error {
	_, key, err := s.EncodeKeyFromMessage(message)
	if err != nil {
		return err
	}
	return store.Delete(key)
}

func (s IndexKeyIndex) ReadValueFromIndexKey(store kvstore.IndexCommitmentReadStore, primaryKey []protoreflect.Value, _ []byte, message proto.Message) error {
	found, err := s.primaryKey.Get(store, primaryKey, message)
	if err != nil {
		return err
	}

	if !found {
		return ormerrors.UnexpectedError.Wrapf("can't find primary key")
	}

	return nil
}
