package collections

import (
	"bytes"
	"context"
	"fmt"
)

// KeySet builds on top of a Map and represents a collection
// retaining only a set of keys and no value.
// It can be used, for example, in an allow list.
type KeySet[K any] Map[K, noValue]

// NewKeySet returns a KeySet given a Schema, Prefix a human name for the collection and a KeyCodec for the key K.
func NewKeySet[K any](schema *SchemaBuilder, prefix Prefix, name string, keyCodec KeyCodec[K]) KeySet[K] {
	return (KeySet[K])(NewMap(schema, prefix, name, keyCodec, noValueCodec))
}

// Set adds the key to the KeySet. Errors on encoding problems.
func (k KeySet[K]) Set(ctx context.Context, key K) error {
	return (Map[K, noValue])(k).Set(ctx, key, noValue{})
}

// Has returns if the key is present in the KeySet.
// An error is returned only in case of encoding problems.
func (k KeySet[K]) Has(ctx context.Context, key K) (bool, error) {
	return (Map[K, noValue])(k).Has(ctx, key)
}

// Remove removes the key for the KeySet. An error is
// returned in case of encoding error, it won't report
// through the error if the key was removed or not.
func (k KeySet[K]) Remove(ctx context.Context, key K) error {
	return (Map[K, noValue])(k).Remove(ctx, key)
}

// Iterate iterates over the keys given the provided Ranger.
// If ranger is nil, the KeySetIterator will include all the
// existing keys within the KeySet.
func (k KeySet[K]) Iterate(ctx context.Context, ranger Ranger[K]) (KeySetIterator[K], error) {
	iter, err := (Map[K, noValue])(k).Iterate(ctx, ranger)
	if err != nil {
		return KeySetIterator[K]{}, err
	}

	return (KeySetIterator[K])(iter), nil
}

// KeySetIterator works like an Iterator, but it does not expose any API to deal with values.
type KeySetIterator[K any] Iterator[K, noValue]

func (i KeySetIterator[K]) Key() (K, error)    { return (Iterator[K, noValue])(i).Key() }
func (i KeySetIterator[K]) Keys() ([]K, error) { return (Iterator[K, noValue])(i).Keys() }
func (i KeySetIterator[K]) Next()              { (Iterator[K, noValue])(i).Next() }
func (i KeySetIterator[K]) Valid() bool        { return (Iterator[K, noValue])(i).Valid() }
func (i KeySetIterator[K]) Close() error       { return (Iterator[K, noValue])(i).Close() }

var noValueCodec ValueCodec[noValue] = noValue{}

const noValueValueType = "no_value"

type noValue struct{}

func (noValue) Encode(_ noValue) ([]byte, error) {
	return []byte{}, nil
}

func (noValue) Decode(b []byte) (noValue, error) {
	if !bytes.Equal(b, []byte{}) {
		return noValue{}, fmt.Errorf("%w: invalid value, wanted an empty non-nil byte slice", ErrEncoding)
	}
	return noValue{}, nil
}

func (noValue) Stringify(_ noValue) string {
	return noValueValueType
}

func (n noValue) ValueType() string {
	return noValueValueType
}
