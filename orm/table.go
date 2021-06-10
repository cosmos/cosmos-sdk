package orm

import (
	"bytes"
	"reflect"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	"github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
)

var _ Indexable = &TableBuilder{}

type TableBuilder struct {
	model         reflect.Type
	prefixData    byte
	storeKey      sdk.StoreKey
	indexKeyCodec IndexKeyCodec
	afterSave     []AfterSaveInterceptor
	afterDelete   []AfterDeleteInterceptor
	cdc           codec.Codec
}

// NewTableBuilder creates a builder to setup a Table object.
func NewTableBuilder(prefixData byte, storeKey sdk.StoreKey, model codec.ProtoMarshaler, idxKeyCodec IndexKeyCodec, cdc codec.Codec) *TableBuilder {
	if model == nil {
		panic("Model must not be nil")
	}
	if storeKey == nil {
		panic("StoreKey must not be nil")
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
		storeKey:      storeKey,
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
	return NewTypeSafeRowGetter(a.storeKey, a.prefixData, a.model, a.cdc)
}

func (a TableBuilder) StoreKey() sdk.StoreKey {
	return a.storeKey
}

// Build creates a new Table object.
func (a TableBuilder) Build() Table {
	return Table{
		model:       a.model,
		prefix:      a.prefixData,
		storeKey:    a.storeKey,
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

var _ TableExportable = &Table{}

// Table is the high level object to storage mapper functionality. Persistent entities are stored by an unique identifier
// called `RowID`.
// The Table struct does not enforce uniqueness of the `RowID` but expects this to be satisfied by the callers and conditions
// to optimize Gas usage.
type Table struct {
	model       reflect.Type
	prefix      byte
	storeKey    sdk.StoreKey
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
func (a Table) Create(ctx HasKVStore, rowID RowID, obj codec.ProtoMarshaler) error {
	if err := assertCorrectType(a.model, obj); err != nil {
		return err
	}
	if err := assertValid(obj); err != nil {
		return err
	}
	store := prefix.NewStore(ctx.KVStore(a.storeKey), []byte{a.prefix})
	v, err := a.cdc.Marshal(obj)
	if err != nil {
		return errors.Wrapf(err, "failed to serialize %T", obj)
	}
	store.Set(rowID, v)
	for i, itc := range a.afterSave {
		if err := itc(ctx, rowID, obj, nil); err != nil {
			return errors.Wrapf(err, "interceptor %d failed", i)
		}
	}
	return nil
}

// Save updates the given object under the rowID key. It expects the key to exists already
// and fails with an `ErrNotFound` otherwise. Any caller must therefore make sure that this contract
// is fulfilled. Parameters must not be nil.
//
// Save iterates though the registered callbacks and may add or remove secondary index keys by them.
func (a Table) Save(ctx HasKVStore, rowID RowID, newValue codec.ProtoMarshaler) error {
	if err := assertCorrectType(a.model, newValue); err != nil {
		return err
	}
	if err := assertValid(newValue); err != nil {
		return err
	}

	store := prefix.NewStore(ctx.KVStore(a.storeKey), []byte{a.prefix})
	var oldValue = reflect.New(a.model).Interface().(codec.ProtoMarshaler)

	if err := a.GetOne(ctx, rowID, oldValue); err != nil {
		return errors.Wrap(err, "load old value")
	}
	newValueEncoded, err := a.cdc.Marshal(newValue)
	if err != nil {
		return errors.Wrapf(err, "failed to serialize %T", newValue)
	}

	store.Set(rowID, newValueEncoded)
	for i, itc := range a.afterSave {
		if err := itc(ctx, rowID, newValue, oldValue); err != nil {
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
func (a Table) Delete(ctx HasKVStore, rowID RowID) error {
	store := prefix.NewStore(ctx.KVStore(a.storeKey), []byte{a.prefix})

	var oldValue = reflect.New(a.model).Interface().(codec.ProtoMarshaler)
	if err := a.GetOne(ctx, rowID, oldValue); err != nil {
		return errors.Wrap(err, "load old value")
	}
	store.Delete(rowID)

	for i, itc := range a.afterDelete {
		if err := itc(ctx, rowID, oldValue); err != nil {
			return errors.Wrapf(err, "delete interceptor %d failed", i)
		}
	}
	return nil
}

// Has checks if a key exists. Panics on nil key.
func (a Table) Has(ctx HasKVStore, rowID RowID) bool {
	store := prefix.NewStore(ctx.KVStore(a.storeKey), []byte{a.prefix})
	it := store.Iterator(PrefixRange(rowID))
	defer it.Close()
	return it.Valid()
}

// GetOne load the object persisted for the given RowID into the dest parameter.
// If none exists `ErrNotFound` is returned instead. Parameters must not be nil.
func (a Table) GetOne(ctx HasKVStore, rowID RowID, dest codec.ProtoMarshaler) error {
	x := NewTypeSafeRowGetter(a.storeKey, a.prefix, a.model, a.cdc)
	return x(ctx, rowID, dest)
}

// PrefixScan returns an Iterator over a domain of keys in ascending order. End is exclusive.
// Start is an MultiKeyIndex key or prefix. It must be less than end, or the Iterator is invalid.
// Iterator must be closed by caller.
// To iterate over entire domain, use PrefixScan(nil, nil)
//
// WARNING: The use of a PrefixScan can be very expensive in terms of Gas. Please make sure you do not expose
// this as an endpoint to the public without further limits.
// Example:
//			it, err := idx.PrefixScan(ctx, start, end)
//			if err !=nil {
//				return err
//			}
//			const defaultLimit = 20
//			it = LimitIterator(it, defaultLimit)
//
// CONTRACT: No writes may happen within a domain while an iterator exists over it.
func (a Table) PrefixScan(ctx HasKVStore, start, end RowID) (Iterator, error) {
	if start != nil && end != nil && bytes.Compare(start, end) >= 0 {
		return NewInvalidIterator(), errors.Wrap(ErrArgument, "start must be before end")
	}
	store := prefix.NewStore(ctx.KVStore(a.storeKey), []byte{a.prefix})
	return &typeSafeIterator{
		ctx:       ctx,
		rowGetter: NewTypeSafeRowGetter(a.storeKey, a.prefix, a.model, a.cdc),
		it:        store.Iterator(start, end),
	}, nil
}

// ReversePrefixScan returns an Iterator over a domain of keys in descending order. End is exclusive.
// Start is an MultiKeyIndex key or prefix. It must be less than end, or the Iterator is invalid  and error is returned.
// Iterator must be closed by caller.
// To iterate over entire domain, use PrefixScan(nil, nil)
//
// WARNING: The use of a ReversePrefixScan can be very expensive in terms of Gas. Please make sure you do not expose
// this as an endpoint to the public without further limits. See `LimitIterator`
//
// CONTRACT: No writes may happen within a domain while an iterator exists over it.
func (a Table) ReversePrefixScan(ctx HasKVStore, start, end RowID) (Iterator, error) {
	if start != nil && end != nil && bytes.Compare(start, end) >= 0 {
		return NewInvalidIterator(), errors.Wrap(ErrArgument, "start must be before end")
	}
	store := prefix.NewStore(ctx.KVStore(a.storeKey), []byte{a.prefix})
	return &typeSafeIterator{
		ctx:       ctx,
		rowGetter: NewTypeSafeRowGetter(a.storeKey, a.prefix, a.model, a.cdc),
		it:        store.ReverseIterator(start, end),
	}, nil
}

func (a Table) Table() Table {
	return a
}

// typeSafeIterator is initialized with a type safe RowGetter only.
type typeSafeIterator struct {
	ctx       HasKVStore
	rowGetter RowGetter
	it        types.Iterator
}

func (i typeSafeIterator) LoadNext(dest codec.ProtoMarshaler) (RowID, error) {
	if !i.it.Valid() {
		return nil, ErrIteratorDone
	}
	rowID := i.it.Key()
	i.it.Next()
	return rowID, i.rowGetter(i.ctx, rowID, dest)
}

func (i typeSafeIterator) Close() error {
	i.it.Close()
	return nil
}
