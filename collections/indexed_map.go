package collections

import (
	"context"
	"errors"
	"fmt"
)

// Indexes represents a type which groups multiple Index
// of one Value saved with the provided PrimaryKey.
type Indexes[PrimaryKey, Value any] interface {
	// IndexesList is implemented by the Indexes type
	// and returns all the grouped Index of Value.
	IndexesList() []Index[PrimaryKey, Value]
}

// Index represents an index of the Value
// indexed using the type PrimaryKey.
type Index[PrimaryKey, Value any] interface {
	// Reference creates a reference between the provided primary key and value.
	// If oldValue is not nil then the Index must update the references
	// of the primary key associated with the new value and remove the
	// old invalid references.
	Reference(ctx context.Context, pk PrimaryKey, newValue Value, oldValue *Value) error
	// Unreference removes the reference between the primary key and value.
	Unreference(ctx context.Context, pk PrimaryKey, value Value) error
}

// IndexedMap works like a Map but creates references
// between fields of Value and its PrimaryKey.
// These relationships are expressed and maintained using
// the Indexes type.
type IndexedMap[PrimaryKey, Value any, Idx Indexes[PrimaryKey, Value]] struct {
	Indexes Idx
	m       Map[PrimaryKey, Value]
}

func NewIndexedMap[PrimaryKey, Value any, Idx Indexes[PrimaryKey, Value]](
	schema *SchemaBuilder,
	prefix Prefix,
	name string,
	pkCodec KeyCodec[PrimaryKey],
	valueCodec ValueCodec[Value],
	indexes Idx,
) *IndexedMap[PrimaryKey, Value, Idx] {
	return &IndexedMap[PrimaryKey, Value, Idx]{
		Indexes: indexes,
		m:       NewMap(schema, prefix, name, pkCodec, valueCodec),
	}
}

func (m *IndexedMap[PrimaryKey, Value, Idx]) Get(ctx context.Context, pk PrimaryKey) (Value, error) {
	return m.m.Get(ctx, pk)
}

func (m *IndexedMap[PrimaryKey, Value, Idx]) Iterate(ctx context.Context, ranger Ranger[PrimaryKey]) (Iterator[PrimaryKey, Value], error) {
	return m.m.Iterate(ctx, ranger)
}

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

func (m *IndexedMap[PrimaryKey, Value, Idx]) Has(ctx context.Context, pk PrimaryKey) (bool, error) {
	return m.m.Has(ctx, pk)
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
