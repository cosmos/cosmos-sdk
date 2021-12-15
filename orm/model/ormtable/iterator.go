package ormtable

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/cosmos/cosmos-sdk/orm/model/kvstore"
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
	// and can be passed to IteratorOptions to restart iteration right after
	// this position.
	Cursor() Cursor

	// Close closes the iterator and must always be called when done using
	// the iterator. The defer keyword should generally be used for this.
	Close()

	doNotImplement()
}

// Cursor defines the cursor type.
type Cursor []byte

func prefixIterator(iteratorStore kvstore.Reader, store kvstore.ReadBackend, index concreteIndex, prefix []byte, options IteratorOptions) (Iterator, error) {
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
			store:    store,
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
			store:    store,
			iterator: it,
			started:  false,
		}, nil
	}
}

// NOTE: fullEndKey indicates whether the end key contained all the fields of the key,
// if it did then we need to use inclusive end bytes, otherwise we prefix the end bytes
func rangeIterator(iteratorStore kvstore.Reader, store kvstore.ReadBackend, index concreteIndex, start, end []byte, fullEndKey bool, options IteratorOptions) (Iterator, error) {
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
			store:    store,
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
			store:    store,
			iterator: it,
			started:  false,
		}, nil
	}
}

type indexIterator struct {
	index    concreteIndex
	store    kvstore.ReadBackend
	iterator kvstore.Iterator

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

	if !i.iterator.Valid() {
		return false
	}

	return true
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

func (i indexIterator) Cursor() Cursor {
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
