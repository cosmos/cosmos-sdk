package stablejson

import (
	"bytes"
	"fmt"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protopath"
	"google.golang.org/protobuf/reflect/protorange"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func Marshal(message proto.Message) ([]byte, error) {
	return MarshalOptions{}.Marshal(message)
}

type MarshalOptions struct {
	// HexBytes specifies whether bytes fields should be marshaled as upper-case
	// hex strings. If set to false, bytes fields will be encoded as padded
	// URL-safe base64 strings as specified by the official proto3 JSON mapping.
	HexBytes bool
}

func (o MarshalOptions) Marshal(message proto.Message) ([]byte, error) {
	buf := &bytes.Buffer{}
	firstStack := []bool{true}
	err := protorange.Options{
		Stable: true,
	}.Range(message.ProtoReflect(),
		// push
		func(p protopath.Values) error {
			// Starting printing the value.
			if !firstStack[len(firstStack)-1] {
				_, _ = fmt.Fprintf(buf, ",")
			}
			firstStack[len(firstStack)-1] = false

			// Print the key.
			var fd protoreflect.FieldDescriptor
			last := p.Index(-1)
			beforeLast := p.Index(-2)
			switch last.Step.Kind() {
			case protopath.FieldAccessStep:
				fd = last.Step.FieldDescriptor()
				_, _ = fmt.Fprintf(buf, "%q:", fd.Name())

			case protopath.ListIndexStep:
				fd = beforeLast.Step.FieldDescriptor() // lists always appear in the context of a repeated field

			case protopath.MapIndexStep:
				fd = beforeLast.Step.FieldDescriptor() // maps always appear in the context of a repeated field
				_, _ = fmt.Fprintf(buf, "%q:", last.Step.MapIndex().String())
				//last.Step.MapIndex().Interface().(type) {
				//switch mapKey := last.Step.MapIndex().Interface().(type) {
				//case string:
				//	_, _ = fmt.Fprintf(buf, "%q", mapKey)
				//default:
				//	_, _ = fmt.Fprintf(buf, "%v", mapKey)
				//}
				//_, _ = fmt.Fprint(buf, ":")

			case protopath.AnyExpandStep:
				_, _ = fmt.Fprintf(buf, `"@type":%q`, last.Value.Message().Descriptor().FullName())
				return nil

			case protopath.UnknownAccessStep:
				_, _ = fmt.Fprintf(buf, "?: ")
			}

			switch value := last.Value.Interface().(type) {
			case protoreflect.Message:
				if value.Descriptor().FullName() == timestampFullName {
					return protorange.Break
				}

				if value.Descriptor().FullName() == durationFullName {
					return protorange.Break
				}

				_, _ = fmt.Fprintf(buf, "{")
				firstStack = append(firstStack, true)
			case protoreflect.List:
				_, _ = fmt.Fprintf(buf, "[")
				firstStack = append(firstStack, true)
			case protoreflect.Map:
				_, _ = fmt.Fprintf(buf, "{")
				firstStack = append(firstStack, true)
			case protoreflect.EnumNumber:
				var ev protoreflect.EnumValueDescriptor
				if fd != nil {
					ev = fd.Enum().Values().ByNumber(value)
				}
				if ev != nil {
					_, _ = fmt.Fprintf(buf, "%q", ev.Name())
				} else {
					_, _ = fmt.Fprintf(buf, "%v", value)
				}
			case string:
				_, _ = fmt.Fprintf(buf, "%q", value)
			case []byte:
				_, _ = fmt.Fprintf(buf, "%X", value)
			default:
				_, _ = fmt.Fprintf(buf, "%v", value)
			}
			return nil
		},
		// pop
		func(p protopath.Values) error {
			last := p.Index(-1)
			switch last.Value.Interface().(type) {
			case protoreflect.Message:
				if last.Step.Kind() != protopath.AnyExpandStep {
					_, _ = fmt.Fprintf(buf, "}")
				}
			case protoreflect.List:
				_, _ = fmt.Fprintf(buf, "]")
				firstStack = firstStack[:len(firstStack)-1]
			case protoreflect.Map:
				_, _ = fmt.Fprintf(buf, "}")
				firstStack = firstStack[:len(firstStack)-1]
			}
			return nil
		},
	)
	return buf.Bytes(), err
}

var (
	timestampMsgType  = (&timestamppb.Timestamp{}).ProtoReflect().Type()
	timestampFullName = timestampMsgType.Descriptor().FullName()
	durationMsgType   = (&durationpb.Duration{}).ProtoReflect().Type()
	durationFullName  = durationMsgType.Descriptor().FullName()
)
