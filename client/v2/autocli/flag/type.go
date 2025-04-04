package flag

import (
	"context"
)

// Type specifies a custom flag type.
type Type interface {
	// NewValue returns a new pflag.Value which must also implement either
	// SimpleValue or ListValue.
	NewValue(context.Context, *Builder) Value

	// DefaultValue is the default value for this type.
	DefaultValue() string
}
