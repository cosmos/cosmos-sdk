package flag

import (
	"context"
	"fmt"

	"github.com/spf13/pflag"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/cosmos/cosmos-sdk/client/v2/internal/util"
)

type jsonMessageFlagType struct {
	messageDesc protoreflect.MessageDescriptor
}

func (j jsonMessageFlagType) NewValue(_ context.Context, builder *Builder) pflag.Value {
	return &jsonMessageFlagValue{
		messageType:          util.ResolveMessageType(builder.TypeResolver, j.messageDesc),
		jsonMarshalOptions:   protojson.MarshalOptions{Resolver: builder.TypeResolver},
		jsonUnmarshalOptions: protojson.UnmarshalOptions{Resolver: builder.TypeResolver},
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
	if j.message == nil {
		return protoreflect.Value{}
	}
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
	return fmt.Sprintf("%s (json)", j.messageType.Descriptor().FullName())
}
