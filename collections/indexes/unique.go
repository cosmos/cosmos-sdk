package indexes

import (
	"context"
	"cosmossdk.io/collections"
)

type Unique[ReferenceKey, PrimaryKey, Value any] collections.GenericUniqueIndex[ReferenceKey, PrimaryKey, PrimaryKey, Value]

func NewUnique[ReferenceKey, PrimaryKey, Value any](
	schema *collections.SchemaBuilder,
	prefix collections.Prefix,
	name string,
	refCodec collections.KeyCodec[ReferenceKey],
	pkCodec collections.KeyCodec[PrimaryKey],
	getRefKeyFunc func(pk PrimaryKey, v Value) (ReferenceKey, error),
) *Unique[ReferenceKey, PrimaryKey, Value] {
	i := collections.NewGenericUniqueIndex(schema, prefix, name, refCodec, pkCodec, func(pk PrimaryKey, value Value) ([]collections.IndexReference[ReferenceKey, PrimaryKey], error) {
		ref, err := getRefKeyFunc(pk, value)
		if err != nil {
			return nil, err
		}

		return []collections.IndexReference[ReferenceKey, PrimaryKey]{
			{
				Referring: ref,
				Referred:  pk,
			},
		}, nil
	})

	return (*Unique[ReferenceKey, PrimaryKey, Value])(i)
}

func (i *Unique[ReferenceKey, PrimaryKey, Value]) Reference(ctx context.Context, pk PrimaryKey, newValue Value, oldValue *Value) error {
	return (*collections.GenericUniqueIndex[ReferenceKey, PrimaryKey, PrimaryKey, Value])(i).Reference(ctx, pk, newValue, oldValue)
}

func (i *Unique[ReferenceKey, PrimaryKey, Value]) Unreference(ctx context.Context, pk PrimaryKey, value Value) error {
	return (*collections.GenericUniqueIndex[ReferenceKey, PrimaryKey, PrimaryKey, Value])(i).Unreference(ctx, pk, value)
}

func (i *Unique[ReferenceKey, PrimaryKey, Value]) ExactMatch(ctx context.Context, ref ReferenceKey) (PrimaryKey, error) {
	return (*collections.GenericUniqueIndex[ReferenceKey, PrimaryKey, PrimaryKey, Value])(i).Get(ctx, ref)
}

func (i *Unique[ReferenceKey, PrimaryKey, Value]) Iterate(ctx context.Context, ranger collections.Ranger[ReferenceKey]) (UniqueIterator[ReferenceKey, PrimaryKey], error) {
	iter, err := (*collections.GenericUniqueIndex[ReferenceKey, PrimaryKey, PrimaryKey, Value])(i).Iterate(ctx, ranger)
	return (UniqueIterator[ReferenceKey, PrimaryKey])(iter), err
}

// UniqueIterator is an Iterator wrapper, that exposes only the functionality needed to work with Unique keys.
type UniqueIterator[ReferenceKey, PrimaryKey any] collections.Iterator[ReferenceKey, PrimaryKey]

// PrimaryKey returns the iterator's current primary key.
func (i UniqueIterator[ReferenceKey, PrimaryKey]) PrimaryKey() (PrimaryKey, error) {
	return (collections.Iterator[ReferenceKey, PrimaryKey])(i).Value()
}

// PrimaryKeys fully consumes the iterator, and returns all the primary keys.
func (i UniqueIterator[ReferenceKey, PrimaryKey]) PrimaryKeys() ([]PrimaryKey, error) {
	return (collections.Iterator[ReferenceKey, PrimaryKey])(i).Values()
}

// FullKey returns the iterator's current full reference key as Pair[ReferenceKey, PrimaryKey].
func (i UniqueIterator[ReferenceKey, PrimaryKey]) FullKey() (collections.Pair[ReferenceKey, PrimaryKey], error) {
	kv, err := (collections.Iterator[ReferenceKey, PrimaryKey])(i).KeyValue()
	return collections.Join(kv.Key, kv.Value), err
}

func (i UniqueIterator[ReferenceKey, PrimaryKey]) FullKeys() ([]collections.Pair[ReferenceKey, PrimaryKey], error) {
	kvs, err := (collections.Iterator[ReferenceKey, PrimaryKey])(i).KeyValues()
	if err != nil {
		return nil, err
	}
	pairKeys := make([]collections.Pair[ReferenceKey, PrimaryKey], len(kvs))
	for index := range kvs {
		kv := kvs[index]
		pairKeys[index] = collections.Join(kv.Key, kv.Value)
	}
	return pairKeys, nil
}
