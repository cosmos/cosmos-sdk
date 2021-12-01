package ormindex

import (
	"github.com/cosmos/cosmos-sdk/orm/model/ormiterator"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/cosmos/cosmos-sdk/orm/backend/kv"
	"github.com/cosmos/cosmos-sdk/orm/encoding/ormkv"
)

type PrimaryKey struct {
	*ormkv.PrimaryKeyCodec
}

func (p PrimaryKey) PrefixIterator(store kv.IndexCommitmentReadStore, prefix []protoreflect.Value, options IteratorOptions) (ormiterator.Iterator, error) {
	prefixBz, err := p.Encode(prefix)
	if err != nil {
		return nil, err
	}

	return prefixIterator(store.ReadCommitmentStore(), store, p, prefixBz, options), nil
}

func (p PrimaryKey) RangeIterator(store kv.IndexCommitmentReadStore, start, end []protoreflect.Value, options IteratorOptions) (ormiterator.Iterator, error) {
	err := p.CheckValidRangeIterationKeys(start, end)
	if err != nil {
		return nil, err
	}

	startBz, err := p.Encode(start)
	if err != nil {
		return nil, err
	}

	endBz, err := p.Encode(end)
	if err != nil {
		return nil, err
	}

	return rangeIterator(store.ReadCommitmentStore(), store, p, startBz, endBz, options), nil
}

func (p PrimaryKey) doNotImplement() {}

func (p PrimaryKey) Has(store kv.IndexCommitmentReadStore, key []protoreflect.Value) (found bool, err error) {
	keyBz, err := p.Encode(key)
	if err != nil {
		return false, err
	}

	return store.ReadCommitmentStore().Has(keyBz)
}

func (p PrimaryKey) Get(store kv.IndexCommitmentReadStore, keyValues []protoreflect.Value, message proto.Message) (found bool, err error) {
	key, err := p.Encode(keyValues)
	if err != nil {
		return false, err
	}

	return p.GetByKeyBytes(store, key, keyValues, message)
}

func (p PrimaryKey) GetByKeyBytes(store kv.IndexCommitmentReadStore, key []byte, keyValues []protoreflect.Value, message proto.Message) (found bool, err error) {
	bz, err := store.ReadCommitmentStore().Get(key)
	if err != nil {
		return true, err
	}

	return true, p.Unmarshal(keyValues, bz, message)
}

func (p PrimaryKey) ReadValueFromIndexKey(_ kv.IndexCommitmentReadStore, primaryKey []protoreflect.Value, value []byte, message proto.Message) error {
	return p.Unmarshal(primaryKey, value, message)
}

var _ UniqueIndex = &PrimaryKey{}
