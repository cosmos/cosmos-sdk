package ormtable

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/cosmos/cosmos-sdk/orm/encoding/ormkv"
	"github.com/cosmos/cosmos-sdk/orm/model/kvstore"
	"github.com/cosmos/cosmos-sdk/orm/types/ormerrors"
)

type UniqueKeyIndex struct {
	*ormkv.UniqueKeyCodec
	primaryKey *PrimaryKeyIndex
}

func NewUniqueKeyIndex(uniqueKeyCodec *ormkv.UniqueKeyCodec, primaryKey *PrimaryKeyIndex) *UniqueKeyIndex {
	return &UniqueKeyIndex{UniqueKeyCodec: uniqueKeyCodec, primaryKey: primaryKey}
}

func (u UniqueKeyIndex) PrefixIterator(store kvstore.IndexCommitmentReadStore, prefix []protoreflect.Value, options IteratorOptions) (Iterator, error) {
	prefixBz, err := u.GetKeyCodec().EncodeKey(prefix)
	if err != nil {
		return nil, err
	}

	return prefixIterator(store.IndexStoreReader(), store, u, prefixBz, options)
}

func (u UniqueKeyIndex) RangeIterator(store kvstore.IndexCommitmentReadStore, start, end []protoreflect.Value, options IteratorOptions) (Iterator, error) {
	keyCodec := u.GetKeyCodec()
	err := keyCodec.CheckValidRangeIterationKeys(start, end)
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

	return rangeIterator(store.IndexStoreReader(), store, u, startBz, endBz, fullEndKey, options)
}

func (u UniqueKeyIndex) doNotImplement() {}

func (u UniqueKeyIndex) Has(store kvstore.IndexCommitmentReadStore, keyValues []protoreflect.Value) (found bool, err error) {
	key, err := u.GetKeyCodec().EncodeKey(keyValues)
	if err != nil {
		return false, err
	}

	return store.IndexStoreReader().Has(key)
}

func (u UniqueKeyIndex) Get(store kvstore.IndexCommitmentReadStore, keyValues []protoreflect.Value, message proto.Message) (found bool, err error) {
	key, err := u.GetKeyCodec().EncodeKey(keyValues)
	if err != nil {
		return false, err
	}

	value, err := store.IndexStoreReader().Get(key)
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

	return u.primaryKey.Get(store, pk, message)
}

func (u UniqueKeyIndex) onInsert(store kvstore.Writer, message protoreflect.Message) error {
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

func (u UniqueKeyIndex) onUpdate(store kvstore.Writer, new, existing protoreflect.Message) error {
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

func (u UniqueKeyIndex) onDelete(store kvstore.Writer, message protoreflect.Message) error {
	_, key, err := u.GetKeyCodec().EncodeKeyFromMessage(message)
	if err != nil {
		return err
	}

	return store.Delete(key)
}

func (u UniqueKeyIndex) readValueFromIndexKey(store kvstore.IndexCommitmentReadStore, primaryKey []protoreflect.Value, _ []byte, message proto.Message) error {
	found, err := u.primaryKey.Get(store, primaryKey, message)
	if err != nil {
		return err
	}

	if !found {
		return ormerrors.UnexpectedError.Wrapf("can't find primary key")
	}

	return nil
}

var _ indexer = &UniqueKeyIndex{}
var _ UniqueIndex = &UniqueKeyIndex{}

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
