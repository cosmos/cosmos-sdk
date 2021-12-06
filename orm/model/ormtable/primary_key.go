package ormtable

import (
	"github.com/cosmos/cosmos-sdk/orm/encoding/ormkv"
	"github.com/cosmos/cosmos-sdk/orm/model/kvstore"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type PrimaryKeyIndex struct {
	*ormkv.PrimaryKeyCodec
}

func NewPrimaryKeyIndex(primaryKeyCodec *ormkv.PrimaryKeyCodec) *PrimaryKeyIndex {
	return &PrimaryKeyIndex{PrimaryKeyCodec: primaryKeyCodec}
}

func (p PrimaryKeyIndex) PrefixIterator(store kvstore.IndexCommitmentReadStore, prefix []protoreflect.Value, options IteratorOptions) (Iterator, error) {
	prefixBz, err := p.EncodeKey(prefix)
	if err != nil {
		return nil, err
	}

	return prefixIterator(store.ReadCommitmentStore(), store, p, prefixBz, options)
}

func (p PrimaryKeyIndex) RangeIterator(store kvstore.IndexCommitmentReadStore, start, end []protoreflect.Value, options IteratorOptions) (Iterator, error) {
	err := p.CheckValidRangeIterationKeys(start, end)
	if err != nil {
		return nil, err
	}

	startBz, err := p.EncodeKey(start)
	if err != nil {
		return nil, err
	}

	endBz, err := p.EncodeKey(end)
	if err != nil {
		return nil, err
	}

	return rangeIterator(store.ReadCommitmentStore(), store, p, startBz, endBz, options)
}

func (p PrimaryKeyIndex) doNotImplement() {}

func (p PrimaryKeyIndex) Has(store kvstore.IndexCommitmentReadStore, key []protoreflect.Value) (found bool, err error) {
	keyBz, err := p.EncodeKey(key)
	if err != nil {
		return false, err
	}

	return store.ReadCommitmentStore().Has(keyBz)
}

func (p PrimaryKeyIndex) Get(store kvstore.IndexCommitmentReadStore, keyValues []protoreflect.Value, message proto.Message) (found bool, err error) {
	key, err := p.EncodeKey(keyValues)
	if err != nil {
		return false, err
	}

	return p.GetByKeyBytes(store, key, keyValues, message)
}

func (p PrimaryKeyIndex) GetByKeyBytes(store kvstore.IndexCommitmentReadStore, key []byte, keyValues []protoreflect.Value, message proto.Message) (found bool, err error) {
	bz, err := store.ReadCommitmentStore().Get(key)
	if err != nil {
		return false, err
	}

	if bz == nil {
		return false, nil
	}

	return true, p.Unmarshal(keyValues, bz, message)
}

func (p PrimaryKeyIndex) ReadValueFromIndexKey(_ kvstore.IndexCommitmentReadStore, primaryKey []protoreflect.Value, value []byte, message proto.Message) error {
	return p.Unmarshal(primaryKey, value, message)
}

var _ UniqueIndex = &PrimaryKeyIndex{}
