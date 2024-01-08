package collections

import (
	"context"
	"fmt"

	"cosmossdk.io/collections/codec"
	"cosmossdk.io/core/store"
)

// LookupMap represents a map that is not iterable.
type LookupMap[K, V any] struct {
	kc codec.KeyCodec[K]
	vc codec.ValueCodec[V]

	// store accessor
	sa     func(context.Context) store.KVStore
	prefix []byte
	name   string
}

// NewLookupMap creates a new LookupMap.
func NewLookupMap[K, V any](
	schemaBuilder *SchemaBuilder,
	prefix Prefix,
	name string,
	keyCodec codec.KeyCodec[K],
	valueCodec codec.ValueCodec[V],
) LookupMap[K, V] {
	m := LookupMap[K, V]{
		kc:     keyCodec,
		vc:     valueCodec,
		sa:     schemaBuilder.schema.storeAccessor,
		prefix: prefix.Bytes(),
		name:   name,
	}
	schemaBuilder.addCollection(lookupMapImpl[K, V]{m})
	return m
}

func (m LookupMap[K, V]) GetName() string {
	return m.name
}

func (m LookupMap[K, V]) GetPrefix() []byte {
	return m.prefix
}

// Set maps the provided value to the provided key in the store.
// Errors with ErrEncoding if key or value encoding fails.
func (m LookupMap[K, V]) Set(ctx context.Context, key K, value V) error {
	bytesKey, err := EncodeKeyWithPrefix(m.prefix, m.kc, key)
	if err != nil {
		return err
	}

	valueBytes, err := m.vc.Encode(value)
	if err != nil {
		return fmt.Errorf("%w: value encode: %s", ErrEncoding, err) // TODO: use multi err wrapping in go1.20: https://github.com/golang/go/issues/53435
	}

	kvStore := m.sa(ctx)
	return kvStore.Set(bytesKey, valueBytes)
}

// Get returns the value associated with the provided key,
// errors with ErrNotFound if the key does not exist, or
// with ErrEncoding if the key or value decoding fails.
func (m LookupMap[K, V]) Get(ctx context.Context, key K) (v V, err error) {
	bytesKey, err := EncodeKeyWithPrefix(m.prefix, m.kc, key)
	if err != nil {
		return v, err
	}

	kvStore := m.sa(ctx)
	valueBytes, err := kvStore.Get(bytesKey)
	if err != nil {
		return v, err
	}
	if valueBytes == nil {
		return v, fmt.Errorf("%w: key '%s' of type %s", ErrNotFound, m.kc.Stringify(key), m.vc.ValueType())
	}

	v, err = m.vc.Decode(valueBytes)
	if err != nil {
		return v, fmt.Errorf("%w: value decode: %s", ErrEncoding, err) // TODO: use multi err wrapping in go1.20: https://github.com/golang/go/issues/53435
	}
	return v, nil
}

// Has reports whether the key is present in storage or not.
// Errors with ErrEncoding if key encoding fails.
func (m LookupMap[K, V]) Has(ctx context.Context, key K) (bool, error) {
	bytesKey, err := EncodeKeyWithPrefix(m.prefix, m.kc, key)
	if err != nil {
		return false, err
	}
	kvStore := m.sa(ctx)
	return kvStore.Has(bytesKey)
}

// Remove removes the key from the storage.
// Errors with ErrEncoding if key encoding fails.
// If the key does not exist then this is a no-op.
func (m LookupMap[K, V]) Remove(ctx context.Context, key K) error {
	bytesKey, err := EncodeKeyWithPrefix(m.prefix, m.kc, key)
	if err != nil {
		return err
	}
	kvStore := m.sa(ctx)
	return kvStore.Delete(bytesKey)
}

// KeyCodec returns the Map's KeyCodec.
func (m LookupMap[K, V]) KeyCodec() codec.KeyCodec[K] { return m.kc }

// ValueCodec returns the Map's ValueCodec.
func (m LookupMap[K, V]) ValueCodec() codec.ValueCodec[V] { return m.vc }

// // ProveExistence proves the existence of a key in the collection.
// func (m LookupMap[K, V]) ProveExistence(ctx context.Context, key K) error {
// 	// Implementation...
// }

// // ProveNonExistence proves the non-existence of a key in the collection.
// func (m LookupMap[K, V]) ProveNonExistence(ctx context.Context, key K) error {
// 	// Implementation...
// }
