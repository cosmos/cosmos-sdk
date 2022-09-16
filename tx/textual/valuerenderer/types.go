package valuerenderer

import (
	"context"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// Item is the abstract unit of Textual rendering.
type Item struct {
	Text   string
	Indent int
	Expert bool
}

// ValueRenderer defines an interface to produce formatted output for all
// protobuf types as well as parse a string into those protobuf types.
//
// The notion of "value renderer" is defined in ADR-050, and that ADR provides
// a default spec for value renderers. However, we define it as an interface
// here, so that optionally more value renderers could be built, for example, a
// separate one for a different language.
type ValueRenderer interface {
	// Format should render the value to a text plus annotation.
	Format(context.Context, protoreflect.Value) ([]Item, error)

	// Parse should be the inverse of Format.
	Parse(context.Context, []Item) (protoreflect.Value, error)
}
