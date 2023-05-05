package aminojson

import (
	"fmt"
	"io"

	"google.golang.org/protobuf/types/dynamicpb"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func (enc Encoder) marshalAny(message protoreflect.Message, writer io.Writer) error {
	_, ok := message.Interface().(*dynamicpb.Message)
	if ok {
		return enc.marshalDynamic(message, writer)
	}

	anyMsg, ok := message.Interface().(*anypb.Any)
	if !ok {
		return fmt.Errorf("expected *anypb.Any, got %T", message.Interface())
	}

	desc, err := enc.fileResolver.FindDescriptorByName(protoreflect.FullName(anyMsg.TypeUrl[1:]))
	if err != nil {
		return errors.Wrapf(err, "can't resolve type URL %s", anyMsg.TypeUrl)
	}

	_, named := getMessageAminoName(desc.Options())
	if !named {
		return fmt.Errorf("message %s is packed into an any field, so requires an amino.name annotation",
			anyMsg.TypeUrl)
	}

	valueMsg := dynamicpb.NewMessageType(desc.(protoreflect.MessageDescriptor)).New().Interface()
	err = proto.Unmarshal(anyMsg.Value, valueMsg)
	if err != nil {
		return err
	}
	protoMessage := valueMsg.ProtoReflect()

	return enc.beginMarshal(protoMessage, writer)
}

func (enc Encoder) marshalDynamic(message protoreflect.Message, writer io.Writer) error {
	msgName := message.Get(message.Descriptor().Fields().ByName("type_url")).String()[1:]
	msgBytes := message.Get(message.Descriptor().Fields().ByName("value")).Bytes()

	desc, err := enc.fileResolver.FindDescriptorByName(protoreflect.FullName(msgName))
	if err != nil {
		return errors.Wrapf(err, "can't resolve type URL %s", msgName)
	}

	_, named := getMessageAminoName(desc.Options())
	if !named {
		return fmt.Errorf("message %s is packed into an any field, so requires an amino.name annotation",
			msgName)
	}

	valueMsg := dynamicpb.NewMessageType(desc.(protoreflect.MessageDescriptor)).New().Interface()
	err = proto.Unmarshal(msgBytes, valueMsg)
	if err != nil {
		return err
	}

	return enc.beginMarshal(valueMsg.ProtoReflect(), writer)
}
