package aminojson

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type JSONMarshaller interface {
	MarshalAmino(proto.Message) ([]byte, error)
}

type AminoJson struct{}

func MarshalAmino(message proto.Message) ([]byte, error) {
	buf := &bytes.Buffer{}
	aj := AminoJson{}

	vmsg := protoreflect.ValueOfMessage(message.ProtoReflect())
	err := aj.marshal(vmsg, nil, buf)
	return buf.Bytes(), err
}

func (aj AminoJson) marshal(
	value protoreflect.Value,
	field protoreflect.FieldDescriptor,
	writer io.Writer) error {

	switch val := value.Interface().(type) {
	case protoreflect.Message:
		return aj.marshalMessage(val, writer)

	case protoreflect.Map:
		return errors.New("maps are not supported")

	case protoreflect.List:
		return aj.marshalList(field, val, writer)

	case string, bool, int32, uint32, protoreflect.EnumNumber:
		return invokeStdlibJSONMarshal(writer, val)

	case uint64, int64:
		_, err := fmt.Fprintf(writer, `"%d"`, val) // quoted
		return err

	case []byte:
		_, err := fmt.Fprintf(writer, `"%s"`,
			base64.StdEncoding.EncodeToString(val))
		return err
	}

	return nil
}

func (aj AminoJson) marshalMessage(msg protoreflect.Message, writer io.Writer) error {
	_, err := writer.Write([]byte("{"))
	if err != nil {
		return err
	}

	fields := msg.Descriptor().Fields()
	first := true
	for i := 0; i < fields.Len(); i++ {
		f := fields.Get(i)
		v := msg.Get(f)

		if !msg.Has(f) {
			continue
		}

		if !first {
			_, err = writer.Write([]byte(","))
			if err != nil {
				return err
			}
		}

		err = invokeStdlibJSONMarshal(writer, f.Name())
		if err != nil {
			return err
		}

		_, err = writer.Write([]byte(":"))
		if err != nil {
			return err
		}

		err = aj.marshal(v, f, writer)
		if err != nil {
			return err
		}

		first = false
	}

	_, err = writer.Write([]byte("}"))
	if err != nil {
		return err
	}
	return nil
}

func invokeStdlibJSONMarshal(w io.Writer, v interface{}) error {
	blob, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = w.Write(blob)
	return err
}

func (aj AminoJson) marshalList(
	fieldDescriptor protoreflect.FieldDescriptor,
	list protoreflect.List,
	writer io.Writer) error {
	n := list.Len()
	_, err := writer.Write([]byte("["))
	if err != nil {
		return err
	}

	first := true
	for i := 0; i < n; i++ {
		if !first {
			_, err := writer.Write([]byte(","))
			if err != nil {
				return err
			}
		}
		first = false

		err = aj.marshal(list.Get(i), fieldDescriptor, writer)
		if err != nil {
			return err
		}
	}

	_, err = writer.Write([]byte("]"))
	return err
}
