package flag

import (
	"context"

	"github.com/spf13/pflag"
)

// Type specifies a custom flag type.
type Type interface {
	// NewValue returns a new pflag.Value which must also implement either
	// SimpleValue or ListValue.
	NewValue(context.Context, *Builder) pflag.Value

	// DefaultValue is the default value for this type.
	DefaultValue() string
}
