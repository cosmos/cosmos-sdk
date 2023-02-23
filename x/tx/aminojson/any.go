package aminojson

import (
	"fmt"
	"io"

	"google.golang.org/protobuf/types/known/anypb"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

func (enc Encoder) marshalAny(message protoreflect.Message, writer io.Writer) error {
	anyMsg := message.Interface().(*anypb.Any)
	resolver := protoregistry.GlobalTypes

	typ, err := resolver.FindMessageByURL(anyMsg.TypeUrl)
	if err != nil {
		return errors.Wrapf(err, "can't resolve type URL %s", anyMsg.TypeUrl)
	}

	valueMsg := typ.New()
	err = proto.Unmarshal(anyMsg.Value, valueMsg.Interface())
	if err != nil {
		return err
	}

	_, named := getMessageAminoName(valueMsg)
	if !named {
		return fmt.Errorf("message %s is packed into an any field, so requires an amino.name annotation",
			anyMsg.TypeUrl)
	}

	return enc.beginMarshal(valueMsg, writer)
}
