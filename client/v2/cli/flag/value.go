package flag

import (
	"google.golang.org/protobuf/reflect/protoreflect"
)

// SimpleValue wraps a simple (non-list and non-map) protobuf value.
type SimpleValue interface {
	// Get returns the value.
	Get() protoreflect.Value
}

// ListValue wraps a protobuf list/repeating value.
type ListValue interface {
	// AppendTo appends the values to the provided list.
	AppendTo(protoreflect.List)
}
