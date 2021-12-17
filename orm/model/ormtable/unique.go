package ormtable

import (
	"context"

	"github.com/cosmos/cosmos-sdk/orm/encoding/encodeutil"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/cosmos/cosmos-sdk/orm/encoding/ormkv"
	"github.com/cosmos/cosmos-sdk/orm/model/kvstore"
	"github.com/cosmos/cosmos-sdk/orm/types/ormerrors"
)

type uniqueKeyIndex struct {
	*ormkv.UniqueKeyCodec
	primaryKey     *primaryKeyIndex
	getReadBackend func(context.Context) (ReadBackend, error)
}

func (u uniqueKeyIndex) PrefixIterator(ctx context.Context, prefix []protoreflect.Value, options IteratorOptions) (Iterator, error) {
	backend, err := u.getReadBackend(ctx)
	if err != nil {
		return nil, err
	}

	prefixBz, err := u.GetKeyCodec().EncodeKey(prefix)
	if err != nil {
		return nil, err
	}

	return prefixIterator(backend.IndexStoreReader(), backend, u, prefixBz, options)
}

func (u uniqueKeyIndex) RangeIterator(ctx context.Context, start, end []protoreflect.Value, options IteratorOptions) (Iterator, error) {
	backend, err := u.getReadBackend(ctx)
	if err != nil {
		return nil, err
	}

	keyCodec := u.GetKeyCodec()
	err = keyCodec.CheckValidRangeIterationKeys(start, end)
	if err != nil {
		return nil, err
	}

	startBz, err := keyCodec.EncodeKey(start)
	if err != nil {
		return nil, err
	}

	endBz, err := keyCodec.EncodeKey(end)
	if err != nil {
		return nil, err
	}

	fullEndKey := len(keyCodec.GetFieldNames()) == len(end)

	return rangeIterator(backend.IndexStoreReader(), backend, u, startBz, endBz, fullEndKey, options)
}

func (u uniqueKeyIndex) doNotImplement() {}

func (u uniqueKeyIndex) Has(ctx context.Context, values ...interface{}) (found bool, err error) {
	backend, err := u.getReadBackend(ctx)
	if err != nil {
		return false, err
	}

	key, err := u.GetKeyCodec().EncodeKey(encodeutil.ValuesOf(values...))
	if err != nil {
		return false, err
	}

	return backend.IndexStoreReader().Has(key)
}

func (u uniqueKeyIndex) Get(ctx context.Context, message proto.Message, keyValues ...interface{}) (found bool, err error) {
	backend, err := u.getReadBackend(ctx)
	if err != nil {
		return false, err
	}

	key, err := u.GetKeyCodec().EncodeKey(encodeutil.ValuesOf(keyValues...))
	if err != nil {
		return false, err
	}

	value, err := backend.IndexStoreReader().Get(key)
	if err != nil {
		return false, err
	}

	// for unique keys, value can be empty and the entry still exists
	if value == nil {
		return false, nil
	}

	_, pk, err := u.DecodeIndexKey(key, value)
	if err != nil {
		return true, err
	}

	return u.primaryKey.get(backend, message, pk)
}

func (u uniqueKeyIndex) DeleteByKey(ctx context.Context, keyValues ...interface{}) error {
	backend, err := u.getReadBackend(ctx)
	if err != nil {
		return err
	}

	key, err := u.GetKeyCodec().EncodeKey(encodeutil.ValuesOf(keyValues...))
	if err != nil {
		return err
	}

	value, err := backend.IndexStoreReader().Get(key)
	if err != nil {
		return err
	}

	// for unique keys, value can be empty and the entry still exists
	if value == nil {
		return nil
	}

	_, pk, err := u.DecodeIndexKey(key, value)
	if err != nil {
		return err
	}

	return u.primaryKey.doDeleteByKey(ctx, pk)
}

func (u uniqueKeyIndex) onInsert(store kvstore.Writer, message protoreflect.Message) error {
	k, v, err := u.EncodeKVFromMessage(message)
	if err != nil {
		return err
	}

	has, err := store.Has(k)
	if err != nil {
		return err
	}

	if has {
		return ormerrors.UniqueKeyViolation
	}

	return store.Set(k, v)
}

func (u uniqueKeyIndex) onUpdate(store kvstore.Writer, new, existing protoreflect.Message) error {
	keyCodec := u.GetKeyCodec()
	newValues := keyCodec.GetKeyValues(new)
	existingValues := keyCodec.GetKeyValues(existing)
	if keyCodec.CompareKeys(newValues, existingValues) == 0 {
		return nil
	}

	newKey, err := keyCodec.EncodeKey(newValues)
	if err != nil {
		return err
	}

	has, err := store.Has(newKey)
	if err != nil {
		return err
	}

	if has {
		return ormerrors.UniqueKeyViolation
	}

	existingKey, err := keyCodec.EncodeKey(existingValues)
	if err != nil {
		return err
	}

	err = store.Delete(existingKey)
	if err != nil {
		return err
	}

	_, value, err := u.GetValueCodec().EncodeKeyFromMessage(new)
	if err != nil {
		return err
	}

	return store.Set(newKey, value)
}

func (u uniqueKeyIndex) onDelete(store kvstore.Writer, message protoreflect.Message) error {
	_, key, err := u.GetKeyCodec().EncodeKeyFromMessage(message)
	if err != nil {
		return err
	}

	return store.Delete(key)
}

func (u uniqueKeyIndex) readValueFromIndexKey(store ReadBackend, primaryKey []protoreflect.Value, _ []byte, message proto.Message) error {
	found, err := u.primaryKey.get(store, message, primaryKey)
	if err != nil {
		return err
	}

	if !found {
		return ormerrors.UnexpectedError.Wrapf("can't find primary key")
	}

	return nil
}

var _ indexer = &uniqueKeyIndex{}
var _ UniqueIndex = &uniqueKeyIndex{}

// isNonTrivialUniqueKey checks if unique key fields are non-trivial, meaning that they
// don't contain the full primary key. If they contain the full primary key, then
// we can just use a regular index because there is no new unique constraint.
func isNonTrivialUniqueKey(fields []protoreflect.Name, primaryKeyFields []protoreflect.Name) bool {
	have := map[protoreflect.Name]bool{}
	for _, field := range fields {
		have[field] = true
	}

	for _, field := range primaryKeyFields {
		if !have[field] {
			return true
		}
	}

	return false
}
