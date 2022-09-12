package flag

import (
	"github.com/spf13/pflag"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type Value interface {
	pflag.Value

	// Get returns
	Get(protoreflect.Value) (protoreflect.Value, error)
	FieldValueBinder
}

// SimpleValue wraps a simple (non-list and non-map) protobuf value.
type SimpleValue interface {
	// Get returns the value.
}

// ListValue wraps a protobuf list/repeating value.
type ListValue interface {
	// AppendTo appends the values to the provided list.
	AppendTo(protoreflect.List)
}
