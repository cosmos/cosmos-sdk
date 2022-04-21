package cli

import (
	"context"

	"github.com/spf13/pflag"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/dynamicpb"
)

type Builder struct {
	GetClientConn       func(context.Context) grpc.ClientConnInterface
	MessageTypeResolver protoregistry.MessageTypeResolver
	JSONMarshalOptions  protojson.MarshalOptions
}

func (b *Builder) resolverMessageType(descriptor protoreflect.MessageDescriptor) protoreflect.MessageType {
	resolver := b.MessageTypeResolver
	if resolver == nil {
		resolver = protoregistry.GlobalTypes
	}

	typ, err := resolver.FindMessageByName(descriptor.FullName())
	if err == nil {
		return typ
	}

	return dynamicpb.NewMessageType(descriptor)
}

func (b *Builder) DefineMessageFlagType(messageName protoreflect.FullName, flagType FlagType) {

}

func (b *Builder) DefineScalarFlagType(scalarName string, flagType FlagType) {

}

type FlagType interface {
	AddFlag(*pflag.FlagSet, protoreflect.FieldDescriptor) FlagValue
}

type FlagValue interface {
	Get() protoreflect.Value
}
