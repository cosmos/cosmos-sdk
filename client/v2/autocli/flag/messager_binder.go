package flag

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// SignerInfo contains information about the signer field.
// That field is special because it needs to be known for signing.
// This struct keeps track of the field name and whether it is a flag.
// IsFlag and PositionalArgIndex are mutually exclusive.
type SignerInfo struct {
	PositionalArgIndex int
	IsFlag             bool

	FieldName string
	FlagName  string // flag name (always set if IsFlag is true)
}

// MessageBinder binds multiple flags in a flag set to a protobuf message.
type MessageBinder struct {
	CobraArgs  cobra.PositionalArgs
	SignerInfo SignerInfo

	positionalFlagSet *pflag.FlagSet
	positionalArgs    []fieldBinding
	flagBindings      []fieldBinding
	messageType       protoreflect.MessageType

	hasVarargs        bool
	hasOptional       bool
	mandatoryArgUntil int
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
		if i == m.mandatoryArgUntil && m.hasVarargs {
			for _, v := range positionalArgs[i:] {
				if err := m.positionalFlagSet.Set(name, v); err != nil {
					return err
				}
			}
		} else {
			if err := m.positionalFlagSet.Set(name, positionalArgs[i]); err != nil {
				return err
			}
		}
	}

	msgName := msg.Descriptor().Name()
	// bind positional arg values to the message
	for _, arg := range m.positionalArgs {
		if msgName == arg.field.Parent().Name() {
			if err := arg.bind(msg); err != nil {
				return err
			}
		} else {
			if err := m.bindNestedField(msg, arg); err != nil {
				return err
			}
		}
	}

	// bind flag values to the message
	for _, binding := range m.flagBindings {
		if err := binding.bind(msg); err != nil {
			return err
		}
	}

	return nil
}

// bindNestedField binds a field value to a nested message field. It handles cases where the field
// belongs to a nested message type by recursively traversing the message structure.
func (m *MessageBinder) bindNestedField(msg protoreflect.Message, arg fieldBinding) error {
	name := protoreflect.Name(strings.ToLower(string(arg.field.Parent().Name())))
	innerMsgValue := msg.Get(msg.Descriptor().Fields().ByName(name))
	if !innerMsgValue.Message().IsValid() {
		msg.Set(msg.Descriptor().Fields().ByName(name), protoreflect.ValueOfMessage(innerMsgValue.Message().New()))
	}

	innerMsg := msg.Get(msg.Descriptor().Fields().ByName(name)).Message()
	argField := innerMsg.Descriptor().Fields().ByName(arg.field.Name())
	if argField.Kind() == protoreflect.MessageKind {
		return m.bindNestedField(innerMsg, arg)
	}

	return arg.bind(innerMsg)
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

	if field.IsMap() {
		return nil
	}

	if msg.IsValid() && val.IsValid() {
		msg.Set(f.field, val)
	}

	return nil
}
