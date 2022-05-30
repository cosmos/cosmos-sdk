package textual

import (
	"context"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// ValueRenderer defines an interface to produce formatted output for all
// protobuf types as well as parse a string into those protobuf types.
//
// The notion of "value renderer" is defined in ADR-050, and that ADR provides
// a default spec for value renderers. However, we define an interface here, so
// that optionally more value renderers could be built, for example, a
// separate one for a different language.
type ValueRenderer interface {
	Format(context.Context, protoreflect.FieldDescriptor, protoreflect.Value) ([]string, error)
	Parse(context.Context, []string) (proto.Message, error)
}
