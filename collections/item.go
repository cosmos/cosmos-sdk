package collections

import (
	"context"
	io "io"
)

// Item is a type declaration based on Map
// with a non-existent key.
type Item[V any] Map[noKey, V]

// NewItem instantiates a new Item instance, given the value encoder of the item V.
// Name and prefix must be unique within the schema and name must match the format specified by NameRegex, or
// else this method will panic.
func NewItem[V any](
	schema Schema,
	prefix Prefix,
	name string,
	valueCodec ValueCodec[V],
) Item[V] {
	item := (Item[V])(newMap[noKey, V](schema, prefix, name, noKey{}, valueCodec))
	schema.addCollection(item)
	return item
}

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

func (i Item[V]) defaultGenesis(writer io.Writer) error {
	var value V
	bz, err := i.vc.EncodeJSON(value)
	_, err = writer.Write(bz)
	return err
}

func (i Item[V]) validateGenesis(reader io.Reader) error {
	bz, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	_, err = i.vc.DecodeJSON(bz)
	return err
}

func (i Item[V]) importGenesis(ctx context.Context, reader io.Reader) error {
	bz, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	value, err := i.vc.DecodeJSON(bz)

	return i.Set(ctx, value)
}

func (i Item[V]) exportGenesis(ctx context.Context, writer io.Writer) error {
	value, err := i.Get(ctx)
	if err != nil {
		return err
	}

	bz, err := i.vc.EncodeJSON(value)
	if err != nil {
		return err
	}

	_, err = writer.Write(bz)
	return err
}

// noKey defines a KeyCodec which decodes nothing.
type noKey struct{}

func (noKey) Stringify(_ noKey) string              { return "no_key" }
func (noKey) KeyType() string                       { return "no_key" }
func (noKey) Size(_ noKey) int                      { return 0 }
func (noKey) Encode(_ []byte, _ noKey) (int, error) { return 0, nil }
func (noKey) Decode(_ []byte) (int, noKey, error)   { return 0, noKey{}, nil }
func (noKey) EncodeJSON(_ noKey) ([]byte, error)    { return []byte("null"), nil }
func (noKey) DecodeJSON(_ []byte) (noKey, error)    { return noKey{}, nil }
