package ormtable

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/cosmos/cosmos-sdk/orm/model/kvstore"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
)

type Iterator interface {
	Next() bool
	Keys() (indexKey, primaryKey []protoreflect.Value, err error)
	UnmarshalMessage(proto.Message) error
	GetMessage() (proto.Message, error)

	Cursor() Cursor
	Close()

	doNotImplement()
}

type Cursor []byte

func prefixIterator(iteratorStore kvstore.ReadStore, store kvstore.IndexCommitmentReadStore, index concreteIndex, prefix []byte, options IteratorOptions) (Iterator, error) {
	if !options.Reverse {
		var start []byte
		if len(options.Cursor) != 0 {
			// must start right after cursor
			start = append(options.Cursor, 0x0)
		} else {
			start = prefix
		}
		end := storetypes.PrefixEndBytes(prefix)
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
			end = storetypes.PrefixEndBytes(prefix)
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

func rangeIterator(iteratorStore kvstore.ReadStore, store kvstore.IndexCommitmentReadStore, index concreteIndex, start, end []byte, options IteratorOptions) (Iterator, error) {
	if !options.Reverse {
		if len(options.Cursor) != 0 {
			start = append(options.Cursor, 0)
		}
		it, err := iteratorStore.Iterator(start, storetypes.PrefixEndBytes(end))
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
			end = storetypes.PrefixEndBytes(end)
		}
		it, err := iteratorStore.ReverseIterator(start, storetypes.InclusiveEndBytes(end))
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
	store    kvstore.IndexCommitmentReadStore
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
	return i.index.ReadValueFromIndexKey(i.store, pk, i.value, message)
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
