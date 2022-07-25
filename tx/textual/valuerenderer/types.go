package valuerenderer

import (
	"context"
	"io"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// ValueRenderer defines an interface to produce formatted output for all
// protobuf types as well as parse a string into those protobuf types.
//
// The notion of "value renderer" is defined in ADR-050, and that ADR provides
// a default spec for value renderers. However, we define it as an interface
// here, so that optionally more value renderers could be built, for example, a
// separate one for a different language.
type ValueRenderer interface {
	Format(context.Context, protoreflect.Value, io.Writer) error
	Parse(context.Context, io.Reader) (protoreflect.Value, error)
}
