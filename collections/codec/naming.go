package codec

import "fmt"

// NameableKeyCodec is a KeyCodec that can be named.
type NameableKeyCodec[T any] interface {
	KeyCodec[T]

	// WithName returns the KeyCodec with the provided name.
	WithName(name string) KeyCodec[T]
}

// NameableValueCodec is a ValueCodec that can be named.
type NameableValueCodec[T any] interface {
	ValueCodec[T]

	// WithName returns the ValueCodec with the provided name.
	WithName(name string) ValueCodec[T]
}

type NamedKeyCodec[T any] struct {
	KeyCodec[T]
	Name string
}

func (n NamedKeyCodec[T]) SchemaCodec() (SchemaCodec[T], error) {
	cdc, err := KeySchemaCodec[T](n.KeyCodec)
	if err != nil {
		return SchemaCodec[T]{}, err
	}
	return withName(cdc, n.Name)
}

type NamedValueCodec[T any] struct {
	ValueCodec[T]
	Name string
}

func (n NamedValueCodec[T]) SchemaCodec() (SchemaCodec[T], error) {
	cdc, err := ValueSchemaCodec[T](n.ValueCodec)
	if err != nil {
		return SchemaCodec[T]{}, err
	}
	return withName(cdc, n.Name)
}

func withName[T any](cdc SchemaCodec[T], name string) (SchemaCodec[T], error) {
	if len(cdc.Fields) != 1 {
		return SchemaCodec[T]{}, fmt.Errorf("expected exactly one field to be named, got %d", len(cdc.Fields))
	}
	cdc.Fields[0].Name = name
	return cdc, nil
}
