package table

import (
	"reflect"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
)

var _ Indexable = &TableBuilder{}

type TableBuilder struct {
	model         reflect.Type
	prefixData    byte
	indexKeyCodec IndexKeyCodec
	afterSave     []AfterSaveInterceptor
	afterDelete   []AfterDeleteInterceptor
	cdc           codec.Codec
}

// NewTableBuilder creates a builder to setup a Table object.
func NewTableBuilder(prefixData byte, model codec.ProtoMarshaler, idxKeyCodec IndexKeyCodec, cdc codec.Codec) *TableBuilder {
	if model == nil {
		panic("Model must not be nil")
	}
	if idxKeyCodec == nil {
		panic("IndexKeyCodec must not be nil")
	}
	tp := reflect.TypeOf(model)
	if tp.Kind() == reflect.Ptr {
		tp = tp.Elem()
	}
	return &TableBuilder{
		prefixData:    prefixData,
		model:         tp,
		indexKeyCodec: idxKeyCodec,
		cdc:           cdc,
	}
}

func (a TableBuilder) IndexKeyCodec() IndexKeyCodec {
	return a.indexKeyCodec
}

// RowGetter returns a type safe RowGetter.
func (a TableBuilder) RowGetter() RowGetter {
	return NewTypeSafeRowGetter(a.prefixData, a.model, a.cdc)
}

// Build creates a new Table object.
func (a TableBuilder) Build() Table {
	return Table{
		model:       a.model,
		prefix:      a.prefixData,
		afterSave:   a.afterSave,
		afterDelete: a.afterDelete,
		cdc:         a.cdc,
	}
}

// AddAfterSaveInterceptor can be used to register a callback function that is executed after an object is created and/or updated.
func (a *TableBuilder) AddAfterSaveInterceptor(interceptor AfterSaveInterceptor) {
	a.afterSave = append(a.afterSave, interceptor)
}

// AddAfterDeleteInterceptor can be used to register a callback function that is executed after an object is deleted.
func (a *TableBuilder) AddAfterDeleteInterceptor(interceptor AfterDeleteInterceptor) {
	a.afterDelete = append(a.afterDelete, interceptor)
}

// Table is the high level object to storage mapper functionality. Persistent entities are stored by an unique identifier
// called `RowID`.
// The Table struct does not enforce uniqueness of the `RowID` but expects this to be satisfied by the callers and conditions
// to optimize Gas usage.
type Table struct {
	model       reflect.Type
	prefix      byte
	afterSave   []AfterSaveInterceptor
	afterDelete []AfterDeleteInterceptor
	cdc         codec.Codec
}

// Create persists the given object under the rowID key. It does not check if the
// key already exists. Any caller must either make sure that this contract is fulfilled
// by providing a universal unique ID or sequence that is guaranteed to not exist yet or
// by checking the state via `Has` function before.
//
// Create iterates though the registered callbacks and may add secondary index keys by them.
func (a Table) Create(store sdk.KVStore, rowID RowID, obj codec.ProtoMarshaler) error {
	if err := assertCorrectType(a.model, obj); err != nil {
		return err
	}
	if err := assertValid(obj); err != nil {
		return err
	}
	pStore := prefix.NewStore(store, []byte{a.prefix})
	v, err := a.cdc.Marshal(obj)
	if err != nil {
		return errors.Wrapf(err, "failed to serialize %T", obj)
	}
	pStore.Set(rowID, v)
	for i, itc := range a.afterSave {
		if err := itc(store, rowID, obj, nil); err != nil {
			return errors.Wrapf(err, "interceptor %d failed", i)
		}
	}
	return nil
}

// Save updates the given object under the rowID key. It expects the key to exist already
// and fails with an `ErrNotFound` otherwise. Any caller must therefore make sure that this contract
// is fulfilled. Parameters must not be nil.
//
// Save iterates though the registered callbacks and may add or remove secondary index keys by them.
func (a Table) Save(store sdk.KVStore, rowID RowID, newValue codec.ProtoMarshaler) error {
	if err := assertCorrectType(a.model, newValue); err != nil {
		return err
	}
	if err := assertValid(newValue); err != nil {
		return err
	}

	pStore := prefix.NewStore(store, []byte{a.prefix})
	var oldValue = reflect.New(a.model).Interface().(codec.ProtoMarshaler)

	if err := a.GetOne(store, rowID, oldValue); err != nil {
		return errors.Wrap(err, "load old value")
	}
	newValueEncoded, err := a.cdc.Marshal(newValue)
	if err != nil {
		return errors.Wrapf(err, "failed to serialize %T", newValue)
	}

	pStore.Set(rowID, newValueEncoded)
	for i, itc := range a.afterSave {
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

// Delete removes the object under the rowID key. It expects the key to exists already
// and fails with a `ErrNotFound` otherwise. Any caller must therefore make sure that this contract
// is fulfilled.
//
// Delete iterates though the registered callbacks and removes secondary index keys by them.
func (a Table) Delete(store sdk.KVStore, rowID RowID) error {
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

// Has checks if a key exists. Panics on nil key.
func (a Table) Has(store sdk.KVStore, rowID RowID) bool {
	Store := prefix.NewStore(store, []byte{a.prefix})
	it := Store.Iterator(PrefixRange(rowID))
	defer it.Close()
	return it.Valid()
}

// GetOne load the object persisted for the given RowID into the dest parameter.
// If none exists `ErrNotFound` is returned instead. Parameters must not be nil.
func (a Table) GetOne(store sdk.KVStore, rowID RowID, dest codec.ProtoMarshaler) error {
	x := NewTypeSafeRowGetter(a.prefix, a.model, a.cdc)
	return x(store, rowID, dest)
}
