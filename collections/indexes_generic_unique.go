package collections

import (
	"context"
	"fmt"
)

// GenericUniqueIndex defines a generic index which enforces uniqueness constraints
// between ReferencingKey and ReferencedKey, meaning that one referencing key maps
// only one referenced key. The same referenced key can be mapped by multiple referencing keys.
//
// The referencing key can be anything, usually it is either a part of the primary
// key when we deal with multipart keys, or a field of Value.
//
// The referenced key usually is the primary key, or it can be a part
// of the primary key in the context of multipart keys.
//
// The referencing and referenced keys are mapped together using a Map.
//
// Unless you're trying to build your generic unique index, you should be using the indexes package.
type GenericUniqueIndex[ReferencingKey, ReferencedKey, PrimaryKey, Value any] struct {
	refs    Map[ReferencingKey, ReferencedKey]
	getRefs func(pk PrimaryKey, value Value) ([]IndexReference[ReferencingKey, ReferencedKey], error)
}

// NewGenericUniqueIndex instantiates a GenericUniqueIndex. Works in the same way as NewGenericMultiIndex.
func NewGenericUniqueIndex[ReferencingKey, ReferencedKey, PrimaryKey, Value any](
	schema *SchemaBuilder,
	prefix Prefix,
	name string,
	referencingKeyCodec KeyCodec[ReferencingKey],
	referencedKeyCodec KeyCodec[ReferencedKey],
	getRefs func(pk PrimaryKey, value Value) ([]IndexReference[ReferencingKey, ReferencedKey], error),
) *GenericUniqueIndex[ReferencingKey, ReferencedKey, PrimaryKey, Value] {
	return &GenericUniqueIndex[ReferencingKey, ReferencedKey, PrimaryKey, Value]{
		refs:    NewMap[ReferencingKey, ReferencedKey](schema, prefix, name, referencingKeyCodec, keyToValueCodec[ReferencedKey]{kc: referencedKeyCodec}),
		getRefs: getRefs,
	}
}

func (i *GenericUniqueIndex[ReferencingKey, ReferencedKey, PrimaryKey, Value]) Iterate(
	ctx context.Context,
	ranger Ranger[ReferencingKey],
) (Iterator[ReferencingKey, ReferencedKey], error) {
	return i.refs.Iterate(ctx, ranger)
}

func (i *GenericUniqueIndex[ReferencingKey, ReferencedKey, PrimaryKey, Value]) Get(ctx context.Context, ref ReferencingKey) (ReferencedKey, error) {
	return i.refs.Get(ctx, ref)
}

// Reference implements Index.
func (i *GenericUniqueIndex[ReferencingKey, ReferencedKey, PrimaryKey, Value]) Reference(
	ctx context.Context,
	pk PrimaryKey,
	newValue Value,
	oldValue *Value,
) error {
	if oldValue != nil {
		err := i.Unreference(ctx, pk, *oldValue)
		if err != nil {
			return err
		}
	}
	refs, err := i.getRefs(pk, newValue)
	if err != nil {
		return err
	}
	for _, ref := range refs {
		has, err := i.refs.Has(ctx, ref.Referring)
		if err != nil {
			return err
		}
		if has {
			return fmt.Errorf("%w: index uniqueness constrain violation: %s", ErrConflict, i.refs.kc.Stringify(ref.Referring))
		}
		err = i.refs.Set(ctx, ref.Referring, ref.Referred)
		if err != nil {
			return err
		}
	}
	return nil
}

// Unreference implements Index.
func (i *GenericUniqueIndex[ReferencingKey, ReferencedKey, PrimaryKey, Value]) Unreference(
	ctx context.Context,
	pk PrimaryKey,
	value Value,
) error {
	refs, err := i.getRefs(pk, value)
	if err != nil {
		return err
	}

	for _, ref := range refs {
		err = i.refs.Remove(ctx, ref.Referring)
		if err != nil {
			return err
		}
	}

	return nil
}

// keyToValueCodec is a ValueCodec that wraps a KeyCodec to make it behave like a ValueCodec.
type keyToValueCodec[K any] struct {
	kc KeyCodec[K]
}

func (k keyToValueCodec[K]) EncodeJSON(value K) ([]byte, error) {
	return k.kc.EncodeJSON(value)
}

func (k keyToValueCodec[K]) DecodeJSON(b []byte) (K, error) {
	return k.kc.DecodeJSON(b)
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
