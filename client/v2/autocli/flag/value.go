package flag

import (
	"github.com/spf13/pflag"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// Value represents a single pflag.Value value.
type Value interface {
	pflag.Value
	HasValue
}

// HasValue wraps a reference to a protobuf value.
type HasValue interface {
	// Get gets the value of the protobuf value reference and returns that value
	// or an error. For composite protoreflect.Value types such as messages,
	// lists and maps, a mutable reference to the value of field obtained with
	// protoreflect.Message.NewField should be passed as the newFieldValue parameter.
	Get(newFieldValue protoreflect.Value) (protoreflect.Value, error)
}
