package collections

import (
	"context"
	"errors"
	"fmt"

	"cosmossdk.io/collections/codec"
)

// Indexes represents a type which groups multiple Index
// of one Value saved with the provided PrimaryKey.
// Indexes is just meant to be a struct containing all
// the indexes to maintain relationship for.
type Indexes[PrimaryKey, Value any] interface {
	// IndexesList is implemented by the Indexes type
	// and returns all the grouped Index of Value.
	IndexesList() []Index[PrimaryKey, Value]
}

// Index represents an index of the Value indexed using the type PrimaryKey.
type Index[PrimaryKey, Value any] interface {
	// Reference creates a reference between the provided primary key and value.
	// If oldValue is not nil then the Index must update the references
	// of the primary key associated with the new value and remove the
	// old invalid references.
	Reference(ctx context.Context, pk PrimaryKey, newValue Value, oldValue *Value) error
	// Unreference removes the reference between the primary key and value.
	Unreference(ctx context.Context, pk PrimaryKey, value Value) error
}

// IndexedMap works like a Map but creates references between fields of Value and its PrimaryKey.
// These relationships are expressed and maintained using the Indexes type.
// Internally IndexedMap can be seen as a partitioned collection, one partition
// is a Map[PrimaryKey, Value], that maintains the object, the second
// are the Indexes.
type IndexedMap[PrimaryKey, Value any, Idx Indexes[PrimaryKey, Value]] struct {
	Indexes Idx
	m       Map[PrimaryKey, Value]
}

// NewIndexedMap instantiates a new IndexedMap. Accepts a SchemaBuilder, a Prefix,
// a humanised name that defines the name of the collection, the primary key codec
// which is basically what IndexedMap uses to encode the primary key to bytes,
// the value codec which is what the IndexedMap uses to encode the value.
// Then it expects the initialised indexes.
func NewIndexedMap[PrimaryKey, Value any, Idx Indexes[PrimaryKey, Value]](
	schema *SchemaBuilder,
	prefix Prefix,
	name string,
	pkCodec codec.KeyCodec[PrimaryKey],
	valueCodec codec.ValueCodec[Value],
	indexes Idx,
) *IndexedMap[PrimaryKey, Value, Idx] {
	return &IndexedMap[PrimaryKey, Value, Idx]{
		Indexes: indexes,
		m:       NewMap(schema, prefix, name, pkCodec, valueCodec),
	}
}

// Get gets the object given its primary key.
func (m *IndexedMap[PrimaryKey, Value, Idx]) Get(ctx context.Context, pk PrimaryKey) (Value, error) {
	return m.m.Get(ctx, pk)
}

// Iterate allows to iterate over the objects given a Ranger of the primary key.
func (m *IndexedMap[PrimaryKey, Value, Idx]) Iterate(ctx context.Context, ranger Ranger[PrimaryKey]) (Iterator[PrimaryKey, Value], error) {
	return m.m.Iterate(ctx, ranger)
}

// Has reports if exists a value with the provided primary key.
func (m *IndexedMap[PrimaryKey, Value, Idx]) Has(ctx context.Context, pk PrimaryKey) (bool, error) {
	return m.m.Has(ctx, pk)
}

// Set maps the value using the primary key. It will also iterate every index and instruct them to
// add or update the indexes.
func (m *IndexedMap[PrimaryKey, Value, Idx]) Set(ctx context.Context, pk PrimaryKey, value Value) error {
	// we need to see if there was a previous instance of the value
	oldValue, err := m.m.Get(ctx, pk)
	switch {
	// update indexes
	case err == nil:
		err = m.ref(ctx, pk, value, &oldValue)
		if err != nil {
			return fmt.Errorf("collections: indexing error: %w", err)
		}
	// create new indexes
	case errors.Is(err, ErrNotFound):
		err = m.ref(ctx, pk, value, nil)
		if err != nil {
			return fmt.Errorf("collections: indexing error: %w", err)
		}
	// cannot move forward error
	default:
		return err
	}

	return m.m.Set(ctx, pk, value)
}

// Remove removes the value associated with the primary key from the map. Then
// it iterates over all the indexes and instructs them to remove all the references
// associated with the removed value.
func (m *IndexedMap[PrimaryKey, Value, Idx]) Remove(ctx context.Context, pk PrimaryKey) error {
	oldValue, err := m.m.Get(ctx, pk)
	if err != nil {
		// TODO retain Map behaviour? which does not error in case we remove a non-existing object
		return err
	}

	err = m.unref(ctx, pk, oldValue)
	if err != nil {
		return err
	}
	return m.m.Remove(ctx, pk)
}

// Walk applies the same semantics as Map.Walk.
func (m *IndexedMap[PrimaryKey, Value, Idx]) Walk(ctx context.Context, ranger Ranger[PrimaryKey], walkFunc func(key PrimaryKey, value Value) bool) error {
	return m.m.Walk(ctx, ranger, walkFunc)
}

// IterateRaw iterates the IndexedMap using raw bytes keys. Follows the same semantics as Map.IterateRaw
func (m *IndexedMap[PrimaryKey, Value, Idx]) IterateRaw(ctx context.Context, start, end []byte, order Order) (Iterator[PrimaryKey, Value], error) {
	return m.m.IterateRaw(ctx, start, end, order)
}

func (m *IndexedMap[PrimaryKey, Value, Idx]) KeyCodec() codec.KeyCodec[PrimaryKey] {
	return m.m.KeyCodec()
}

func (m *IndexedMap[PrimaryKey, Value, Idx]) ValueCodec() codec.ValueCodec[Value] {
	return m.m.ValueCodec()
}

func (m *IndexedMap[PrimaryKey, Value, Idx]) ref(ctx context.Context, pk PrimaryKey, value Value, oldValue *Value) error {
	for _, index := range m.Indexes.IndexesList() {
		err := index.Reference(ctx, pk, value, oldValue)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *IndexedMap[PrimaryKey, Value, Idx]) unref(ctx context.Context, pk PrimaryKey, value Value) error {
	for _, index := range m.Indexes.IndexesList() {
		err := index.Unreference(ctx, pk, value)
		if err != nil {
			return err
		}
	}
	return nil
}
