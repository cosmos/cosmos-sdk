package flag

import (
	"context"

	"github.com/spf13/pflag"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/cosmos/cosmos-sdk/client/v2/internal/util"
)

func (b *Builder) bindPageRequest(ctx context.Context, flagSet *pflag.FlagSet, field protoreflect.FieldDescriptor) FieldValueBinder {
	handler := b.AddMessageFlags(
		ctx,
		flagSet,
		util.ResolveMessageType(b.TypeResolver, field.Message()),
		Options{Prefix: "page-"},
	)
	return simpleValueBinder{handler}
}
