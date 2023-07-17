package ormtable

import (
	"context"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"cosmossdk.io/orm/encoding/encodeutil"
	"cosmossdk.io/orm/encoding/ormkv"
	"cosmossdk.io/orm/internal/fieldnames"
	"cosmossdk.io/orm/model/ormlist"
	"cosmossdk.io/orm/types/kv"
	"cosmossdk.io/orm/types/ormerrors"
)

type uniqueKeyIndex struct {
	*ormkv.UniqueKeyCodec
	fields         fieldnames.FieldNames
	primaryKey     *primaryKeyIndex
	getReadBackend func(context.Context) (ReadBackend, error)
}

func (u uniqueKeyIndex) List(ctx context.Context, prefixKey []interface{}, options ...ormlist.Option) (Iterator, error) {
	backend, err := u.getReadBackend(ctx)
	if err != nil {
		return nil, err
	}

	return prefixIterator(backend.IndexStoreReader(), backend, u, u.GetKeyCodec(), prefixKey, options)
}

func (u uniqueKeyIndex) ListRange(ctx context.Context, from, to []interface{}, options ...ormlist.Option) (Iterator, error) {
	backend, err := u.getReadBackend(ctx)
	if err != nil {
		return nil, err
	}

	return rangeIterator(backend.IndexStoreReader(), backend, u, u.GetKeyCodec(), from, to, options)
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

func (u uniqueKeyIndex) DeleteBy(ctx context.Context, keyValues ...interface{}) error {
	it, err := u.List(ctx, keyValues)
	if err != nil {
		return err
	}

	return u.primaryKey.deleteByIterator(ctx, it)
}

func (u uniqueKeyIndex) DeleteRange(ctx context.Context, from, to []interface{}) error {
	it, err := u.ListRange(ctx, from, to)
	if err != nil {
		return err
	}

	return u.primaryKey.deleteByIterator(ctx, it)
}

func (u uniqueKeyIndex) onInsert(store kv.Store, message protoreflect.Message) error {
	k, v, err := u.EncodeKVFromMessage(message)
	if err != nil {
		return err
	}

	has, err := store.Has(k)
	if err != nil {
		return err
	}

	if has {
		return ormerrors.UniqueKeyViolation.Wrapf("%q", u.fields)
	}

	return store.Set(k, v)
}

func (u uniqueKeyIndex) onUpdate(store kv.Store, new, existing protoreflect.Message) error {
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
		return ormerrors.UniqueKeyViolation.Wrapf("%q", u.fields)
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

func (u uniqueKeyIndex) onDelete(store kv.Store, message protoreflect.Message) error {
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

func (u uniqueKeyIndex) Fields() string {
	return u.fields.String()
}

var (
	_ indexer     = &uniqueKeyIndex{}
	_ UniqueIndex = &uniqueKeyIndex{}
)

// isNonTrivialUniqueKey checks if unique key fields are non-trivial, meaning that they
// don't contain the full primary key. If they contain the full primary key, then
// we can just use a regular index because there is no new unique constraint.
func isNonTrivialUniqueKey(fields, primaryKeyFields []protoreflect.Name) bool {
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
