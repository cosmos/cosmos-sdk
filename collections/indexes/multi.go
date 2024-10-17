package indexes

import (
	"context"
	"crypto/sha256"
	"errors"
	"unsafe"

	"cosmossdk.io/collections"
	"cosmossdk.io/collections/codec"
)

type multiOptions struct {
	uncheckedValue bool
}

// WithMultiUncheckedValue is an option that can be passed to NewMulti to
// ignore index values different from '[]byte{}' and continue with the operation.
// This should be used only to behave nicely in case you have used values different
// from '[]byte{}' in your storage before migrating to collections. Refer to
// WithKeySetUncheckedValue for more information.
func WithMultiUncheckedValue() func(*multiOptions) {
	return func(o *multiOptions) {
		o.uncheckedValue = true
	}
}

// Multi defines the most common index. It can be used to create a reference between
// a field of value and its primary key. Multiple primary keys can be mapped to the same
// reference key as the index does not enforce uniqueness constraints.
type Multi[ReferenceKey, PrimaryKey, Value any] struct {
	getRefKey func(pk PrimaryKey, value Value) (ReferenceKey, error)
	refKeys   collections.KeySet[collections.Pair[ReferenceKey, PrimaryKey]]
}

// NewMulti instantiates a new Multi instance given a schema,
// a Prefix, the humanized name for the index, the reference key key codec
// and the primary key key codec. The getRefKeyFunc is a function that
// given the primary key and value returns the referencing key.
func NewMulti[ReferenceKey, PrimaryKey, Value any](
	schema *collections.SchemaBuilder,
	prefix collections.Prefix,
	name string,
	refCodec codec.KeyCodec[ReferenceKey],
	pkCodec codec.KeyCodec[PrimaryKey],
	getRefKeyFunc func(pk PrimaryKey, value Value) (ReferenceKey, error),
	options ...func(*multiOptions),
) *Multi[ReferenceKey, PrimaryKey, Value] {
	o := new(multiOptions)
	for _, opt := range options {
		opt(o)
	}
	if o.uncheckedValue {
		return &Multi[ReferenceKey, PrimaryKey, Value]{
			getRefKey: getRefKeyFunc,
			refKeys:   collections.NewKeySet(schema, prefix, name, collections.PairKeyCodec(refCodec, pkCodec), collections.WithKeySetUncheckedValue()),
		}
	}

	return &Multi[ReferenceKey, PrimaryKey, Value]{
		getRefKey: getRefKeyFunc,
		refKeys:   collections.NewKeySet(schema, prefix, name, collections.PairKeyCodec(refCodec, pkCodec)),
	}
}

func (m *Multi[ReferenceKey, PrimaryKey, Value]) Reference(ctx context.Context, pk PrimaryKey, newValue Value, lazyOldValue func() (Value, error)) error {
	oldValue, err := lazyOldValue()
	switch {
	// if no error it means the value existed, and we need to remove the old indexes
	case err == nil:
		err = m.unreference(ctx, pk, oldValue)
		if err != nil {
			return err
		}
	// if error is ErrNotFound, it means that the object does not exist, so we're creating indexes for the first time.
	// we do nothing.
	case errors.Is(err, collections.ErrNotFound):
	// default case means that there was some other error
	default:
		return err
	}
	// create new indexes
	refKey, err := m.getRefKey(pk, newValue)
	if err != nil {
		return err
	}
	return m.refKeys.Set(ctx, collections.Join(refKey, pk))
}

func (m *Multi[ReferenceKey, PrimaryKey, Value]) Unreference(ctx context.Context, pk PrimaryKey, getValue func() (Value, error)) error {
	value, err := getValue()
	if err != nil {
		return err
	}
	return m.unreference(ctx, pk, value)
}

func (m *Multi[ReferenceKey, PrimaryKey, Value]) unreference(ctx context.Context, pk PrimaryKey, value Value) error {
	refKey, err := m.getRefKey(pk, value)
	if err != nil {
		return err
	}
	return m.refKeys.Remove(ctx, collections.Join(refKey, pk))
}

func (m *Multi[ReferenceKey, PrimaryKey, Value]) Iterate(ctx context.Context, ranger collections.Ranger[collections.Pair[ReferenceKey, PrimaryKey]]) (MultiIterator[ReferenceKey, PrimaryKey], error) {
	iter, err := m.refKeys.Iterate(ctx, ranger)
	return (MultiIterator[ReferenceKey, PrimaryKey])(iter), err
}

func (m *Multi[ReferenceKey, PrimaryKey, Value]) Walk(
	ctx context.Context,
	ranger collections.Ranger[collections.Pair[ReferenceKey, PrimaryKey]],
	walkFunc func(indexingKey ReferenceKey, indexedKey PrimaryKey) (stop bool, err error),
) error {
	return m.refKeys.Walk(ctx, ranger, func(key collections.Pair[ReferenceKey, PrimaryKey]) (bool, error) {
		return walkFunc(key.K1(), key.K2())
	})
}

// MatchExact returns a MultiIterator containing all the primary keys referenced by the provided reference key.
func (m *Multi[ReferenceKey, PrimaryKey, Value]) MatchExact(ctx context.Context, refKey ReferenceKey) (MultiIterator[ReferenceKey, PrimaryKey], error) {
	return m.Iterate(ctx, collections.NewPrefixedPairRange[ReferenceKey, PrimaryKey](refKey))
}

// RefKeys returns a list of all the MultiIterator's reference keys (may contain duplicates).
// Enable the "unique" argument to get a unique list of reference keys (the reference key must be comparable)
// WARNING: The use of RefKeys() can be very expensive in terms of Gas. Please make sure you iterate over a relatively
// small set of reference keys.
func (m *Multi[ReferenceKey, PrimaryKey, Value]) RefKeys(ctx context.Context, unique bool) ([]ReferenceKey, error) {
	iter, err := m.refKeys.IterateRaw(ctx, nil, nil, collections.OrderAscending)
	if err != nil {
		return nil, err
	}

	keys := []ReferenceKey{}
	visited := map[[32]byte]struct{}{}
	for ; iter.Valid(); iter.Next() {
		key, err := iter.Key()
		if err != nil {
			return nil, err
		}

		if unique {
			// compare the byte representation of ref keys
			refKey := key.K1()
			unsafeRefKey := *(*[]byte)(unsafe.Pointer(&refKey))

			// use SHA256 hash as map keys
			if _, ok := visited[sha256.Sum256(unsafeRefKey)]; ok {
				continue
			}
			visited[sha256.Sum256(unsafeRefKey)] = struct{}{}
		}
		keys = append(keys, key.K1())
	}

	return keys, nil
}

func (m *Multi[K1, K2, Value]) KeyCodec() codec.KeyCodec[collections.Pair[K1, K2]] {
	return m.refKeys.KeyCodec()
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
