package orm

import (
	"reflect"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
)

var _ Indexable = &tableBuilder{}

type tableBuilder struct {
	model       reflect.Type
	prefixData  byte
	afterSave   []AfterSaveInterceptor
	afterDelete []AfterDeleteInterceptor
	cdc         codec.Codec
}

// newTableBuilder creates a builder to setup a table object.
func newTableBuilder(prefixData byte, model codec.ProtoMarshaler, cdc codec.Codec) (*tableBuilder, error) {
	if model == nil {
		return nil, ErrArgument.Wrap("Model must not be nil")
	}
	tp := reflect.TypeOf(model)
	if tp.Kind() == reflect.Ptr {
		tp = tp.Elem()
	}
	return &tableBuilder{
		prefixData: prefixData,
		model:      tp,
		cdc:        cdc,
	}
}

// RowGetter returns a type safe RowGetter.
func (a tableBuilder) RowGetter() RowGetter {
	return NewTypeSafeRowGetter(a.prefixData, a.model, a.cdc)
}

// Build creates a new table object.
func (a tableBuilder) Build() table {
	return table{
		model:       a.model,
		prefix:      a.prefixData,
		afterSave:   a.afterSave,
		afterDelete: a.afterDelete,
		cdc:         a.cdc,
	}
}

// AddAfterSaveInterceptor can be used to register a callback function that is executed after an object is created and/or updated.
func (a *tableBuilder) AddAfterSaveInterceptor(interceptor AfterSaveInterceptor) {
	a.afterSave = append(a.afterSave, interceptor)
}

// AddAfterDeleteInterceptor can be used to register a callback function that is executed after an object is deleted.
func (a *tableBuilder) AddAfterDeleteInterceptor(interceptor AfterDeleteInterceptor) {
	a.afterDelete = append(a.afterDelete, interceptor)
}

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
	prefix      byte
	afterSave   []AfterSaveInterceptor
	afterDelete []AfterDeleteInterceptor
	cdc         codec.Codec
}

// Create persists the given object under the rowID key, returning an
// ErrUniqueConstraint if a value already exists at that key.
//
// Create iterates through the registered callbacks that may add secondary index
// keys.
func (a table) Create(store sdk.KVStore, rowID RowID, obj codec.ProtoMarshaler) error {
	if a.Has(store, rowID) {
		return ErrUniqueConstraint
	}

	return a.Set(store, rowID, obj)
}

// Update updates the given object under the rowID key. It expects the key to
// exists already and fails with an `ErrNotFound` otherwise. Any caller must
// therefore make sure that this contract is fulfilled. Parameters must not be
// nil.
//
// Update iterates through the registered callbacks that may add or remove
// secondary index keys.
func (a table) Update(store sdk.KVStore, rowID RowID, newValue codec.ProtoMarshaler) error {
	if !a.Has(store, rowID) {
		return ErrNotFound
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
		return ErrEmptyKey
	}
	if err := assertCorrectType(a.model, newValue); err != nil {
		return err
	}
	if err := assertValid(newValue); err != nil {
		return err
	}

	pStore := prefix.NewStore(store, []byte{a.prefix})

	var oldValue codec.ProtoMarshaler
	if a.Has(ctx, rowID) {
		oldValue = reflect.New(a.model).Interface().(codec.ProtoMarshaler)
		a.GetOne(ctx, rowID, oldValue)
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
// already and fails with a `ErrNotFound` otherwise. Any caller must therefore
// make sure that this contract is fulfilled.
//
// Delete iterates through the registered callbacks that remove secondary index
// keys.
func (a table) Delete(store sdk.KVStore, rowID RowID) error {
	pStore := prefix.NewStore(store, []byte{a.prefix})

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
	pStore := prefix.NewStore(store, []byte{a.prefix})
	it := pStore.Iterator(PrefixRange(key))
	defer it.Close()
	return it.Valid()
}

// GetOne load the object persisted for the given RowID into the dest parameter.
// If none exists or `rowID==nil` then `ErrNotFound` is returned instead.
// Parameters must not be nil - we don't allow creation of values with empty keys.
func (a table) GetOne(store sdk.KVStore, rowID RowID, dest codec.ProtoMarshaler) error {
	if len(rowID) == 0 {
		return ErrNotFound
	}
	x := NewTypeSafeRowGetter(a.prefix, a.model, a.cdc)
	return x(store, rowID, dest)
}
