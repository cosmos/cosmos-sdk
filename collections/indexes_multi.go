package collections

import (
	"context"
)

// MultiIndex is an Index that can map the same ReferenceKey to multiple PrimaryKey.
type MultiIndex[ReferenceKey, PrimaryKey, Value any] struct {
	getRefKey func(value Value) (ReferenceKey, error)
	refs      KeySet[Pair[ReferenceKey, PrimaryKey]]
}

// NewMultiIndex instantiates a new MultiIndex instance.
func NewMultiIndex[ReferenceKey, PrimaryKey, Value any](
	schema Schema,
	prefix Prefix,
	name string,
	refCodec KeyCodec[ReferenceKey],
	pkCodec KeyCodec[PrimaryKey],
	getRefKeyFunc func(value Value) (ReferenceKey, error),
) *MultiIndex[ReferenceKey, PrimaryKey, Value] {
	return &MultiIndex[ReferenceKey, PrimaryKey, Value]{
		getRefKey: getRefKeyFunc,
		refs:      NewKeySet(schema, prefix, name, PairKeyCodec(refCodec, pkCodec)),
	}
}

func (m *MultiIndex[ReferenceKey, PrimaryKey, Value]) Reference(ctx context.Context, pk PrimaryKey, newValue Value, oldValue *Value) error {
	// if old value is not nil then we need to remove the old reference.
	if oldValue != nil {
		err := m.Unreference(ctx, pk, *oldValue)
		if err != nil {
			return err
		}
	}

	// ref the new value
	refKey, err := m.getRefKey(newValue)
	if err != nil {
		return err
	}
	return m.refs.Set(ctx, Join(refKey, pk))
}

func (m *MultiIndex[ReferenceKey, PrimaryKey, Value]) Unreference(ctx context.Context, pk PrimaryKey, value Value) error {
	refKey, err := m.getRefKey(value)
	if err != nil {
		return err
	}
	err = m.refs.Remove(ctx, Join(refKey, pk))
	if err != nil {
		return err
	}
	return nil
}

func (m *MultiIndex[ReferenceKey, PrimaryKey, Value]) Iterate(ctx context.Context, ranger Ranger[Pair[ReferenceKey, PrimaryKey]]) (MultiIndexIterator[ReferenceKey, PrimaryKey], error) {
	iter, err := m.refs.Iterate(ctx, ranger)
	return (MultiIndexIterator[ReferenceKey, PrimaryKey])(iter), err
}

func (m *MultiIndex[ReferenceKey, PrimaryKey, Value]) ExactMatch(ctx context.Context, refKey ReferenceKey) (MultiIndexIterator[ReferenceKey, PrimaryKey], error) {
	return m.Iterate(ctx, new(PairRange[ReferenceKey, PrimaryKey]).Prefix(refKey))
}

// MultiIndexIterator is just a KeySetIterator with key as Pair[ReferenceKey, PrimaryKey].
type MultiIndexIterator[ReferenceKey, PrimaryKey any] KeySetIterator[Pair[ReferenceKey, PrimaryKey]]

func (i MultiIndexIterator[ReferenceKey, PrimaryKey]) PrimaryKey() PrimaryKey {
	return (KeySetIterator[Pair[ReferenceKey, PrimaryKey]])(i).Key()
}

func (i MultiIndexIterator[ReferenceKey, PrimaryKey]) Next() {
	(KeySetIterator[Pair[ReferenceKey, PrimaryKey]])(i).Next()
}
func (i MultiIndexIterator[ReferenceKey, PrimaryKey]) Valid() bool {
	return (KeySetIterator[Pair[ReferenceKey, PrimaryKey]])(i).Valid()
}
