package aminojson

import (
	"fmt"
	"google.golang.org/protobuf/types/known/anypb"
	"io"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

func (aj AminoJSON) marshalAny(message protoreflect.Message, writer io.Writer) error {
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

	aminoName, named := getMessageAminoName(valueMsg)
	if !named {
		return fmt.Errorf("message %s is packed into an any field, so requires an amino.name annotation")
	}

	_, err = writer.Write([]byte(fmt.Sprintf(`{"type":"%s","value":`, aminoName)))
	if err != nil {
		return err
	}

	err = aj.marshalMessage(valueMsg, writer)
	if err != nil {
		return err
	}

	_, err = writer.Write([]byte("}"))
	return err
}
