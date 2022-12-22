package collections

import (
	"context"
	"fmt"
)

type UniqueIndex[ReferenceKey, PrimaryKey, Value any] struct {
	getRefKey func(value Value) (ReferenceKey, error)
	refs      Map[ReferenceKey, PrimaryKey]
}

func NewUniqueIndex[ReferenceKey, PrimaryKey, Value any](
	schema Schema,
	prefix Prefix,
	name string,
	refCodec KeyCodec[ReferenceKey],
	pkCodec KeyCodec[PrimaryKey],
	getRefKeyFunc func(v Value) (ReferenceKey, error),
) *UniqueIndex[ReferenceKey, PrimaryKey, Value] {
	return &UniqueIndex[ReferenceKey, PrimaryKey, Value]{
		getRefKey: getRefKeyFunc,
		refs:      newMap[ReferenceKey, PrimaryKey](schema, prefix, name, refCodec, keyToValueCodec[PrimaryKey]{kc: pkCodec}),
	}
}

func (i *UniqueIndex[ReferenceKey, PrimaryKey, Value]) Reference(ctx context.Context, pk PrimaryKey, newValue Value, oldValue *Value) error {
	if oldValue != nil {
		err := i.Unreference(ctx, pk, *oldValue)
		if err != nil {
			return err
		}
	}
	refKey, err := i.getRefKey(newValue)
	if err != nil {
		return err
	}
	has, err := i.refs.Has(ctx, refKey)
	if err != nil {
		return err
	}
	if has {
		return fmt.Errorf("%w: uniqueness contraint violation: key %s already references a primary key", ErrConflict, i.refs.kc.Stringify(refKey))
	}

	return i.refs.Set(ctx, refKey, pk)
}

func (i *UniqueIndex[ReferenceKey, PrimaryKey, Value]) Unreference(ctx context.Context, _ PrimaryKey, value Value) error {
	refKey, err := i.getRefKey(value)
	if err != nil {
		return err
	}
	return i.refs.Remove(ctx, refKey)
}

func (i *UniqueIndex[ReferenceKey, PrimaryKey, Value]) ExactMatch(ctx context.Context, ref ReferenceKey) (PrimaryKey, error) {
	return i.refs.Get(ctx, ref)
}

func (i *UniqueIndex[ReferenceKey, PrimaryKey, Value]) Iterate(ctx context.Context, ranger Ranger[ReferenceKey]) (UniqueIndexIterator[ReferenceKey, PrimaryKey], error) {
	iter, err := i.refs.Iterate(ctx, ranger)
	return (UniqueIndexIterator[ReferenceKey, PrimaryKey])(iter), err
}

// UniqueIndexIterator is an Iterator wrapper, that exposes only the functionality needed to work with UniqueIndex keys.
type UniqueIndexIterator[ReferenceKey, PrimaryKey any] Iterator[ReferenceKey, PrimaryKey]

// PrimaryKey returns the iterator's current primary key.
func (i UniqueIndexIterator[ReferenceKey, PrimaryKey]) PrimaryKey() (PrimaryKey, error) {
	return (Iterator[ReferenceKey, PrimaryKey])(i).Value()
}

// PrimaryKeys fully consumes the iterator, and returns all the primary keys.
func (i UniqueIndexIterator[ReferenceKey, PrimaryKey]) PrimaryKeys() ([]PrimaryKey, error) {
	return (Iterator[ReferenceKey, PrimaryKey])(i).Values()
}

// FullKey returns the iterator's current full reference key as Pair[ReferenceKey, PrimaryKey].
func (i UniqueIndexIterator[ReferenceKey, PrimaryKey]) FullKey() (Pair[ReferenceKey, PrimaryKey], error) {
	kv, err := (Iterator[ReferenceKey, PrimaryKey])(i).KeyValue()
	return Join(kv.Key, kv.Value), err
}

func (i UniqueIndexIterator[ReferenceKey, PrimaryKey]) FullKeys() ([]Pair[ReferenceKey, PrimaryKey], error) {
	kvs, err := (Iterator[ReferenceKey, PrimaryKey])(i).KeyValues()
	if err != nil {
		return nil, err
	}
	pairKeys := make([]Pair[ReferenceKey, PrimaryKey], len(kvs))
	for index := range kvs {
		kv := kvs[index]
		pairKeys[index] = Join(kv.Key, kv.Value)
	}
	return pairKeys, nil
}

// keyToValueCodec is a ValueCodec that wraps a KeyCodec to make it behave like a ValueCodec.
type keyToValueCodec[K any] struct {
	kc KeyCodec[K]
}

func (k keyToValueCodec[K]) Encode(value K) ([]byte, error) {
	buf := make([]byte, k.kc.Size(value))
	_, err := k.kc.Encode(buf, value)
	return buf, err
}

func (k keyToValueCodec[K]) Decode(b []byte) (K, error) {
	r, key, err := k.kc.Decode(b)
	if err != nil {
		var key K
		return key, err
	}

	if r != len(b) {
		var key K
		return key, fmt.Errorf("%w: was supposed to fully consume the key '%x', consumed %d out of %d", ErrEncoding, b, r, len(b))
	}
	return key, nil
}

func (k keyToValueCodec[K]) Stringify(value K) string {
	return k.kc.Stringify(value)
}

func (k keyToValueCodec[K]) ValueType() string {
	return k.kc.KeyType()
}
