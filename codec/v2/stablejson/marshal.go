package stablejson

import (
	"bytes"
	"fmt"
	"io"

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
	// hex strings. If set to false, bytes fields will be encoded as standard
	// base64 strings as specified by the official proto3 JSON mapping.
	HexBytes bool
}

func (opts MarshalOptions) Marshal(message proto.Message) ([]byte, error) {
	buf := &bytes.Buffer{}
	err := opts.MarshalTo(message, buf)
	return buf.Bytes(), err
}

func (opts MarshalOptions) MarshalTo(message proto.Message, writer io.Writer) error {
	firstStack := []bool{true}
	skipNext := false
	return protorange.Options{
		Stable: true,
	}.Range(message.ProtoReflect(),
		// push
		func(p protopath.Values) error {
			if skipNext {
				skipNext = false
				return protorange.Break
			}

			// Starting printing the value.
			if !firstStack[len(firstStack)-1] {
				_, err := writer.Write([]byte(","))
				if err != nil {
					return err
				}
			}
			firstStack[len(firstStack)-1] = false

			// Print the key.
			var fd protoreflect.FieldDescriptor
			last := p.Index(-1)
			beforeLast := p.Index(-2)
			switch last.Step.Kind() {
			case protopath.FieldAccessStep:
				fd = last.Step.FieldDescriptor()
				_, err := fmt.Fprintf(writer, "%q:", fd.Name())
				if err != nil {
					return err
				}

			case protopath.ListIndexStep:
				fd = beforeLast.Step.FieldDescriptor() // lists always appear in the context of a repeated field

			case protopath.MapIndexStep:
				fd = beforeLast.Step.FieldDescriptor() // maps always appear in the context of a repeated field
				_, err := fmt.Fprintf(writer, "%q:", last.Step.MapIndex().String())
				if err != nil {
					return err
				}

			case protopath.AnyExpandStep:
				_, err := fmt.Fprintf(writer, `"@type":%q`, last.Value.Message().Descriptor().FullName())
				return err

			case protopath.UnknownAccessStep:
				return fmt.Errorf("unexpected %s", protopath.UnknownAccessStep)
			}

			switch value := last.Value.Interface().(type) {
			case protoreflect.Message:
				continueRange, err := opts.marshalMessage(writer, value)
				if err != nil {
					return err
				}

				if !continueRange {
					skipNext = true
					return nil
				}

				firstStack = append(firstStack, true)
			case protoreflect.List:
				_, err := writer.Write([]byte("["))
				if err != nil {
					return err
				}
				firstStack = append(firstStack, true)
			case protoreflect.Map:
				_, err := fmt.Fprintf(writer, "{")
				if err != nil {
					return err
				}
				firstStack = append(firstStack, true)
			case protoreflect.EnumNumber:
				var ev protoreflect.EnumValueDescriptor
				if fd != nil {
					ev = fd.Enum().Values().ByNumber(value)
				}
				var err error
				if ev != nil {
					_, err = fmt.Fprintf(writer, "%q", ev.Name())
				} else {
					_, err = fmt.Fprintf(writer, "%v", value)
				}
				if err != nil {
					return err
				}
			case string:
				_, err := fmt.Fprintf(writer, "%q", value)
				if err != nil {
					return err
				}
			default:
				return opts.marshalScalar(writer, value)
			}
			return nil
		},
		// pop
		func(p protopath.Values) error {
			last := p.Index(-1)
			switch last.Value.Interface().(type) {
			case protoreflect.Message:
				if last.Step.Kind() != protopath.AnyExpandStep {
					_, _ = fmt.Fprintf(writer, "}")
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
}
