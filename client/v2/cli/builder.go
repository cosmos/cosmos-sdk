package cli

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/dynamicpb"
)

type Builder struct {
	GetClientConn       func(context.Context) grpc.ClientConn
	MessageTypeResolver protoregistry.MessageTypeResolver
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
