package ormtable

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/cosmos/cosmos-sdk/orm/encoding/ormkv"
	"github.com/cosmos/cosmos-sdk/orm/model/kvstore"
)

// PrimaryKeyIndex defines an UniqueIndex for the primary key.
type PrimaryKeyIndex struct {
	*ormkv.PrimaryKeyCodec
}

// NewPrimaryKeyIndex returns a new PrimaryKeyIndex.
func NewPrimaryKeyIndex(primaryKeyCodec *ormkv.PrimaryKeyCodec) *PrimaryKeyIndex {
	return &PrimaryKeyIndex{PrimaryKeyCodec: primaryKeyCodec}
}

func (p PrimaryKeyIndex) PrefixIterator(store kvstore.IndexCommitmentReadStore, prefix []protoreflect.Value, options IteratorOptions) (Iterator, error) {
	prefixBz, err := p.EncodeKey(prefix)
	if err != nil {
		return nil, err
	}

	return prefixIterator(store.CommitmentStoreReader(), store, p, prefixBz, options)
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

	fullEndKey := len(p.GetFieldNames()) == len(end)

	return rangeIterator(store.CommitmentStoreReader(), store, p, startBz, endBz, fullEndKey, options)
}

func (p PrimaryKeyIndex) doNotImplement() {}

func (p PrimaryKeyIndex) Has(store kvstore.IndexCommitmentReadStore, key []protoreflect.Value) (found bool, err error) {
	keyBz, err := p.EncodeKey(key)
	if err != nil {
		return false, err
	}

	return store.CommitmentStoreReader().Has(keyBz)
}

func (p PrimaryKeyIndex) Get(store kvstore.IndexCommitmentReadStore, keyValues []protoreflect.Value, message proto.Message) (found bool, err error) {
	key, err := p.EncodeKey(keyValues)
	if err != nil {
		return false, err
	}

	return p.GetByKeyBytes(store, key, keyValues, message)
}

func (p PrimaryKeyIndex) GetByKeyBytes(store kvstore.IndexCommitmentReadStore, key []byte, keyValues []protoreflect.Value, message proto.Message) (found bool, err error) {
	bz, err := store.CommitmentStoreReader().Get(key)
	if err != nil {
		return false, err
	}

	if bz == nil {
		return false, nil
	}

	return true, p.Unmarshal(keyValues, bz, message)
}

func (p PrimaryKeyIndex) readValueFromIndexKey(_ kvstore.IndexCommitmentReadStore, primaryKey []protoreflect.Value, value []byte, message proto.Message) error {
	return p.Unmarshal(primaryKey, value, message)
}

var _ UniqueIndex = &PrimaryKeyIndex{}
