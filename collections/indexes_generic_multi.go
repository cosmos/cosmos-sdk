package collections

import (
	"context"

	"cosmossdk.io/collections/codec"
)

func NewIndexReference[ReferencingKey, ReferencedKey any](referencing ReferencingKey, referenced ReferencedKey) IndexReference[ReferencingKey, ReferencedKey] {
	return IndexReference[ReferencingKey, ReferencedKey]{
		Referring: referencing,
		Referred:  referenced,
	}
}

// IndexReference defines a generic index reference.
type IndexReference[ReferencingKey, ReferencedKey any] struct {
	// Referring is the key that refers, points to the Referred key.
	Referring ReferencingKey
	// Referred is the key that is being pointed to by the Referring key.
	Referred ReferencedKey
}

// GenericMultiIndex defines a generic Index type that given a primary key
// and the value associated with that primary key returns one or multiple IndexReference.
//
// The referencing key can be anything, usually it is either a part of the primary
// key when we deal with multipart keys, or a field of Value.
//
// The referenced key usually is the primary key, or it can be a part
// of the primary key in the context of multipart keys.
//
// The Referencing and Referenced keys are joined and saved as a Pair in a KeySet
// where the key is Pair[ReferencingKey, ReferencedKey].
// So if we wanted to get all the keys referenced by a generic (concrete) ReferencingKey
// we would just need to iterate over all the keys starting with bytes(ReferencingKey).
//
// Unless you're trying to build your generic multi index, you should be using the indexes package.
type GenericMultiIndex[ReferencingKey, ReferencedKey, PrimaryKey, Value any] struct {
	refs    KeySet[Pair[ReferencingKey, ReferencedKey]]
	getRefs func(pk PrimaryKey, v Value) ([]IndexReference[ReferencingKey, ReferencedKey], error)
}

// NewGenericMultiIndex instantiates a GenericMultiIndex, given
// schema, Prefix, humanised name, the key codec used to encode the referencing key
// to bytes, the key codec used to encode the referenced key to bytes and a function
// which given the primary key and a value of an object being saved or removed in IndexedMap
// returns all the possible IndexReference of that object.
//
// The IndexReference is usually just one. But in certain cases can be multiple,
// for example when the Value has an array field, and we want to create a relationship
// between the object and all the elements of the array contained in the object.
func NewGenericMultiIndex[ReferencingKey, ReferencedKey, PrimaryKey, Value any](
	schema *SchemaBuilder,
	prefix Prefix,
	name string,
	referencingKeyCodec codec.KeyCodec[ReferencingKey],
	referencedKeyCodec codec.KeyCodec[ReferencedKey],
	getRefsFunc func(pk PrimaryKey, value Value) ([]IndexReference[ReferencingKey, ReferencedKey], error),
) *GenericMultiIndex[ReferencingKey, ReferencedKey, PrimaryKey, Value] {
	return &GenericMultiIndex[ReferencingKey, ReferencedKey, PrimaryKey, Value]{
		getRefs: getRefsFunc,
		refs:    NewKeySet(schema, prefix, name, PairKeyCodec(referencingKeyCodec, referencedKeyCodec)),
	}
}

// Iterate allows to iterate over the index. It returns a KeySetIterator of Pair[ReferencingKey, ReferencedKey].
// K1 of the Pair is the key (referencing) pointing to K2 (referenced).
func (i *GenericMultiIndex[ReferencingKey, ReferencedKey, PrimaryKey, Value]) Iterate(
	ctx context.Context,
	ranger Ranger[Pair[ReferencingKey, ReferencedKey]],
) (KeySetIterator[Pair[ReferencingKey, ReferencedKey]], error) {
	return i.refs.Iterate(ctx, ranger)
}

// Has reports if there is a relationship in the index between the referencing and the referenced key.
func (i *GenericMultiIndex[ReferencingKey, ReferencedKey, PrimaryKey, Value]) Has(
	ctx context.Context,
	referencing ReferencingKey,
	referenced ReferencedKey,
) (bool, error) {
	return i.refs.Has(ctx, Join(referencing, referenced))
}

// Reference implements the Index interface.
func (i *GenericMultiIndex[ReferencingKey, ReferencedKey, PrimaryKey, Value]) Reference(
	ctx context.Context,
	pk PrimaryKey,
	value Value,
	oldValue *Value,
) error {
	if oldValue != nil {
		err := i.Unreference(ctx, pk, *oldValue)
		if err != nil {
			return err
		}
	}

	refKeys, err := i.getRefs(pk, value)
	if err != nil {
		return err
	}

	for _, ref := range refKeys {
		err := i.refs.Set(ctx, Join(ref.Referring, ref.Referred))
		if err != nil {
			return err
		}
	}

	return nil
}

// Unreference implements the Index interface.
func (i *GenericMultiIndex[ReferencingKey, ReferencedKey, PrimaryKey, Value]) Unreference(
	ctx context.Context,
	pk PrimaryKey,
	value Value,
) error {
	refs, err := i.getRefs(pk, value)
	if err != nil {
		return err
	}

	for _, ref := range refs {
		err = i.refs.Remove(ctx, Join(ref.Referring, ref.Referred))
		if err != nil {
			return err
		}
	}

	return nil
}

func (i *GenericMultiIndex[ReferencingKey, ReferencedKey, PrimaryKey, Value]) IterateRaw(
	ctx context.Context,
	start, end []byte,
	order Order,
) (Iterator[Pair[ReferencingKey, ReferencedKey], NoValue], error) {
	return i.refs.IterateRaw(ctx, start, end, order)
}

func (i *GenericMultiIndex[ReferencingKey, ReferencedKey, PrimaryKey, Value]) Walk(
	ctx context.Context,
	ranger Ranger[Pair[ReferencingKey, ReferencedKey]],
	walkFunc func(referencingKey ReferencingKey, referencedKey ReferencedKey) bool,
) error {
	return i.refs.Walk(ctx, ranger, func(key Pair[ReferencingKey, ReferencedKey]) bool { return walkFunc(key.K1(), key.K2()) })
}

func (i *GenericMultiIndex[ReferencingKey, ReferencedKey, PrimaryKey, Value]) KeyCodec() codec.KeyCodec[Pair[ReferencingKey, ReferencedKey]] {
	return i.refs.KeyCodec()
}

func (i *GenericMultiIndex[ReferencingKey, ReferencedKey, PrimaryKey, Value]) ValueCodec() codec.ValueCodec[NoValue] {
	return i.refs.ValueCodec()
}
