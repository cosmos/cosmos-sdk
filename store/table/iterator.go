package table

import (
	"reflect"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/errors"
)

// IteratorFunc is a function type that satisfies the Iterator interface
// The passed function is called on LoadNext operations.
type IteratorFunc func(dest codec.ProtoMarshaler) (RowID, error)

// LoadNext loads the next value in the sequence into the pointer passed as dest and returns the key. If there
// are no more items the ErrIteratorDone error is returned
// The key is the rowID and not any MultiKeyIndex key.
func (i IteratorFunc) LoadNext(dest codec.ProtoMarshaler) (RowID, error) {
	return i(dest)
}

// Close always returns nil
func (i IteratorFunc) Close() error {
	return nil
}

func NewSingleValueIterator(rowID RowID, val []byte) Iterator {
	var closed bool
	return IteratorFunc(func(dest codec.ProtoMarshaler) (RowID, error) {
		if dest == nil {
			return nil, errors.Wrap(ErrArgument, "destination object must not be nil")
		}
		if closed || val == nil {
			return nil, ErrIteratorDone
		}
		closed = true
		return rowID, dest.Unmarshal(val)
	})
}

// Iterator that return ErrIteratorInvalid only.
func NewInvalidIterator() Iterator {
	return IteratorFunc(func(dest codec.ProtoMarshaler) (RowID, error) {
		return nil, ErrIteratorInvalid
	})
}

// LimitedIterator returns up to defined maximum number of elements.
type LimitedIterator struct {
	remainingCount int
	parentIterator Iterator
}

// LimitIterator returns a new iterator that returns max number of elements.
// The parent iterator must not be nil
// max can be 0 or any positive number
func LimitIterator(parent Iterator, max int) *LimitedIterator {
	if max < 0 {
		panic("quantity must not be negative")
	}
	if parent == nil {
		panic("parent iterator must not be nil")
	}
	return &LimitedIterator{remainingCount: max, parentIterator: parent}
}

// LoadNext loads the next value in the sequence into the pointer passed as dest and returns the key. If there
// are no more items or the defined max number of elements was returned the `ErrIteratorDone` error is returned
// The key is the rowID and not any MultiKeyIndex key.
func (i *LimitedIterator) LoadNext(dest codec.ProtoMarshaler) (RowID, error) {
	if i.remainingCount == 0 {
		return nil, ErrIteratorDone
	}
	i.remainingCount--
	return i.parentIterator.LoadNext(dest)
}

// Close releases the iterator and should be called at the end of iteration
func (i LimitedIterator) Close() error {
	return i.parentIterator.Close()
}

// First loads the first element into the given destination type and closes the iterator.
// When the iterator is closed or has no elements the according error is passed as return value.
func First(it Iterator, dest codec.ProtoMarshaler) (RowID, error) {
	if it == nil {
		return nil, errors.Wrap(ErrArgument, "iterator must not be nil")
	}
	defer it.Close()
	binKey, err := it.LoadNext(dest)
	if err != nil {
		return nil, err
	}
	return binKey, nil
}

// ModelSlicePtr represents a pointer to a slice of models. Think of it as
// *[]Model Because of Go's type system, using []Model type would not work for us.
// Instead we use a placeholder type and the validation is done during the
// runtime.
type ModelSlicePtr interface{}

// ReadAll consumes all values for the iterator and stores them in a new slice at the passed ModelSlicePtr.
// The slice can be empty when the iterator does not return any values but not nil. The iterator
// is closed afterwards.
// Example:
// 			var loaded []testdata.GroupInfo
//			rowIDs, err := ReadAll(it, &loaded)
//			require.NoError(t, err)
//
func ReadAll(it Iterator, dest ModelSlicePtr) ([]RowID, error) {
	if it == nil {
		return nil, errors.Wrap(ErrArgument, "iterator must not be nil")
	}
	defer it.Close()

	var destRef, tmpSlice reflect.Value
	elemType, err := assertDest(dest, &destRef, &tmpSlice)
	if err != nil {
		return nil, err
	}

	var rowIDs []RowID
	for {
		obj := reflect.New(elemType)
		val := obj.Elem()
		model := obj
		if elemType.Kind() == reflect.Ptr {
			val.Set(reflect.New(elemType.Elem()))
			model = val
		}

		binKey, err := it.LoadNext(model.Interface().(codec.ProtoMarshaler))
		switch {
		case err == nil:
			tmpSlice = reflect.Append(tmpSlice, val)
		case ErrIteratorDone.Is(err):
			destRef.Set(tmpSlice)
			return rowIDs, nil
		default:
			return nil, err
		}
		rowIDs = append(rowIDs, binKey)
	}
}

// assertDest checks that the provided dest is not nil and a pointer to a slice.
// It also verifies that the slice elements implement *codec.ProtoMarshaler.
// It overwrites destRef and tmpSlice using reflection.
func assertDest(dest ModelSlicePtr, destRef *reflect.Value, tmpSlice *reflect.Value) (reflect.Type, error) {
	if dest == nil {
		return nil, errors.Wrap(ErrArgument, "destination must not be nil")
	}
	tp := reflect.ValueOf(dest)
	if tp.Kind() != reflect.Ptr {
		return nil, errors.Wrap(ErrArgument, "destination must be a pointer to a slice")
	}
	if tp.Elem().Kind() != reflect.Slice {
		return nil, errors.Wrap(ErrArgument, "destination must point to a slice")
	}

	// Since dest is just an interface{}, we overwrite destRef using reflection
	// to have an assignable copy of it.
	*destRef = tp.Elem()
	// We need to verify that we can call Set() on destRef.
	if !destRef.CanSet() {
		return nil, errors.Wrap(ErrArgument, "destination not assignable")
	}

	elemType := reflect.TypeOf(dest).Elem().Elem()

	protoMarshaler := reflect.TypeOf((*codec.ProtoMarshaler)(nil)).Elem()
	if !elemType.Implements(protoMarshaler) &&
		!reflect.PtrTo(elemType).Implements(protoMarshaler) {
		return nil, errors.Wrapf(ErrArgument, "unsupported type :%s", elemType)
	}

	// tmpSlice is a slice value for the specified type
	// that we'll use for appending new elements.
	*tmpSlice = reflect.MakeSlice(reflect.SliceOf(elemType), 0, 0)

	return elemType, nil
}
