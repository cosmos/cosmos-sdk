package flag

import (
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/dynamicpb"
)

type Options struct {
	Resolver interface {
		protoregistry.MessageTypeResolver
		protoregistry.ExtensionTypeResolver
	}
	messageFlagTypes map[protoreflect.FullName]interface{}
	scalarFlagTypes  map[string]interface{}
}

func (b *Options) resolverMessageType(descriptor protoreflect.MessageDescriptor) protoreflect.MessageType {
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

func (b *Options) init() {
	if b.messageFlagTypes == nil {
		b.messageFlagTypes = map[protoreflect.FullName]interface{}{}
		// TODO b.messageFlagTypes["cosmos.base.query.v1beta1.PageRequest"] = pageRequestTYpe{}
	}

	if b.scalarFlagTypes == nil {
		b.scalarFlagTypes = map[string]interface{}{}
		b.scalarFlagTypes["cosmos.AddressString"] = addressStringType{}
	}
}

func (b *Options) DefineMessageFlagType(messageName protoreflect.FullName, flagType Type) {
	b.init()
	b.messageFlagTypes[messageName] = flagType
}

func (b *Options) DefineScalarFlagType(scalarName string, flagType Type) {
	b.init()
	b.scalarFlagTypes[scalarName] = flagType
}
