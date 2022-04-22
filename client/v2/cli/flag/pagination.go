package flag

import (
	"context"

	"github.com/spf13/pflag"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func (b *Builder) bindPageRequest(ctx context.Context, flagSet *pflag.FlagSet, field protoreflect.FieldDescriptor) FieldBinder {
	handler := b.RegisterMessageFlags(ctx, flagSet, b.ResolveMessageType(field.Message()), Options{Prefix: "page-"})
	return simpleValueBinder{handler}
}
