/*
Package store/table is a convenient object to data store mapper.
*/
package table

import (
	"io"
	"reflect"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
)

const tableCodespace = "table"

var (
	ErrNotFound          = errors.Register(tableCodespace, 100, "not found")
	ErrIteratorDone      = errors.Register(tableCodespace, 101, "iterator done")
	ErrIteratorInvalid   = errors.Register(tableCodespace, 102, "iterator invalid")
	ErrType              = errors.Register(tableCodespace, 110, "invalid type")
	ErrUniqueConstraint  = errors.Register(tableCodespace, 111, "unique constraint violation")
	ErrArgument          = errors.Register(tableCodespace, 112, "invalid argument")
	ErrIndexKeyMaxLength = errors.Register(tableCodespace, 113, "index key exceeds max length")
)

// Unique identifier of a persistent table.
type RowID []byte

// Bytes returns raw bytes.
func (r RowID) Bytes() []byte {
	return r
}

// Validateable is an interface that ProtoMarshaler types can implement and is called on any orm save or update operation.
type Validateable interface {
	// ValidateBasic is a sanity check on the data. Any error returned prevents create or updates.
	ValidateBasic() error
}

// Iterator allows iteration through a sequence of key value pairs
type Iterator interface {
	// LoadNext loads the next value in the sequence into the pointer passed as dest and returns the key. If there
	// are no more items the ErrIteratorDone error is returned
	// The key is the rowID.
	LoadNext(dest codec.ProtoMarshaler) (RowID, error)
	// Close releases the iterator and should be called at the end of iteration
	io.Closer
}

// IndexKeyCodec defines the encoding/decoding methods for building/splitting index keys.
type IndexKeyCodec interface {
	// BuildIndexKey encodes a searchable key and the target RowID.
	BuildIndexKey(searchableKey []byte, rowID RowID) []byte
	// StripRowID returns the RowID from the combined persistentIndexKey. It is the reverse operation to BuildIndexKey
	// but with the searchableKey dropped.
	StripRowID(persistentIndexKey []byte) RowID
}

// Indexable types are used to setup new tables.
// This interface provides a set of functions that can be called by indexes to register and interact with the tables.
type Indexable interface {
	RowGetter() RowGetter
	IndexKeyCodec() IndexKeyCodec
	AddAfterSaveInterceptor(interceptor AfterSaveInterceptor)
	AddAfterDeleteInterceptor(interceptor AfterDeleteInterceptor)
}

// AfterSaveInterceptor defines a callback function to be called on Create + Update.
type AfterSaveInterceptor func(store sdk.KVStore, rowID RowID, newValue, oldValue codec.ProtoMarshaler) error

// AfterDeleteInterceptor defines a callback function to be called on Delete operations.
type AfterDeleteInterceptor func(store sdk.KVStore, rowID RowID, value codec.ProtoMarshaler) error

// RowGetter loads a persistent object by row ID into the destination object. The dest parameter must therefore be a pointer.
// Any implementation must return `ErrNotFound` when no object for the rowID exists
type RowGetter func(store sdk.KVStore, rowID RowID, dest codec.ProtoMarshaler) error

// NewTypeSafeRowGetter returns a `RowGetter` with type check on the dest parameter.
func NewTypeSafeRowGetter(prefixKey byte, model reflect.Type, cdc codec.Codec) RowGetter {
	return func(store sdk.KVStore, rowID RowID, dest codec.ProtoMarshaler) error {
		if len(rowID) == 0 {
			return errors.Wrap(ErrArgument, "key must not be nil")
		}
		if err := assertCorrectType(model, dest); err != nil {
			return err
		}

		pStore := prefix.NewStore(store, []byte{prefixKey})
		it := pStore.Iterator(PrefixRange(rowID))
		defer it.Close()
		if !it.Valid() {
			return ErrNotFound
		}
		return cdc.Unmarshal(it.Value(), dest)
	}
}

func assertCorrectType(model reflect.Type, obj codec.ProtoMarshaler) error {
	tp := reflect.TypeOf(obj)
	if tp.Kind() != reflect.Ptr {
		return errors.Wrap(ErrType, "model destination must be a pointer")
	}
	if model != tp.Elem() {
		return errors.Wrapf(ErrType, "can not use %T with this bucket", obj)
	}
	return nil
}
