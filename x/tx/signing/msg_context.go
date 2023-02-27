package signing

import (
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

type MsgContext struct {
	protoFiles      *protoregistry.Files
	protoTypes      protoregistry.MessageTypeResolver
	getSignersFuncs map[protoreflect.FullName]getSignersFunc
}

type MsgContextOptions struct {
	ProtoFiles *protoregistry.Files
	ProtoTypes protoregistry.MessageTypeResolver
}

func (c MsgContextOptions) Build() *MsgContext {
	protoFiles := c.ProtoFiles
	if protoFiles == nil {
		protoFiles = protoregistry.GlobalFiles
	}

	protoTypes := c.ProtoTypes
	if protoTypes == nil {
		protoTypes = protoregistry.GlobalTypes
	}

	return &MsgContext{
		protoFiles:      protoFiles,
		protoTypes:      protoTypes,
		getSignersFuncs: map[protoreflect.FullName]getSignersFunc{},
	}
}
