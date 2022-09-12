package flag

import (
	"context"

	"github.com/spf13/pflag"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/cosmos/cosmos-sdk/client/v2/internal/util"
)

func (b *Builder) bindPageRequest(ctx context.Context, flagSet *pflag.FlagSet, field protoreflect.FieldDescriptor) (FieldValueBinder, error) {
	handler, err := b.AddMessageFlags(ctx, flagSet, util.ResolveMessageType(b.TypeResolver, field.Message()), nil, Options{Prefix: "page-"})
	if err != nil {
		return nil, err
	}
	return simpleValueBinder{handler}, nil
}
