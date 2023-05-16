package collections

import (
	"bytes"
	"context"
	"fmt"

	"cosmossdk.io/collections/codec"
)

// KeySet builds on top of a Map and represents a collection retaining only a set
// of keys and no value. It can be used, for example, in an allow list.
type KeySet[K any] Map[K, NoValue]

// NewKeySet returns a KeySet given a Schema, Prefix a human name for the collection
// and a KeyCodec for the key K.
func NewKeySet[K any](schema *SchemaBuilder, prefix Prefix, name string, keyCodec codec.KeyCodec[K]) KeySet[K] {
	return (KeySet[K])(NewMap(schema, prefix, name, keyCodec, noValueCodec))
}

// Set adds the key to the KeySet. Errors on encoding problems.
func (k KeySet[K]) Set(ctx context.Context, key K) error {
	return (Map[K, NoValue])(k).Set(ctx, key, NoValue{})
}

// Has returns if the key is present in the KeySet.
// An error is returned only in case of encoding problems.
func (k KeySet[K]) Has(ctx context.Context, key K) (bool, error) {
	return (Map[K, NoValue])(k).Has(ctx, key)
}

// Remove removes the key for the KeySet. An error is returned in case of
// encoding error, it won't report through the error if the key was
// removed or not.
func (k KeySet[K]) Remove(ctx context.Context, key K) error {
	return (Map[K, NoValue])(k).Remove(ctx, key)
}

// Iterate iterates over the keys given the provided Ranger. If ranger is nil,
// the KeySetIterator will include all the existing keys within the KeySet.
func (k KeySet[K]) Iterate(ctx context.Context, ranger Ranger[K]) (KeySetIterator[K], error) {
	iter, err := (Map[K, NoValue])(k).Iterate(ctx, ranger)
	if err != nil {
		return KeySetIterator[K]{}, err
	}

	return (KeySetIterator[K])(iter), nil
}

func (k KeySet[K]) IterateRaw(ctx context.Context, start, end []byte, order Order) (Iterator[K, NoValue], error) {
	return (Map[K, NoValue])(k).IterateRaw(ctx, start, end, order)
}

// Walk provides the same functionality as Map.Walk, but callbacks the walk
// function only with the key.
func (k KeySet[K]) Walk(ctx context.Context, ranger Ranger[K], walkFunc func(key K) (stop bool, err error)) error {
	return (Map[K, NoValue])(k).Walk(ctx, ranger, func(key K, value NoValue) (bool, error) { return walkFunc(key) })
}

func (k KeySet[K]) KeyCodec() codec.KeyCodec[K]           { return (Map[K, NoValue])(k).KeyCodec() }
func (k KeySet[K]) ValueCodec() codec.ValueCodec[NoValue] { return (Map[K, NoValue])(k).ValueCodec() }

// KeySetIterator works like an Iterator, but it does not expose any API to deal with values.
type KeySetIterator[K any] Iterator[K, NoValue]

func (i KeySetIterator[K]) Key() (K, error)    { return (Iterator[K, NoValue])(i).Key() }
func (i KeySetIterator[K]) Keys() ([]K, error) { return (Iterator[K, NoValue])(i).Keys() }
func (i KeySetIterator[K]) Next()              { (Iterator[K, NoValue])(i).Next() }
func (i KeySetIterator[K]) Valid() bool        { return (Iterator[K, NoValue])(i).Valid() }
func (i KeySetIterator[K]) Close() error       { return (Iterator[K, NoValue])(i).Close() }

var noValueCodec codec.ValueCodec[NoValue] = NoValue{}

const noValueValueType = "no_value"

type NoValue struct{}

func (n NoValue) EncodeJSON(_ NoValue) ([]byte, error) {
	return nil, nil
}

func (n NoValue) DecodeJSON(b []byte) (NoValue, error) {
	if b != nil {
		return NoValue{}, fmt.Errorf("%w: expected nil json bytes, got: %x", ErrEncoding, b)
	}
	return NoValue{}, nil
}

func (NoValue) Encode(_ NoValue) ([]byte, error) {
	return []byte{}, nil
}

func (NoValue) Decode(b []byte) (NoValue, error) {
	if !bytes.Equal(b, []byte{}) {
		return NoValue{}, fmt.Errorf("%w: invalid value, wanted an empty non-nil byte slice", ErrEncoding)
	}
	return NoValue{}, nil
}

func (NoValue) Stringify(_ NoValue) string {
	return noValueValueType
}

func (n NoValue) ValueType() string {
	return noValueValueType
}
