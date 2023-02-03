package indexes

import (
	"context"
	"cosmossdk.io/collections"
)

// Multi defines the most common index. It can be used to create a reference between
// a field of value and its primary key. Multiple primary keys can be mapped to the same
// reference key as the index does not enforce uniqueness constraints.
type Multi[ReferenceKey, PrimaryKey, Value any] collections.GenericMultiIndex[ReferenceKey, PrimaryKey, PrimaryKey, Value]

// NewMulti instantiates a new Multi instance given a schema,
// a Prefix, the humanized name for the index, the reference key key codec
// and the primary key key codec. The getRefKeyFunc is a function that
// given the primary key and value returns the referencing key.
func NewMulti[ReferenceKey, PrimaryKey, Value any](
	schema *collections.SchemaBuilder,
	prefix collections.Prefix,
	name string,
	refCodec collections.KeyCodec[ReferenceKey],
	pkCodec collections.KeyCodec[PrimaryKey],
	getRefKeyFunc func(pk PrimaryKey, value Value) (ReferenceKey, error),
) *Multi[ReferenceKey, PrimaryKey, Value] {
	i := collections.NewGenericMultiIndex(
		schema, prefix, name, refCodec, pkCodec,
		func(pk PrimaryKey, value Value) ([]collections.IndexReference[ReferenceKey, PrimaryKey], error) {
			ref, err := getRefKeyFunc(pk, value)
			if err != nil {
				return nil, err
			}
			return []collections.IndexReference[ReferenceKey, PrimaryKey]{
				collections.NewIndexReference(ref, pk),
			}, nil
		},
	)

	return (*Multi[ReferenceKey, PrimaryKey, Value])(i)
}

func (m *Multi[ReferenceKey, PrimaryKey, Value]) Reference(ctx context.Context, pk PrimaryKey, newValue Value, oldValue *Value) error {
	return (*collections.GenericMultiIndex[ReferenceKey, PrimaryKey, PrimaryKey, Value])(m).Reference(ctx, pk, newValue, oldValue)
}

func (m *Multi[ReferenceKey, PrimaryKey, Value]) Unreference(ctx context.Context, pk PrimaryKey, value Value) error {
	return (*collections.GenericMultiIndex[ReferenceKey, PrimaryKey, PrimaryKey, Value])(m).Unreference(ctx, pk, value)
}

func (m *Multi[ReferenceKey, PrimaryKey, Value]) Iterate(ctx context.Context, ranger collections.Ranger[collections.Pair[ReferenceKey, PrimaryKey]]) (MultiIterator[ReferenceKey, PrimaryKey], error) {
	iter, err := (*collections.GenericMultiIndex[ReferenceKey, PrimaryKey, PrimaryKey, Value])(m).Iterate(ctx, ranger)
	return (MultiIterator[ReferenceKey, PrimaryKey])(iter), err
}

// MatchExact returns a MultiIterator containing all the primary keys referenced by the provided reference key.
func (m *Multi[ReferenceKey, PrimaryKey, Value]) MatchExact(ctx context.Context, refKey ReferenceKey) (MultiIterator[ReferenceKey, PrimaryKey], error) {
	return m.Iterate(ctx, collections.NewPrefixedPairRange[ReferenceKey, PrimaryKey](refKey))
}

// MultiIterator is just a KeySetIterator with key as Pair[ReferenceKey, PrimaryKey].
type MultiIterator[ReferenceKey, PrimaryKey any] collections.KeySetIterator[collections.Pair[ReferenceKey, PrimaryKey]]

// PrimaryKey returns the iterator's current primary key.
func (i MultiIterator[ReferenceKey, PrimaryKey]) PrimaryKey() (PrimaryKey, error) {
	fullKey, err := i.FullKey()
	return fullKey.K2(), err
}

// PrimaryKeys fully consumes the iterator and returns the list of primary keys.
func (i MultiIterator[ReferenceKey, PrimaryKey]) PrimaryKeys() ([]PrimaryKey, error) {
	fullKeys, err := i.FullKeys()
	if err != nil {
		return nil, err
	}
	pks := make([]PrimaryKey, len(fullKeys))
	for i, fullKey := range fullKeys {
		pks[i] = fullKey.K2()
	}
	return pks, nil
}

// FullKey returns the current full reference key as Pair[ReferenceKey, PrimaryKey].
func (i MultiIterator[ReferenceKey, PrimaryKey]) FullKey() (collections.Pair[ReferenceKey, PrimaryKey], error) {
	return (collections.KeySetIterator[collections.Pair[ReferenceKey, PrimaryKey]])(i).Key()
}

// FullKeys fully consumes the iterator and returns all the list of full reference keys.
func (i MultiIterator[ReferenceKey, PrimaryKey]) FullKeys() ([]collections.Pair[ReferenceKey, PrimaryKey], error) {
	return (collections.KeySetIterator[collections.Pair[ReferenceKey, PrimaryKey]])(i).Keys()
}

// Next advances the iterator.
func (i MultiIterator[ReferenceKey, PrimaryKey]) Next() {
	(collections.KeySetIterator[collections.Pair[ReferenceKey, PrimaryKey]])(i).Next()
}

// Valid asserts if the iterator is still valid or not.
func (i MultiIterator[ReferenceKey, PrimaryKey]) Valid() bool {
	return (collections.KeySetIterator[collections.Pair[ReferenceKey, PrimaryKey]])(i).Valid()
}

// Close closes the iterator.
func (i MultiIterator[ReferenceKey, PrimaryKey]) Close() error {
	return (collections.KeySetIterator[collections.Pair[ReferenceKey, PrimaryKey]])(i).Close()
}
