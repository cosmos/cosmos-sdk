package collections

import (
	"context"
)

// NewItem instantiates a new Item instance, given the value encoder of the item V.
// Name and prefix must be unique within the schema and name must match the format specified by NameRegex, or
// else this method will panic.
func NewItem[V any](schema Schema, prefix Prefix, name string, valueCodec ValueCodec[V]) Item[V] {
	item := (Item[V])(newMap[noKey, V](schema, prefix, name, noKey{}, valueCodec))
	schema.addCollection(item)
	return item
}

// Item is a type declaration based on Map
// with a non-existent key.
type Item[V any] Map[noKey, V]

func (i Item[V]) getName() string {
	return i.name
}

func (i Item[V]) getPrefix() []byte {
	return i.prefix
}

// Get gets the item, if it is not set it returns an ErrNotFound error.
// If value decoding fails then an ErrEncoding is returned.
func (i Item[V]) Get(ctx context.Context) (V, error) {
	return (Map[noKey, V])(i).Get(ctx, noKey{})
}

// Set sets the item in the store. If Value encoding fails then an ErrEncoding is returned.
func (i Item[V]) Set(ctx context.Context, value V) error {
	return (Map[noKey, V])(i).Set(ctx, noKey{}, value)
}

// Has reports whether the item exists in the store or not.
// Returns an error in case
func (i Item[V]) Has(ctx context.Context) (bool, error) {
	return (Map[noKey, V])(i).Has(ctx, noKey{})
}

// Remove removes the item in the store.
func (i Item[V]) Remove(ctx context.Context) error {
	return (Map[noKey, V])(i).Remove(ctx, noKey{})
}

// noKey defines a KeyCodec which decodes nothing.
type noKey struct{}

func (noKey) Stringify(_ noKey) string              { return "no_key" }
func (noKey) KeyType() string                       { return "no_key" }
func (noKey) Size(_ noKey) int                      { return 0 }
func (noKey) Encode(_ []byte, _ noKey) (int, error) { return 0, nil }
func (noKey) Decode(_ []byte) (int, noKey, error)   { return 0, noKey{}, nil }
