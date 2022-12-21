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

func (i *UniqueIndex[ReferenceKey, PrimaryKey, Value]) Find(ctx context.Context, ref ReferenceKey) (PrimaryKey, error) {
	return i.refs.Get(ctx, ref)
}

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
