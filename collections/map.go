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
// Errors with ErrEncoding if key or value encoding fails.
func (m Map[K, V]) Set(ctx StorageProvider, key K, value V) error {
	keyBytes, err := m.encodeKey(key)
	if err != nil {
		return err
	}

	valueBytes, err := m.vc.Encode(value)
	if err != nil {
		return fmt.Errorf("%w: value encode: %s", ErrEncoding, err) // TODO: use multi err wrapping in go1.20: https://github.com/golang/go/issues/53435
	}

	m.getStore(ctx).Set(keyBytes, valueBytes)
	return nil
}

// Get returns the value associated with the provided key,
// errors with ErrNotFound if the key does not exist, or
// with ErrEncoding if the key or value decoding fails.
func (m Map[K, V]) Get(ctx StorageProvider, key K) (V, error) {
	keyBytes, err := m.encodeKey(key)
	if err != nil {
		var v V
		return v, err
	}

	valueBytes := m.getStore(ctx).Get(keyBytes)
	if valueBytes == nil {
		var v V
		return v, fmt.Errorf("%w: key '%s' of type %s", ErrNotFound, m.kc.Stringify(key), m.vc.ValueType())
	}

	v, err := m.vc.Decode(valueBytes)
	if err != nil {
		var v V
		return v, fmt.Errorf("%w: value decode: %s", ErrEncoding, err) // TODO: use multi err wrapping in go1.20: https://github.com/golang/go/issues/53435
	}
	return v, nil
}

// Has reports whether the key is present in storage or not.
// Errors with ErrEncoding if key encoding fails.
func (m Map[K, V]) Has(ctx StorageProvider, key K) (bool, error) {
	bytesKey, err := m.encodeKey(key)
	if err != nil {
		return false, err
	}
	return m.getStore(ctx).Has(bytesKey), nil
}

// Remove removes the key from the storage.
// Errors with ErrEncoding if key encoding fails.
// If the key does not exist then this is a no-op.
func (m Map[K, V]) Remove(ctx StorageProvider, key K) error {
	bytesKey, err := m.encodeKey(key)
	if err != nil {
		return err
	}
	m.getStore(ctx).Delete(bytesKey)
	return nil
}

func (m Map[K, V]) getStore(provider StorageProvider) storetypes.KVStore {
	return provider.KVStore(m.sk)
}

func (m Map[K, V]) encodeKey(key K) ([]byte, error) {
	keyBytes := make([]byte, m.kc.Size(key))
	_, err := m.kc.PutKey(keyBytes, key)
	if err != nil {
		return nil, fmt.Errorf("%w: key encode: %s", ErrEncoding, err) // TODO: use multi err wrapping in go1.20: https://github.com/golang/go/issues/53435
	}
	return append(m.prefix, keyBytes...), nil
}
