package aminojson

import (
	"fmt"
	"io"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

const (
	typeUrlName protoreflect.Name = "type_url"
	valueName   protoreflect.Name = "value"
)

func (aj AminoJSON) marshalAny(message protoreflect.Message, writer io.Writer) error {
	fields := message.Descriptor().Fields()
	typeUrlField := fields.ByName(typeUrlName)
	if typeUrlField == nil {
		return fmt.Errorf("expected type_url field")
	}

	typeUrl := message.Get(typeUrlField).String()
	// TODO
	// need an override for this?
	resolver := protoregistry.GlobalTypes

	typ, err := resolver.FindMessageByURL(typeUrl)
	if err != nil {
		return errors.Wrapf(err, "can't resolve type URL %s", typeUrl)
	}

	valueField := fields.ByName(valueName)
	if valueField == nil {
		return fmt.Errorf("expected value field")
	}

	valueBz := message.Get(valueField).Bytes()

	valueMsg := typ.New()
	err = proto.Unmarshal(valueBz, valueMsg.Interface())
	if err != nil {
		return err
	}

	aminoName, named := getMessageName(valueMsg)
	if !named {
		return fmt.Errorf("any fields require an amino.name annotation")
	}

	_, err = writer.Write([]byte(fmt.Sprintf(`{"type":"%s","value":`, aminoName)))

	err = aj.marshalMessage(valueMsg, writer)
	if err != nil {
		return err
	}

	_, err = writer.Write([]byte("}"))
	return err
}
