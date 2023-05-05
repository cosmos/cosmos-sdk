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
	anyMsg := message.Interface().(*anypb.Any)

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
