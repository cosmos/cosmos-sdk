package collections

import (
	"context"
	"cosmossdk.io/core/store"
	"fmt"
)

// Map represents the basic collections object.
// It is used to map arbitrary keys to arbitrary
// objects.
type Map[K, V any] struct {
	kc KeyCodec[K]
	vc ValueCodec[V]

	sk     func(context.Context) store.KVStore
	prefix []byte
	name   string
}

// NewMap returns a Map given a StoreKey, a Prefix, human-readable name and the relative value and key encoders.
// Name and prefix must be unique within the schema and name must match the format specified by NameRegex, or
// else this method will panic.
func NewMap[K, V any](
	schema Schema,
	prefix Prefix,
	name string,
	keyCodec KeyCodec[K],
	valueCodec ValueCodec[V],
) Map[K, V] {
	m := newMap(schema, prefix, name, keyCodec, valueCodec)
	schema.addCollection(m)
	return m
}

func newMap[K, V any](
	schema Schema, prefix Prefix, name string,
	keyCodec KeyCodec[K], valueCodec ValueCodec[V],
) Map[K, V] {
	return Map[K, V]{
		kc:     keyCodec,
		vc:     valueCodec,
		sk:     schema.storeAccessor,
		prefix: prefix.Bytes(),
		name:   name,
	}
}

func (m Map[K, V]) getName() string {
	return m.name
}

func (m Map[K, V]) getPrefix() []byte {
	return m.prefix
}

// Set maps the provided value to the provided key in the store.
// Errors with ErrEncoding if key or value encoding fails.
func (m Map[K, V]) Set(ctx context.Context, key K, value V) error {
	bytesKey, err := encodeKeyWithPrefix(m.prefix, m.kc, key)

	if err != nil {
		return err
	}

	valueBytes, err := m.vc.Encode(value)
	if err != nil {
		return fmt.Errorf("%w: value encode: %s", ErrEncoding, err) // TODO: use multi err wrapping in go1.20: https://github.com/golang/go/issues/53435
	}

	kvStore := m.sk(ctx)
	kvStore.Set(bytesKey, valueBytes)
	return nil
}

// Get returns the value associated with the provided key,
// errors with ErrNotFound if the key does not exist, or
// with ErrEncoding if the key or value decoding fails.
func (m Map[K, V]) Get(ctx context.Context, key K) (V, error) {
	bytesKey, err := encodeKeyWithPrefix(m.prefix, m.kc, key)
	if err != nil {
		var v V
		return v, err
	}

	kvStore := m.sk(ctx)
	valueBytes := kvStore.Get(bytesKey)
	if valueBytes == nil {
		var v V
		return v, fmt.Errorf("%w: key '%s' of type %s", ErrNotFound, m.kc.Stringify(key), m.vc.ValueType())
	}

	v, err := m.vc.Decode(valueBytes)
	if err != nil {
		return v, fmt.Errorf("%w: value decode: %s", ErrEncoding, err) // TODO: use multi err wrapping in go1.20: https://github.com/golang/go/issues/53435
	}
	return v, nil
}

// Has reports whether the key is present in storage or not.
// Errors with ErrEncoding if key encoding fails.
func (m Map[K, V]) Has(ctx context.Context, key K) (bool, error) {
	bytesKey, err := encodeKeyWithPrefix(m.prefix, m.kc, key)
	if err != nil {
		return false, err
	}
	kvStore := m.sk(ctx)
	return kvStore.Has(bytesKey), nil
}

// Remove removes the key from the storage.
// Errors with ErrEncoding if key encoding fails.
// If the key does not exist then this is a no-op.
func (m Map[K, V]) Remove(ctx context.Context, key K) error {
	bytesKey, err := encodeKeyWithPrefix(m.prefix, m.kc, key)
	if err != nil {
		return err
	}
	kvStore := m.sk(ctx)
	kvStore.Delete(bytesKey)
	return nil
}

// Iterate provides an Iterator over K and V. It accepts a Ranger interface.
// A nil ranger equals to iterate over all the keys in ascending order.
func (m Map[K, V]) Iterate(ctx context.Context, ranger Ranger[K]) (Iterator[K, V], error) {
	return iteratorFromRanger(ctx, m, ranger)
}

func encodeKeyWithPrefix[K any](prefix []byte, kc KeyCodec[K], key K) ([]byte, error) {
	prefixLen := len(prefix)
	// preallocate buffer
	keyBytes := make([]byte, prefixLen+kc.Size(key))
	// put prefix
	copy(keyBytes, prefix)
	// put key
	_, err := kc.Encode(keyBytes[prefixLen:], key)
	if err != nil {
		return nil, fmt.Errorf("%w: key encode: %s", ErrEncoding, err) // TODO: use multi err wrapping in go1.20: https://github.com/golang/go/issues/53435
	}
	return keyBytes, nil
}
