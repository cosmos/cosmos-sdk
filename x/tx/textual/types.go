package textual

import (
	"context"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// Screen is the abstract unit of Textual rendering.
type Screen struct {
	// Title is the text (sequence of Unicode code points) to display first,
	// generally on the device's title section. It can be empty.
	Title string

	// Content is the text (sequence of Unicode code points) to display after
	// the Title, generally on the device's content section. It must be
	// non-empty.
	Content string

	// Indent is the indentation level of the screen.
	// Zero indicates top-level. Should be less than 16.
	Indent int

	// Expert indicates that the screen should only be displayed
	// via an opt-in from the user.
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
	Format(context.Context, protoreflect.Value) ([]Screen, error)

	// Parse should be the inverse of Format.
	Parse(context.Context, []Screen) (protoreflect.Value, error)
}

// RepeatedValueRenderer defines an interface to produce formatted output for
// protobuf message fields that are repeated.
type RepeatedValueRenderer interface {
	ValueRenderer

	// FormatRepeated should render the value to a text plus annotation.
	FormatRepeated(context.Context, protoreflect.Value) ([]Screen, error)

	// ParseRepeated should be the inverse of Format.  The list will be populated with the repeated values.
	ParseRepeated(context.Context, []Screen, protoreflect.List) error
}
