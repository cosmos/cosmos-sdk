package ormtable

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/cosmos/cosmos-sdk/orm/backend/kv"
	"github.com/cosmos/cosmos-sdk/orm/model/ormiterator"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
)

func prefixIterator(iteratorStore kv.ReadStore, store kv.IndexCommitmentReadStore, index Index, prefix []byte, options IteratorOptions) (ormiterator.Iterator, error) {
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

func rangeIterator(iteratorStore kv.ReadStore, store kv.IndexCommitmentReadStore, index Index, start, end []byte, options IteratorOptions) (ormiterator.Iterator, error) {
	if !options.Reverse {
		if len(options.Cursor) != 0 {
			start = append(options.Cursor, 0)
		}
		it, err := iteratorStore.Iterator(start, storetypes.InclusiveEndBytes(end))
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
	ormiterator.UnimplementedIterator

	index    Index
	store    kv.IndexCommitmentReadStore
	iterator kv.Iterator

	indexValues []protoreflect.Value
	primaryKey  []protoreflect.Value
	value       []byte
	started     bool
}

func (i *indexIterator) Next() (bool, error) {
	if !i.started {
		i.started = true
	} else {
		i.iterator.Next()
	}

	if !i.iterator.Valid() {
		return false, nil
	}

	var err error
	i.value = i.iterator.Value()
	i.indexValues, i.primaryKey, err = i.index.DecodeIndexKey(i.iterator.Key(), i.value)
	if err != nil {
		return true, err
	}

	return true, err
}

func (i indexIterator) IndexKey() ([]protoreflect.Value, error) {
	return i.indexValues, nil
}

func (i indexIterator) PrimaryKey() ([]protoreflect.Value, error) {
	return i.primaryKey, nil
}

func (i indexIterator) GetMessage(message proto.Message) error {
	return i.index.ReadValueFromIndexKey(i.store, i.primaryKey, i.value, message)
}

func (i indexIterator) Cursor() ormiterator.Cursor {
	return i.iterator.Key()
}

func (i indexIterator) Close() {
	err := i.iterator.Close()
	if err != nil {
		panic(err)
	}
}

var _ ormiterator.Iterator = &indexIterator{}
