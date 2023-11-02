package ormtable

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	queryv1beta1 "cosmossdk.io/api/cosmos/base/query/v1beta1"
	"cosmossdk.io/orm/encoding/encodeutil"
	"cosmossdk.io/orm/encoding/ormkv"
	"cosmossdk.io/orm/internal/listinternal"
	"cosmossdk.io/orm/model/ormlist"
	"cosmossdk.io/orm/types/kv"
)

// Iterator defines the interface for iterating over indexes.
//
// WARNING: it is generally unsafe to mutate a table while iterating over it.
// Instead you should do reads and writes separately, or use a helper
// function like DeleteBy which does this efficiently.
type Iterator interface {
	// Next advances the iterator and returns true if a valid entry is found.
	// Next must be called before starting iteration.
	Next() bool

	// Keys returns the current index key and primary key values that the
	// iterator points to.
	Keys() (indexKey, primaryKey []protoreflect.Value, err error)

	// UnmarshalMessage unmarshals the entry the iterator currently points to
	// the provided proto.Message.
	UnmarshalMessage(proto.Message) error

	// GetMessage retrieves the proto.Message that the iterator currently points
	// to.
	GetMessage() (proto.Message, error)

	// Cursor returns the cursor referencing the current iteration position
	// and can be used to restart iteration right after this position.
	Cursor() ormlist.CursorT

	// PageResponse returns a non-nil page response after Next() returns false
	// if pagination was requested in list options.
	PageResponse() *queryv1beta1.PageResponse

	// Close closes the iterator and must always be called when done using
	// the iterator. The defer keyword should generally be used for this.
	Close()

	doNotImplement()
}

func prefixIterator(iteratorStore kv.ReadonlyStore, backend ReadBackend, index concreteIndex, codec *ormkv.KeyCodec, prefix []interface{}, opts []listinternal.Option) (Iterator, error) {
	options := &listinternal.Options{}
	listinternal.ApplyOptions(options, opts)
	if err := options.Validate(); err != nil {
		return nil, err
	}

	var prefixBz []byte
	prefixBz, err := codec.EncodeKey(encodeutil.ValuesOf(prefix...))
	if err != nil {
		return nil, err
	}

	var res Iterator
	if !options.Reverse {
		var start []byte
		if len(options.Cursor) != 0 {
			// must start right after cursor
			start = append(options.Cursor, 0x0)
		} else {
			start = prefixBz
		}
		end := prefixEndBytes(prefixBz)
		it, err := iteratorStore.Iterator(start, end)
		if err != nil {
			return nil, err
		}
		res = &indexIterator{
			index:    index,
			store:    backend,
			iterator: it,
			started:  false,
		}
	} else {
		var end []byte
		if len(options.Cursor) != 0 {
			// end bytes is already exclusive by default
			end = options.Cursor
		} else {
			end = prefixEndBytes(prefixBz)
		}
		it, err := iteratorStore.ReverseIterator(prefixBz, end)
		if err != nil {
			return nil, err
		}

		res = &indexIterator{
			index:    index,
			store:    backend,
			iterator: it,
			started:  false,
		}
	}

	return applyCommonIteratorOptions(res, options)
}

func rangeIterator(iteratorStore kv.ReadonlyStore, reader ReadBackend, index concreteIndex, codec *ormkv.KeyCodec, start, end []interface{}, opts []listinternal.Option) (Iterator, error) {
	options := &listinternal.Options{}
	listinternal.ApplyOptions(options, opts)
	if err := options.Validate(); err != nil {
		return nil, err
	}

	startValues := encodeutil.ValuesOf(start...)
	endValues := encodeutil.ValuesOf(end...)
	err := codec.CheckValidRangeIterationKeys(startValues, endValues)
	if err != nil {
		return nil, err
	}

	startBz, err := codec.EncodeKey(startValues)
	if err != nil {
		return nil, err
	}

	endBz, err := codec.EncodeKey(endValues)
	if err != nil {
		return nil, err
	}

	// NOTE: fullEndKey indicates whether the end key contained all the fields of the key,
	// if it did then we need to use inclusive end bytes, otherwise we prefix the end bytes
	fullEndKey := len(codec.GetFieldNames()) == len(end)

	var res Iterator
	if !options.Reverse {
		if len(options.Cursor) != 0 {
			startBz = append(options.Cursor, 0)
		}

		if fullEndKey {
			endBz = inclusiveEndBytes(endBz)
		} else {
			endBz = prefixEndBytes(endBz)
		}

		it, err := iteratorStore.Iterator(startBz, endBz)
		if err != nil {
			return nil, err
		}
		res = &indexIterator{
			index:    index,
			store:    reader,
			iterator: it,
			started:  false,
		}
	} else {
		if len(options.Cursor) != 0 {
			endBz = options.Cursor
		} else {
			if fullEndKey {
				endBz = inclusiveEndBytes(endBz)
			} else {
				endBz = prefixEndBytes(endBz)
			}
		}
		it, err := iteratorStore.ReverseIterator(startBz, endBz)
		if err != nil {
			return nil, err
		}

		res = &indexIterator{
			index:    index,
			store:    reader,
			iterator: it,
			started:  false,
		}
	}

	return applyCommonIteratorOptions(res, options)
}

func applyCommonIteratorOptions(iterator Iterator, options *listinternal.Options) (Iterator, error) {
	if options.Filter != nil {
		iterator = &filterIterator{Iterator: iterator, filter: options.Filter}
	}

	if options.CountTotal || options.Limit != 0 || options.Offset != 0 || options.DefaultLimit != 0 {
		iterator = paginate(iterator, options)
	}

	return iterator, nil
}

type indexIterator struct {
	index    concreteIndex
	store    ReadBackend
	iterator kv.Iterator

	indexValues []protoreflect.Value
	primaryKey  []protoreflect.Value
	value       []byte
	started     bool
}

func (i *indexIterator) PageResponse() *queryv1beta1.PageResponse {
	return nil
}

func (i *indexIterator) Next() bool {
	if !i.started {
		i.started = true
	} else {
		i.iterator.Next()
		i.indexValues = nil
	}

	return i.iterator.Valid()
}

func (i *indexIterator) Keys() (indexKey, primaryKey []protoreflect.Value, err error) {
	if i.indexValues != nil {
		return i.indexValues, i.primaryKey, nil
	}

	i.value = i.iterator.Value()
	i.indexValues, i.primaryKey, err = i.index.DecodeIndexKey(i.iterator.Key(), i.value)
	if err != nil {
		return nil, nil, err
	}

	return i.indexValues, i.primaryKey, nil
}

func (i indexIterator) UnmarshalMessage(message proto.Message) error {
	_, pk, err := i.Keys()
	if err != nil {
		return err
	}
	return i.index.readValueFromIndexKey(i.store, pk, i.value, message)
}

func (i *indexIterator) GetMessage() (proto.Message, error) {
	msg := i.index.MessageType().New().Interface()
	err := i.UnmarshalMessage(msg)
	return msg, err
}

func (i indexIterator) Cursor() ormlist.CursorT {
	return i.iterator.Key()
}

func (i indexIterator) Close() {
	err := i.iterator.Close()
	if err != nil {
		panic(err)
	}
}

func (indexIterator) doNotImplement() {}

var _ Iterator = &indexIterator{}
