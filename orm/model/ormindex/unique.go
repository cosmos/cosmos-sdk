package ormindex

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/cosmos/cosmos-sdk/orm/types/ormerrors"

	"github.com/cosmos/cosmos-sdk/orm/model/ormiterator"

	"github.com/cosmos/cosmos-sdk/orm/backend/kv"
	"github.com/cosmos/cosmos-sdk/orm/encoding/ormkv"
)

type UniqueIndexImpl struct {
	ormkv.UniqueKeyCodec
	primaryKey PrimaryKey
}

func (u UniqueIndexImpl) PrefixIterator(store kv.IndexCommitmentReadStore, prefix []protoreflect.Value, options IteratorOptions) (ormiterator.Iterator, error) {
	prefixBz, err := u.GetKeyCodec().Encode(prefix)
	if err != nil {
		return nil, err
	}

	return prefixIterator(store.ReadIndexStore(), store, u, prefixBz, options)
}

func (u UniqueIndexImpl) RangeIterator(store kv.IndexCommitmentReadStore, start, end []protoreflect.Value, options IteratorOptions) (ormiterator.Iterator, error) {
	keyCodec := u.GetKeyCodec()
	err := keyCodec.CheckValidRangeIterationKeys(start, end)
	if err != nil {
		return nil, err
	}

	startBz, err := keyCodec.Encode(start)
	if err != nil {
		return nil, err
	}

	endBz, err := keyCodec.Encode(end)
	if err != nil {
		return nil, err
	}

	return rangeIterator(store.ReadIndexStore(), store, u, startBz, endBz, options)
}

func (u UniqueIndexImpl) doNotImplement() {}

func (u UniqueIndexImpl) Has(store kv.IndexCommitmentReadStore, keyValues []protoreflect.Value) (found bool, err error) {
	key, err := u.GetKeyCodec().Encode(keyValues)
	if err != nil {
		return false, err
	}

	return store.ReadIndexStore().Has(key)
}

func (u UniqueIndexImpl) Get(store kv.IndexCommitmentReadStore, keyValues []protoreflect.Value, message proto.Message) (found bool, err error) {
	key, err := u.GetKeyCodec().Encode(keyValues)
	if err != nil {
		return false, err
	}

	bz, err := store.ReadIndexStore().Get(key)
	if err != nil {
		return false, err
	}

	if len(bz) == 0 {
		return false, nil
	}

	return true, proto.Unmarshal(bz, message)
}

func (u UniqueIndexImpl) OnCreate(store kv.Store, message protoreflect.Message) error {
	k, v, err := u.EncodeKVFromMessage(message)
	if err != nil {
		return err
	}

	return store.Set(k, v)
}

func (u UniqueIndexImpl) OnUpdate(store kv.Store, new, existing protoreflect.Message) error {
	keyCodec := u.GetKeyCodec()
	newValues := keyCodec.GetValues(new)
	existingValues := keyCodec.GetValues(existing)
	if keyCodec.CompareValues(newValues, existingValues) == 0 {
		return nil
	}

	existingKey, err := keyCodec.Encode(existingValues)
	if err != nil {
		return err
	}
	err = store.Delete(existingKey)
	if err != nil {
		return err
	}

	newKey, err := keyCodec.Encode(newValues)
	if err != nil {
		return err
	}

	_, value, err := u.GetValueCodec().EncodeFromMessage(new)
	if err != nil {
		return err
	}

	return store.Set(newKey, value)
}

func (u UniqueIndexImpl) OnDelete(store kv.Store, message protoreflect.Message) error {
	_, key, err := u.GetKeyCodec().EncodeFromMessage(message)
	if err != nil {
		return err
	}

	return store.Delete(key)
}

func (u UniqueIndexImpl) ReadValueFromIndexKey(store kv.IndexCommitmentReadStore, primaryKey []protoreflect.Value, _ []byte, message proto.Message) error {
	found, err := u.primaryKey.Get(store, primaryKey, message)
	if err != nil {
		return err
	}

	if !found {
		return ormerrors.UnexpectedError.Wrapf("can't find primary key")
	}

	return nil
}

var _ Indexer = &UniqueIndexImpl{}
var _ UniqueIndex = &UniqueIndexImpl{}
