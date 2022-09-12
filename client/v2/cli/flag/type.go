package flag

import (
	"context"
)

// Type specifies a custom flag type.
type Type struct {
	// NewValue returns a new pflag.Value which must also implement either
	// SimpleValue or ListValue.
	NewValue func(context.Context, *Builder) Value

	// DefaultValue is the default value for this type. If it is set to the
	// empty string, NewValue().String will be used.
	DefaultValue string

	NoOptDefaultValue string
}
