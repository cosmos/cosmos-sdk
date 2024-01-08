package collections

import (
	"context"

	"cosmossdk.io/collections/codec"
)

// LookupMap represents a map that is not iterable.
type LookupMap[K any] Map[K, NoValue]

// NewLookupMap creates a new LookupMap.
func NewLookupMap[K any](
	schemaBuilder *SchemaBuilder,
	prefix Prefix,
	name string,
	keyCodec codec.KeyCodec[K],
) LookupMap[K] {
	m := LookupMap[K](NewMap[K](schemaBuilder, prefix, name, keyCodec, noValueCodec))
	return m
}

// GetName returns the name of the collection.
func (m LookupMap[K]) GetName() string {
	return m.name
}

// GetPrefix returns the prefix of the collection.
func (m LookupMap[K]) GetPrefix() []byte {
	return m.prefix
}

// Set adds the key to the LookupMap.
// Errors on encoding problems.
func (m LookupMap[K]) Set(ctx context.Context, key K) error {
	// return m.(LookupMap[K]).Set(ctx, key)
	return (Map[K, NoValue])(m).Set(ctx, key, NoValue{})
}

// Has returns if the key is present in the LookupMap.
// An error is returned only in case of encoding problems.
func (m LookupMap[K]) Has(ctx context.Context, key K) (bool, error) {
	return (Map[K, NoValue])(m).Has(ctx, key)
}

// Remove removes the key for the LookupMap. An error is returned in case of
// encoding error, it won't report through the error if the key was
// removed or not.
func (m LookupMap[K]) Remove(ctx context.Context, key K) error {
	return (Map[K, NoValue])(m).Remove(ctx, key)
}

// KeyCodec returns the Map's KeyCodec.
func (m LookupMap[K]) KeyCodec() codec.KeyCodec[K] { return (Map[K, NoValue])(m).KeyCodec() }

// ValueCodec returns the Map's ValueCodec.
func (m LookupMap[K]) ValueCodec() codec.ValueCodec[NoValue] {
	return (Map[K, NoValue])(m).ValueCodec()
}
