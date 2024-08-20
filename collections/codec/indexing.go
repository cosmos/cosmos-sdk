package codec

import (
	"encoding/json"
	"fmt"

	"cosmossdk.io/schema"
)

// HasSchemaCodec is an interface that all codec's should implement in order
// to properly support indexing. It is not required by KeyCodec or ValueCodec
// in order to preserve backwards compatibility, but a future version of collections
// may make it required and all codec's should aim to implement it. If it is not
// implemented, fallback defaults will be used for indexing that may be sub-optimal.
//
// Implementations of HasSchemaCodec should test that they are conformant using
// schema.ValidateObjectKey or schema.ValidateObjectValue depending on whether
// the codec is a KeyCodec or ValueCodec respectively.
type HasSchemaCodec[T any] interface {
	// SchemaCodec returns the schema codec for the collections codec.
	SchemaCodec() (SchemaCodec[T], error)
}

// SchemaCodec is a codec that supports converting collection codec values to and
// from schema codec values.
type SchemaCodec[T any] struct {
	// Fields are the schema fields that the codec represents. If this is empty,
	// it will be assumed that this codec represents no value (such as an item key
	// or key set value).
	Fields []schema.Field

	// ToSchemaType converts a codec value of type T to a value corresponding to
	// a schema object key or value (depending on whether this is a key or value
	// codec). The returned value should pass validation with schema.ValidateObjectKey
	// or schema.ValidateObjectValue with the fields specified in Fields.
	// If this function is nil, it will be assumed that T already represents a
	// value that conforms to a schema value without any further conversion.
	ToSchemaType func(T) (any, error)

	// FromSchemaType converts a schema object key or value to T.
	// If this function is nil, it will be assumed that T already represents a
	// value that conforms to a schema value without any further conversion.
	FromSchemaType func(any) (T, error)
}

// KeySchemaCodec gets the schema codec for the provided KeyCodec either
// by casting to HasSchemaCodec or returning a fallback codec.
func KeySchemaCodec[K any](cdc KeyCodec[K]) (SchemaCodec[K], error) {
	if indexable, ok := cdc.(HasSchemaCodec[K]); ok {
		return indexable.SchemaCodec()
	} else {
		return FallbackSchemaCodec[K](), nil
	}
}

// ValueSchemaCodec gets the schema codec for the provided ValueCodec either
// by casting to HasSchemaCodec or returning a fallback codec.
func ValueSchemaCodec[V any](cdc ValueCodec[V]) (SchemaCodec[V], error) {
	if indexable, ok := cdc.(HasSchemaCodec[V]); ok {
		return indexable.SchemaCodec()
	} else {
		return FallbackSchemaCodec[V](), nil
	}
}

// FallbackSchemaCodec returns a fallback schema codec for T when one isn't explicitly
// specified with HasSchemaCodec. It maps all simple types directly to schema kinds
// and converts everything else to JSON.
func FallbackSchemaCodec[T any]() SchemaCodec[T] {
	var t T
	kind := schema.KindForGoValue(t)
	if err := kind.Validate(); err == nil {
		return SchemaCodec[T]{
			Fields: []schema.Field{{
				// we don't set any name so that this can be set to a good default by the caller
				Name: "",
				Kind: kind,
			}},
			// these can be nil because T maps directly to a schema value for this kind
			ToSchemaType:   nil,
			FromSchemaType: nil,
		}
	} else {
		// we default to encoding everything to JSON
		return SchemaCodec[T]{
			Fields: []schema.Field{{Kind: schema.JSONKind}},
			ToSchemaType: func(t T) (any, error) {
				bz, err := json.Marshal(t)
				return json.RawMessage(bz), err
			},
			FromSchemaType: func(a any) (T, error) {
				var t T
				bz, ok := a.(json.RawMessage)
				if !ok {
					return t, fmt.Errorf("expected json.RawMessage, got %T", a)
				}
				err := json.Unmarshal(bz, &t)
				return t, err
			},
		}
	}
}
