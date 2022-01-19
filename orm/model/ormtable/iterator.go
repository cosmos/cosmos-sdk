package ormtable

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/cosmos/cosmos-sdk/orm/encoding/ormkv"
	"github.com/cosmos/cosmos-sdk/orm/internal/listinternal"
	"github.com/cosmos/cosmos-sdk/orm/model/kv"
	"github.com/cosmos/cosmos-sdk/orm/model/ormlist"
)

// Iterator defines the interface for iterating over indexes.
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

	// Close closes the iterator and must always be called when done using
	// the iterator. The defer keyword should generally be used for this.
	Close()

	doNotImplement()
}

func iterator(
	backend ReadBackend,
	reader kv.ReadonlyStore,
	index concreteIndex,
	codec *ormkv.KeyCodec,
	options []listinternal.Option,
) (Iterator, error) {
	opts := &listinternal.Options{}
	listinternal.ApplyOptions(opts, options)
	if err := opts.Validate(); err != nil {
		return nil, err
	}

	if opts.Start != nil || opts.End != nil {
		err := codec.CheckValidRangeIterationKeys(opts.Start, opts.End)
		if err != nil {
			return nil, err
		}

		startBz, err := codec.EncodeKey(opts.Start)
		if err != nil {
			return nil, err
		}

		endBz, err := codec.EncodeKey(opts.End)
		if err != nil {
			return nil, err
		}

		fullEndKey := len(codec.GetFieldNames()) == len(opts.End)

		return rangeIterator(reader, backend, index, startBz, endBz, fullEndKey, opts)
	} else {
		prefixBz, err := codec.EncodeKey(opts.Prefix)
		if err != nil {
			return nil, err
		}

		return prefixIterator(reader, backend, index, prefixBz, opts)
	}
}

func prefixIterator(iteratorStore kv.ReadonlyStore, backend ReadBackend, index concreteIndex, prefix []byte, options *listinternal.Options) (Iterator, error) {
	if !options.Reverse {
		var start []byte
		if len(options.Cursor) != 0 {
			// must start right after cursor
			start = append(options.Cursor, 0x0)
		} else {
			start = prefix
		}
		end := prefixEndBytes(prefix)
		it, err := iteratorStore.Iterator(start, end)
		if err != nil {
			return nil, err
		}
		return &indexIterator{
			index:    index,
			store:    backend,
			iterator: it,
			started:  false,
		}, nil
	} else {
		var end []byte
		if len(options.Cursor) != 0 {
			// end bytes is already exclusive by default
			end = options.Cursor
		} else {
			end = prefixEndBytes(prefix)
		}
		it, err := iteratorStore.ReverseIterator(prefix, end)
		if err != nil {
			return nil, err
		}

		return &indexIterator{
			index:    index,
			store:    backend,
			iterator: it,
			started:  false,
		}, nil
	}
}

// NOTE: fullEndKey indicates whether the end key contained all the fields of the key,
// if it did then we need to use inclusive end bytes, otherwise we prefix the end bytes
func rangeIterator(iteratorStore kv.ReadonlyStore, reader ReadBackend, index concreteIndex, start, end []byte, fullEndKey bool, options *listinternal.Options) (Iterator, error) {
	if !options.Reverse {
		if len(options.Cursor) != 0 {
			start = append(options.Cursor, 0)
		}

		if fullEndKey {
			end = inclusiveEndBytes(end)
		} else {
			end = prefixEndBytes(end)
		}

		it, err := iteratorStore.Iterator(start, end)
		if err != nil {
			return nil, err
		}
		return &indexIterator{
			index:    index,
			store:    reader,
			iterator: it,
			started:  false,
		}, nil
	} else {
		if len(options.Cursor) != 0 {
			end = options.Cursor
		} else {
			if fullEndKey {
				end = inclusiveEndBytes(end)
			} else {
				end = prefixEndBytes(end)
			}
		}
		it, err := iteratorStore.ReverseIterator(start, end)
		if err != nil {
			return nil, err
		}

		return &indexIterator{
			index:    index,
			store:    reader,
			iterator: it,
			started:  false,
		}, nil
	}
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

func (i *indexIterator) Next() bool {
	if !i.started {
		i.started = true
	} else {
		i.iterator.Next()
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

func (indexIterator) doNotImplement() {

}

var _ Iterator = &indexIterator{}
