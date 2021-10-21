package orm

import (
	"reflect"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
)

var _ Indexable = &table{}

// table is the high level object to storage mapper functionality. Persistent
// entities are stored by an unique identifier called `RowID`. The table struct
// does not:
// - enforce uniqueness of the `RowID`
// - enforce prefix uniqueness of keys, i.e. not allowing one key to be a prefix
// of another
// - optimize Gas usage conditions
// The caller must ensure that these things are handled. The table struct is
// private, so that we only have custom tables built on top of table, that do satisfy
// these requirements.
type table struct {
	model       reflect.Type
	prefix      [2]byte
	afterSet    []AfterSetInterceptor
	afterDelete []AfterDeleteInterceptor
	cdc         codec.Codec
}

// newTable creates a new table
func newTable(prefix [2]byte, model codec.ProtoMarshaler, cdc codec.Codec) (*table, error) {
	if model == nil {
		return nil, errors.ErrORMEmptyModel.Wrap("Model must not be nil")
	}
	tp := reflect.TypeOf(model)
	if tp.Kind() == reflect.Ptr {
		tp = tp.Elem()
	}
	return &table{
		prefix: prefix,
		model:  tp,
		cdc:    cdc,
	}, nil
}

// RowGetter returns a type safe RowGetter.
func (a table) RowGetter() RowGetter {
	return NewTypeSafeRowGetter(a.prefix, a.model, a.cdc)
}

// AddAfterSetInterceptor can be used to register a callback function that is executed after an object is created and/or updated.
func (a *table) AddAfterSetInterceptor(interceptor AfterSetInterceptor) {
	a.afterSet = append(a.afterSet, interceptor)
}

// AddAfterDeleteInterceptor can be used to register a callback function that is executed after an object is deleted.
func (a *table) AddAfterDeleteInterceptor(interceptor AfterDeleteInterceptor) {
	a.afterDelete = append(a.afterDelete, interceptor)
}

// Create persists the given object under the rowID key, returning an
// errors.ErrORMUniqueConstraint if a value already exists at that key.
//
// Create iterates through the registered callbacks that may add secondary index
// keys.
func (a table) Create(store sdk.KVStore, rowID RowID, obj codec.ProtoMarshaler) error {
	if a.Has(store, rowID) {
		return errors.ErrORMUniqueConstraint
	}

	return a.Set(store, rowID, obj)
}

// Update updates the given object under the rowID key. It expects the key to
// exists already and fails with an `errors.ErrNotFound` otherwise. Any caller must
// therefore make sure that this contract is fulfilled. Parameters must not be
// nil.
//
// Update triggers all "after set" hooks that may add or remove secondary index keys.
func (a table) Update(store sdk.KVStore, rowID RowID, newValue codec.ProtoMarshaler) error {
	if !a.Has(store, rowID) {
		return errors.ErrNotFound
	}

	return a.Set(store, rowID, newValue)
}

// Set persists the given object under the rowID key. It does not check if the
// key already exists and overwrites the value if it does.
//
// Set iterates through the registered callbacks that may add secondary index
// keys.
func (a table) Set(store sdk.KVStore, rowID RowID, newValue codec.ProtoMarshaler) error {
	if len(rowID) == 0 {
		return errors.ErrORMEmptyKey
	}
	if err := assertCorrectType(a.model, newValue); err != nil {
		return err
	}
	if err := assertValid(newValue); err != nil {
		return err
	}

	pStore := prefix.NewStore(store, a.prefix[:])

	var oldValue codec.ProtoMarshaler
	if a.Has(store, rowID) {
		oldValue = reflect.New(a.model).Interface().(codec.ProtoMarshaler)
		a.GetOne(store, rowID, oldValue)
	}

	newValueEncoded, err := a.cdc.Marshal(newValue)
	if err != nil {
		return errors.Wrapf(err, "failed to serialize %T", newValue)
	}

	pStore.Set(rowID, newValueEncoded)
	for i, itc := range a.afterSet {
		if err := itc(store, rowID, newValue, oldValue); err != nil {
			return errors.Wrapf(err, "interceptor %d failed", i)
		}
	}
	return nil
}

func assertValid(obj codec.ProtoMarshaler) error {
	if v, ok := obj.(Validateable); ok {
		if err := v.ValidateBasic(); err != nil {
			return err
		}
	}
	return nil
}

// Delete removes the object under the rowID key. It expects the key to exists
// already and fails with a `errors.ErrNotFound` otherwise. Any caller must therefore
// make sure that this contract is fulfilled.
//
// Delete iterates through the registered callbacks that remove secondary index
// keys.
func (a table) Delete(store sdk.KVStore, rowID RowID) error {
	pStore := prefix.NewStore(store, a.prefix[:])

	var oldValue = reflect.New(a.model).Interface().(codec.ProtoMarshaler)
	if err := a.GetOne(store, rowID, oldValue); err != nil {
		return errors.Wrap(err, "load old value")
	}
	pStore.Delete(rowID)

	for i, itc := range a.afterDelete {
		if err := itc(store, rowID, oldValue); err != nil {
			return errors.Wrapf(err, "delete interceptor %d failed", i)
		}
	}
	return nil
}

// Has checks if a key exists. Returns false when the key is empty or nil
// because we don't allow creation of values without a key.
func (a table) Has(store sdk.KVStore, key RowID) bool {
	if len(key) == 0 {
		return false
	}
	pStore := prefix.NewStore(store, a.prefix[:])
	it := pStore.Iterator(PrefixRange(key))
	defer it.Close()
	return it.Valid()
}

// GetOne load the object persisted for the given RowID into the dest parameter.
// If none exists or `rowID==nil` then `errors.ErrNotFound` is returned instead.
// Parameters must not be nil - we don't allow creation of values with empty keys.
func (a table) GetOne(store sdk.KVStore, rowID RowID, dest codec.ProtoMarshaler) error {
	if len(rowID) == 0 {
		return errors.ErrNotFound
	}
	x := NewTypeSafeRowGetter(a.prefix, a.model, a.cdc)
	return x(store, rowID, dest)
}
