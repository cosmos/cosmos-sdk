package stablejson

import (
	"encoding/base64"
	"fmt"
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protopath"
	"google.golang.org/protobuf/reflect/protorange"
	"google.golang.org/protobuf/reflect/protoreflect"
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

func (opts MarshalOptions) Marshal(message proto.Message) ([]byte, error) {
	writer := &strings.Builder{}
	firstStack := []bool{true}
	closingBraceStack := []bool{}
	err := protorange.Options{
		Stable: true,
	}.Range(message.ProtoReflect(),
		// push
		func(p protopath.Values) error {
			// Starting printing the value.
			if !firstStack[len(firstStack)-1] {
				writer.WriteString(",")
			}
			firstStack[len(firstStack)-1] = false

			// Print the key.
			var fd protoreflect.FieldDescriptor
			last := p.Index(-1)
			beforeLast := p.Index(-2)
			switch last.Step.Kind() {
			case protopath.FieldAccessStep:
				fd = last.Step.FieldDescriptor()
				fullName := fd.FullName()
				if fullName != structFieldsFullName && fullName != listValueValuesFullName {
					_, _ = fmt.Fprintf(writer, "%q:", fd.Name())
				}

			case protopath.ListIndexStep:
				fd = beforeLast.Step.FieldDescriptor() // lists always appear in the context of a repeated field

			case protopath.MapIndexStep:
				fd = beforeLast.Step.FieldDescriptor() // maps always appear in the context of a repeated field
				_, _ = fmt.Fprintf(writer, "%q:", last.Step.MapIndex().String())

			case protopath.AnyExpandStep:
				_, _ = fmt.Fprintf(writer, `"@type":%q`, last.Value.Message().Descriptor().FullName())
				return nil

			case protopath.UnknownAccessStep:
				writer.WriteString("?:")
			}

			switch value := last.Value.Interface().(type) {
			case protoreflect.Message:
				closingBrace, err := marshalMessage(writer, value)
				if err != nil {
					return err
				}

				firstStack = append(firstStack, true)
				closingBraceStack = append(closingBraceStack, closingBrace)
			case protoreflect.List:
				writer.WriteString("[")
				firstStack = append(firstStack, true)
			case protoreflect.Map:
				_, _ = fmt.Fprintf(writer, "{")
				firstStack = append(firstStack, true)
			case protoreflect.EnumNumber:
				var ev protoreflect.EnumValueDescriptor
				if fd != nil {
					ev = fd.Enum().Values().ByNumber(value)
				}
				if ev != nil {
					_, _ = fmt.Fprintf(writer, "%q", ev.Name())
				} else {
					_, _ = fmt.Fprintf(writer, "%v", value)
				}
			case string:
				_, _ = fmt.Fprintf(writer, "%q", value)
			case []byte:
				writer.WriteString(`"`)
				if opts.HexBytes {
					_, _ = fmt.Fprintf(writer, "%X", value)
				} else {
					b64 := base64.URLEncoding.EncodeToString(value)
					writer.WriteString(b64)
				}
				writer.WriteString(`"`)
			case bool:
				_, _ = fmt.Fprintf(writer, "%t", value)
			case int32:
				_, _ = fmt.Fprintf(writer, "%d", value)
			case uint32:
				_, _ = fmt.Fprintf(writer, "%d", value)
			case int64:
				_, _ = fmt.Fprintf(writer, `"%d"`, value) // quoted
			case uint64:
				_, _ = fmt.Fprintf(writer, `"%d"`, value) // quoted
			case float32:
				marshalFloat(writer, float64(value))
			case float64:
				marshalFloat(writer, value)
			default:
				return fmt.Errorf("unexpected type %T", value)
			}
			return nil
		},
		// pop
		func(p protopath.Values) error {
			last := p.Index(-1)
			switch last.Value.Interface().(type) {
			case protoreflect.Message:
				if last.Step.Kind() != protopath.AnyExpandStep {
					n := len(closingBraceStack)
					if n > 0 {
						if closingBraceStack[n-1] {
							_, _ = fmt.Fprintf(writer, "}")
						}
						closingBraceStack = closingBraceStack[:n-1]
					}
				}
			case protoreflect.List:
				_, _ = fmt.Fprintf(writer, "]")
				firstStack = firstStack[:len(firstStack)-1]
			case protoreflect.Map:
				_, _ = fmt.Fprintf(writer, "}")
				firstStack = firstStack[:len(firstStack)-1]
			}
			return nil
		},
	)
	return []byte(writer.String()), err
}
