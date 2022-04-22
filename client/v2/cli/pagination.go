package cli

import (
	"context"

	"github.com/spf13/pflag"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type pageRequestFlagType struct{}

func (p pageRequestFlagType) AddFlag(ctx context.Context, builder *Builder, set *pflag.FlagSet, descriptor protoreflect.FieldDescriptor) FlagValue {
	typ := builder.resolverMessageType(descriptor.Message())
	handler := builder.registerMessageFlagSet(ctx, set, typ)
	return &pageRequestFlagValue{handler: handler}
}

type pageRequestFlagValue struct {
	handler *messageFlagHandler
}

func (p pageRequestFlagValue) Get() protoreflect.Value {
	return protoreflect.ValueOfMessage(p.handler.buildMessage())
}
