package flag

import (
	"context"
)

type Type interface {
	NewValue(context.Context, *Options) SimpleValue
	DefaultValue() string
}
