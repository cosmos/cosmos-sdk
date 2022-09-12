package flag

import (
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// MessageBinder binds multiple flags in a flag set to a protobuf message.
type MessageBinder struct {
	CobraArgs cobra.PositionalArgs

	positionalArgs []struct {
		binder  Value
		field   protoreflect.FieldDescriptor
		varargs bool
	}

	flagFieldPairs []struct {
		binder FieldValueBinder
		field  protoreflect.FieldDescriptor
	}
	messageType protoreflect.MessageType
}

// BuildMessage builds and returns a new message for the bound flags.
func (m MessageBinder) BuildMessage(positionalArgs []string) (protoreflect.Message, error) {
	msg := m.messageType.New()
	err := m.Bind(msg, positionalArgs)
	return msg, err
}

// Bind binds the flag values to an existing protobuf message.
func (m MessageBinder) Bind(msg protoreflect.Message, positionalArgs []string) error {
	n := len(positionalArgs)
	for i, arg := range m.positionalArgs {
		if i >= n {
			panic("unexpected: validate args should have caught this")
		}

		if arg.varargs {
			for _, v := range positionalArgs[i:] {
				err := arg.binder.Set(v)
				if err != nil {
					return err
				}
			}
		} else {
			err := arg.binder.Set(positionalArgs[i])
			if err != nil {
				return err
			}
		}
	}

	for _, pair := range m.flagFieldPairs {
		pair.binder.Bind(msg, pair.field)
	}

	return nil
}

// Get calls BuildMessage and wraps the result in a protoreflect.Value.
func (m MessageBinder) Get() (protoreflect.Value, error) {
	msg, err := m.BuildMessage(nil)
	return protoreflect.ValueOfMessage(msg), err
}
