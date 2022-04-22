package flag

import (
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/dynamicpb"
)

type Builder struct {
	Resolver interface {
		protoregistry.MessageTypeResolver
		protoregistry.ExtensionTypeResolver
	}
	messageFlagTypes map[protoreflect.FullName]Type
	scalarFlagTypes  map[string]Type
}

func (b *Builder) ResolveMessageType(descriptor protoreflect.MessageDescriptor) protoreflect.MessageType {
	resolver := b.Resolver
	if resolver == nil {
		resolver = protoregistry.GlobalTypes
	}

	typ, err := resolver.FindMessageByName(descriptor.FullName())
	if err == nil {
		return typ
	}

	return dynamicpb.NewMessageType(descriptor)
}

func (b *Builder) init() {
	if b.messageFlagTypes == nil {
		b.messageFlagTypes = map[protoreflect.FullName]Type{}
		b.messageFlagTypes["google.protobuf.Timestamp"] = timestampType{}
		b.messageFlagTypes["google.protobuf.Duration"] = durationType{}
	}

	if b.scalarFlagTypes == nil {
		b.scalarFlagTypes = map[string]Type{}
		b.scalarFlagTypes["cosmos.AddressString"] = addressStringType{}
	}
}

func (b *Builder) DefineMessageFlagType(messageName protoreflect.FullName, flagType Type) {
	b.init()
	b.messageFlagTypes[messageName] = flagType
}

func (b *Builder) DefineScalarFlagType(scalarName string, flagType Type) {
	b.init()
	b.scalarFlagTypes[scalarName] = flagType
}
