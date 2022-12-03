package collections

import (
	"fmt"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
)

// NewMap returns a Map given a StoreKey, a Prefix and the relative value and key encoders.
func NewMap[K, V any](
	sk storetypes.StoreKey, prefix Prefix,
	keyEncoder KeyEncoder[K], valueEncoder ValueEncoder[V],
) Map[K, V] {
	return Map[K, V]{
		kc:     keyEncoder,
		vc:     valueEncoder,
		sk:     sk,
		prefix: prefix.Bytes(),
	}
}

// Map represents the basic collections object.
// It is used to map arbitrary keys to arbitrary
// objects.
type Map[K, V any] struct {
	kc KeyEncoder[K]
	vc ValueEncoder[V]

	sk     storetypes.StoreKey
	prefix []byte
}

// Set maps the provided value to the provided key in the store.
func (m Map[K, V]) Set(ctx StorageProvider, key K, value V) {
	keyBytes := m.encodeKey(key)

	valueBytes, err := m.vc.Encode(value)
	if err != nil {
		panic(err)
	}

	m.getStore(ctx).Set(keyBytes, valueBytes)
}

// Get returns the value associated with the provided key,
// or a ErrNotFound error in case the key is not present.
func (m Map[K, V]) Get(ctx StorageProvider, key K) (V, error) {
	keyBytes := m.encodeKey(key)

	valueBytes := m.getStore(ctx).Get(keyBytes)
	if valueBytes == nil {
		var v V
		return v, fmt.Errorf("%w: key '%s' of type %s", ErrNotFound, m.kc.Stringify(key), m.vc.ValueType())
	}

	v, err := m.vc.Decode(valueBytes)
	if err != nil {
		panic(err)
	}
	return v, nil
}

// GetOr returns the value associated with the provided key.
// If the key is not present in store, then we return the provided
// default value.
func (m Map[K, V]) GetOr(ctx StorageProvider, key K, defaultValue V) V {
	v, err := m.Get(ctx, key)
	if err != nil {
		return defaultValue
	}
	return v
}

// Has reports whether the key is present in storage or not.
func (m Map[K, V]) Has(ctx StorageProvider, key K) bool {
	return m.getStore(ctx).Has(m.encodeKey(key))
}

// Remove removes the key from the storage.
// If the key does not exist then this is a no-op.
func (m Map[K, V]) Remove(ctx StorageProvider, key K) {
	m.getStore(ctx).Delete(m.encodeKey(key))
}

func (m Map[K, V]) getStore(provider StorageProvider) storetypes.KVStore {
	return provider.KVStore(m.sk)
}

func (m Map[K, V]) encodeKey(key K) []byte {
	bytes, err := m.kc.Encode(key)
	if err != nil {
		panic(err)
	}
	return append(m.prefix, bytes...)
}
