package flag

import (
	"context"

	"github.com/spf13/pflag"
)

type Type interface {
	NewValue(context.Context, *Builder) pflag.Value
	DefaultValue() string
}
