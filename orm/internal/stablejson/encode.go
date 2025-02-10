package stablejson

import (
	"bytes"
	"fmt"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protopath"
	"google.golang.org/protobuf/reflect/protorange"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// Marshal marshals the provided message to JSON with a stable ordering based
// on ascending field numbers.
func Marshal(message proto.Message) ([]byte, error) {
	buf := &bytes.Buffer{}
	firstStack := []bool{true}
	err := protorange.Options{
		Stable: true,
	}.Range(message.ProtoReflect(),
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
				_, _ = fmt.Fprintf(buf, "%v:", last.Step.MapIndex().Interface())
			case protopath.AnyExpandStep:
				_, _ = fmt.Fprintf(buf, `"@type":%q`, last.Value.Message().Descriptor().FullName())
				return nil
			case protopath.UnknownAccessStep:
				_, _ = fmt.Fprintf(buf, "?: ")
			}

			switch v := last.Value.Interface().(type) {
			case protoreflect.Message:
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
					ev = fd.Enum().Values().ByNumber(v)
				}
				if ev != nil {
					_, _ = fmt.Fprintf(buf, "%v", ev.Name())
				} else {
					_, _ = fmt.Fprintf(buf, "%v", v)
				}
			case string, []byte:
				_, _ = fmt.Fprintf(buf, "%q", v)
			default:
				_, _ = fmt.Fprintf(buf, "%v", v)
			}
			return nil
		},
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
