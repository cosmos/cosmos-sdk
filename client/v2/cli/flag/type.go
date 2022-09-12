package flag

import (
	"context"
)

// Type specifies a custom flag type.
type Type interface {
	// NewValue returns a new pflag.Value which must also implement either
	// SimpleValue or ListValue.
	NewValue(context.Context, *Builder) Value
}

// HasDefaultValue defines a type which has an explicitly specified default
// value. If this interface isn't specified, Type.NewValue().String() will
// be used.
type HasDefaultValue interface {
	// DefaultValue is the default value for this type.
	DefaultValue() string
}

// HasNoOptDefaultValue defines a type which has an explicitly specified no-option
// default value. If this interface isn't specified, the empty string will be used.
type HasNoOptDefaultValue interface {
	// NoOptDefaultValue is the no-option default value for this type.
	NoOptDefaultValue() string
}
