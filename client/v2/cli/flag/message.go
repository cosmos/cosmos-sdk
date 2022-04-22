package flag

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type jsonMessageFlagType struct {
	messageDesc protoreflect.MessageDescriptor
}

func (j jsonMessageFlagType) NewValue(ctx context.Context, builder *Options) SimpleValue {
	return &jsonMessageFlagValue{
		messageType:          builder.resolverMessageType(j.messageDesc),
		jsonMarshalOptions:   protojson.MarshalOptions{Resolver: builder.Resolver},
		jsonUnmarshalOptions: protojson.UnmarshalOptions{Resolver: builder.Resolver},
	}
}

func (j jsonMessageFlagType) DefaultValue() string {
	return ""
}

type jsonMessageFlagValue struct {
	jsonMarshalOptions   protojson.MarshalOptions
	jsonUnmarshalOptions protojson.UnmarshalOptions
	messageType          protoreflect.MessageType
	message              proto.Message
}

func (j jsonMessageFlagValue) Get() protoreflect.Value {
	return protoreflect.ValueOfMessage(j.message.ProtoReflect())
}

func (j jsonMessageFlagValue) String() string {
	if j.message == nil {
		return ""
	}

	bz, err := j.jsonMarshalOptions.Marshal(j.message)
	if err != nil {
		return err.Error()
	}
	return string(bz)
}

func (j *jsonMessageFlagValue) Set(s string) error {
	j.message = j.messageType.New().Interface()
	return j.jsonUnmarshalOptions.Unmarshal([]byte(s), j.message)
}

func (j jsonMessageFlagValue) Type() string {
	return fmt.Sprintf("%s (json string or file)", j.messageType.Descriptor().FullName())
}
