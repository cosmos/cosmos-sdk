package collections

import (
	"context"
	"fmt"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
)

// NewMap returns a Map given a StoreKey, a Prefix and the relative value and key encoders.
func NewMap[K, V any](
	sk storetypes.StoreKey, prefix Prefix,
	keyCodec KeyCodec[K], valueCodec ValueCodec[V],
) Map[K, V] {
	return Map[K, V]{
		kc:     keyCodec,
		vc:     valueCodec,
		sk:     sk,
		prefix: prefix.Bytes(),
	}
}

// Map represents the basic collections object.
// It is used to map arbitrary keys to arbitrary
// objects.
type Map[K, V any] struct {
	kc KeyCodec[K]
	vc ValueCodec[V]

	sk     storetypes.StoreKey
	prefix []byte
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

	store, err := m.getStore(ctx)
	if err != nil {
		return err
	}
	store.Set(bytesKey, valueBytes)
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

	store, err := m.getStore(ctx)
	if err != nil {
		var v V
		return v, err
	}
	valueBytes := store.Get(bytesKey)
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
func (m Map[K, V]) Has(ctx context.Context, key K) (bool, error) {
	bytesKey, err := encodeKeyWithPrefix(m.prefix, m.kc, key)
	if err != nil {
		return false, err
	}
	store, err := m.getStore(ctx)
	if err != nil {
		return false, err
	}
	return store.Has(bytesKey), nil
}

// Remove removes the key from the storage.
// Errors with ErrEncoding if key encoding fails.
// If the key does not exist then this is a no-op.
func (m Map[K, V]) Remove(ctx context.Context, key K) error {
	bytesKey, err := encodeKeyWithPrefix(m.prefix, m.kc, key)
	if err != nil {
		return err
	}
	store, err := m.getStore(ctx)
	if err != nil {
		return err
	}
	store.Delete(bytesKey)
	return nil
}

// Iterate provides an Iterator over K and V. It accepts a Ranger interface.
// A nil ranger equals to iterate over all the keys in ascending order.
func (m Map[K, V]) Iterate(ctx context.Context, ranger Ranger[K]) (Iterator[K, V], error) {
	return iteratorFromRanger(ctx, m, ranger)
}

func (m Map[K, V]) getStore(ctx context.Context) (storetypes.KVStore, error) {
	provider, ok := ctx.(StorageProvider)
	if !ok {
		return nil, fmt.Errorf("context is not a StorageProvider: underlying type %T", ctx)
	}
	return provider.KVStore(m.sk), nil
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
