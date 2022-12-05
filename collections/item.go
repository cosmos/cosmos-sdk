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
// If value decoding fails then an ErrEncoding is returned.
func (i Item[V]) Get(ctx StorageProvider) (V, error) { return (Map[noKey, V])(i).Get(ctx, noKey{}) }

// Set sets the item in the store. If Value encoding fails then an ErrEncoding is returned.
func (i Item[V]) Set(ctx StorageProvider, value V) error {
	return (Map[noKey, V])(i).Set(ctx, noKey{}, value)
}

// Has reports whether the item exists in the store or not.
// Returns an error in case
func (i Item[V]) Has(ctx StorageProvider) (bool, error) { return (Map[noKey, V])(i).Has(ctx, noKey{}) }

// Remove removes the item in the store.
func (i Item[V]) Remove(ctx StorageProvider) error { return (Map[noKey, V])(i).Remove(ctx, noKey{}) }

// noKey defines a KeyEncoder which decodes nothing.
type noKey struct{}

func (noKey) Stringify(_ noKey) string              { return "no_key" }
func (noKey) KeyType() string                       { return "no_key" }
func (noKey) Size(_ noKey) int                      { return 0 }
func (noKey) PutKey(_ []byte, _ noKey) (int, error) { return 0, nil }
func (n noKey) ReadKey(buffer []byte) (int, noKey, error) {
	if len(buffer) != 0 {
		return 0, noKey{}, fmt.Errorf("%w: must be empty for %s", errDecodeKeySize, n.KeyType())
	}
	return 0, noKey{}, nil
}
