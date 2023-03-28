package ormtable

import (
	"context"

	"github.com/cosmos/cosmos-sdk/orm/types/kv"

	"github.com/cosmos/cosmos-sdk/orm/internal/fieldnames"

	"github.com/cosmos/cosmos-sdk/orm/model/ormlist"

	"github.com/cosmos/cosmos-sdk/orm/types/ormerrors"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/cosmos/cosmos-sdk/orm/encoding/ormkv"
)

// indexKeyIndex implements Index for a regular IndexKey.
type indexKeyIndex struct {
	*ormkv.IndexKeyCodec
	fields         fieldnames.FieldNames
	primaryKey     *primaryKeyIndex
	getReadBackend func(context.Context) (ReadBackend, error)
}

func (i indexKeyIndex) DeleteBy(ctx context.Context, keyValues ...interface{}) error {
	it, err := i.List(ctx, keyValues)
	if err != nil {
		return err
	}

	return i.primaryKey.deleteByIterator(ctx, it)
}

func (i indexKeyIndex) DeleteRange(ctx context.Context, from, to []interface{}) error {
	it, err := i.ListRange(ctx, from, to)
	if err != nil {
		return err
	}

	return i.primaryKey.deleteByIterator(ctx, it)
}

func (i indexKeyIndex) List(ctx context.Context, prefixKey []interface{}, options ...ormlist.Option) (Iterator, error) {
	backend, err := i.getReadBackend(ctx)
	if err != nil {
		return nil, err
	}

	return prefixIterator(backend.IndexStoreReader(), backend, i, i.KeyCodec, prefixKey, options)
}

func (i indexKeyIndex) ListRange(ctx context.Context, from, to []interface{}, options ...ormlist.Option) (Iterator, error) {
	backend, err := i.getReadBackend(ctx)
	if err != nil {
		return nil, err
	}

	return rangeIterator(backend.IndexStoreReader(), backend, i, i.KeyCodec, from, to, options)
}

var (
	_ indexer = &indexKeyIndex{}
	_ Index   = &indexKeyIndex{}
)

func (i indexKeyIndex) doNotImplement() {}

func (i indexKeyIndex) onInsert(store kv.Store, message protoreflect.Message) error {
	k, v, err := i.EncodeKVFromMessage(message)
	if err != nil {
		return err
	}
	return store.Set(k, v)
}

func (i indexKeyIndex) onUpdate(store kv.Store, new, existing protoreflect.Message) error {
	newValues := i.GetKeyValues(new)
	existingValues := i.GetKeyValues(existing)
	if i.CompareKeys(newValues, existingValues) == 0 {
		return nil
	}

	existingKey, err := i.EncodeKey(existingValues)
	if err != nil {
		return err
	}
	err = store.Delete(existingKey)
	if err != nil {
		return err
	}

	newKey, err := i.EncodeKey(newValues)
	if err != nil {
		return err
	}
	return store.Set(newKey, []byte{})
}

func (i indexKeyIndex) onDelete(store kv.Store, message protoreflect.Message) error {
	_, key, err := i.EncodeKeyFromMessage(message)
	if err != nil {
		return err
	}
	return store.Delete(key)
}

func (i indexKeyIndex) readValueFromIndexKey(backend ReadBackend, primaryKey []protoreflect.Value, _ []byte, message proto.Message) error {
	found, err := i.primaryKey.get(backend, message, primaryKey)
	if err != nil {
		return err
	}

	if !found {
		return ormerrors.UnexpectedError.Wrapf("can't find primary key")
	}

	return nil
}

func (i indexKeyIndex) Fields() string {
	return i.fields.String()
}
