package stablejson

import (
	"fmt"
	"io"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

const (
	typeUrlName protoreflect.Name = "type_url"
	valueName   protoreflect.Name = "value"
)

func (opts MarshalOptions) marshalAny(message protoreflect.Message, writer io.Writer) error {
	fields := message.Descriptor().Fields()
	typeUrlField := fields.ByName(typeUrlName)
	if typeUrlField == nil {
		return fmt.Errorf("expected type_url field")
	}

	_, err := writer.Write([]byte("{"))
	if err != nil {
		return err
	}

	typeUrl := message.Get(typeUrlField).String()
	resolver := opts.Resolver
	if resolver == nil {
		resolver = protoregistry.GlobalTypes
	}
	typ, err := resolver.FindMessageByURL(typeUrl)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(writer, `"@type_url":%q`, typeUrl)
	if err != nil {
		return err
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

	err = opts.marshalMessageFields(valueMsg, writer, false)
	if err != nil {
		return err
	}

	_, err = writer.Write([]byte("}"))
	return err
}
