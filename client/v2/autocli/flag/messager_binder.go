package flag

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// MessageBinder binds multiple flags in a flag set to a protobuf message.
type MessageBinder struct {
	CobraArgs cobra.PositionalArgs

	positionalFlagSet *pflag.FlagSet
	positionalArgs    []fieldBinding
	hasVarargs        bool
	hasOptional       bool

	flagBindings []fieldBinding
	messageType  protoreflect.MessageType
}

// BuildMessage builds and returns a new message for the bound flags.
func (m MessageBinder) BuildMessage(positionalArgs []string) (protoreflect.Message, error) {
	msg := m.messageType.New()
	err := m.Bind(msg, positionalArgs)
	return msg, err
}

// Bind binds the flag values to an existing protobuf message.
func (m MessageBinder) Bind(msg protoreflect.Message, positionalArgs []string) error {
	// first set positional args in the positional arg flag set
	n := len(positionalArgs)
	for i := range m.positionalArgs {
		if i == n {
			break
		}

		name := fmt.Sprintf("%d", i)
		if i == n-1 && m.hasVarargs {
			for _, v := range positionalArgs[i:] {
				err := m.positionalFlagSet.Set(name, v)
				if err != nil {
					return err
				}
			}
		} else {
			err := m.positionalFlagSet.Set(name, positionalArgs[i])
			if err != nil {
				return err
			}
		}
	}

	// bind positional arg values to the message
	for _, arg := range m.positionalArgs {
		err := arg.bind(msg)
		if err != nil {
			return err
		}
	}

	// bind flag values to the message
	for _, binding := range m.flagBindings {
		err := binding.bind(msg)
		if err != nil {
			return err
		}
	}

	return nil
}

// Get calls BuildMessage and wraps the result in a protoreflect.Value.
func (m MessageBinder) Get(protoreflect.Value) (protoreflect.Value, error) {
	msg, err := m.BuildMessage(nil)
	return protoreflect.ValueOfMessage(msg), err
}

type fieldBinding struct {
	hasValue HasValue
	field    protoreflect.FieldDescriptor
}

func (f fieldBinding) bind(msg protoreflect.Message) error {
	field := f.field
	val, err := f.hasValue.Get(msg.NewField(field))
	if err != nil {
		return err
	}

	kind := f.field.Kind()
	if !(field.IsList() ||
		field.IsMap() ||
		kind == protoreflect.MessageKind ||
		kind == protoreflect.GroupKind) {
		msg.Set(f.field, val)
	}

	return nil
}
