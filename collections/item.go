package collections

import (
	"fmt"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
)

// NewItem instantiates a new Item instance, given the value encoder of the item V.
func NewItem[V any](sk storetypes.StoreKey, prefix Prefix, valueEncoder ValueEncoder[V]) Item[V] {
	return (Item[V])(NewMap[noKey, V](sk, prefix, noKey{}, valueEncoder))
}

// Item is a type declaration based on Map
// with a non-existent key.
type Item[V any] Map[noKey, V]

// Get gets the item, if it is not set it returns an ErrNotFound error.
func (i Item[V]) Get(ctx StorageProvider) (V, error) { return (Map[noKey, V])(i).Get(ctx, noKey{}) }

// GetOr gets the item if it is present in the store.
// If it is not found, it returns the provided default value.
func (i Item[V]) GetOr(ctx StorageProvider, defaultValue V) V {
	return (Map[noKey, V])(i).GetOr(ctx, noKey{}, defaultValue)
}

// Set sets the item in the store.
func (i Item[V]) Set(ctx StorageProvider, value V) { (Map[noKey, V])(i).Set(ctx, noKey{}, value) }

// Has reports whether the item exists in the store or not.
func (i Item[V]) Has(ctx StorageProvider) bool { return (Map[noKey, V])(i).Has(ctx, noKey{}) }

// Remove removes the item in the store.
func (i Item[V]) Remove(ctx StorageProvider) { (Map[noKey, V])(i).Remove(ctx, noKey{}) }

// noKey defines a KeyEncoder which decodes nothing.
type noKey struct{}

func (n noKey) Stringify(_ noKey) string       { return "no_key" }
func (n noKey) KeyType() string                { return "no_key" }
func (n noKey) Encode(_ noKey) ([]byte, error) { return []byte{}, nil }

func (n noKey) Decode(b []byte) (int, noKey, error) {
	if len(b) != 0 {
		return 0, noKey{}, fmt.Errorf("%w: must be empty for %s", errDecodeKeySize, n.KeyType())
	}
	return 0, noKey{}, nil
}
