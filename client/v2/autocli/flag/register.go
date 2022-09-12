package flag

import (
	"context"
	"fmt"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// AddMessageFlags adds flags for each field in the message to the flag set.
func (b *Builder) AddMessageFlags(ctx context.Context, set *pflag.FlagSet, messageType protoreflect.MessageType, commandOptions *autocliv1.RpcCommandOptions, options Options) (*MessageBinder, error) {
	fields := messageType.Descriptor().Fields()
	numFields := fields.Len()
	handler := &MessageBinder{
		messageType: messageType,
	}

	isPositional := map[string]bool{}
	hasVarargs := false
	n := len(commandOptions.PositionalArgs)
	for i, arg := range commandOptions.PositionalArgs {
		field := fields.ByName(protoreflect.Name(arg.ProtoField))
		if field == nil {
			return nil, fmt.Errorf("can't find field %s on %s", arg.ProtoField, messageType.Descriptor().FullName())
		}

		if arg.Varargs {
			if i != n-1 {
				return nil, fmt.Errorf("varargs positional argument %s must be the last argument", arg.ProtoField)
			}

			hasVarargs = true
		}

		typ := b.resolveFlagType(field)
		if typ == nil {
			return nil, fmt.Errorf("can't resolve field %v", field)
		}

		value := typ.NewValue(ctx, b)

		handler.positionalArgs = append(handler.positionalArgs, struct {
			binder  Value
			field   protoreflect.FieldDescriptor
			varargs bool
		}{binder: value, field: field, varargs: arg.Varargs})
	}

	if hasVarargs {
		handler.CobraArgs = cobra.MinimumNArgs(n)
	} else {
		handler.CobraArgs = cobra.ExactArgs(n)
	}

	for i := 0; i < numFields; i++ {
		field := fields.Get(i)
		if isPositional[string(field.Name())] {
			continue
		}

		flagOpts := commandOptions.FlagOptions[string(field.Name())]
		binder, err := b.AddFieldFlag(ctx, set, field, flagOpts, options)
		if err != nil {
			return nil, err
		}

		handler.flagFieldPairs = append(handler.flagFieldPairs, struct {
			binder FieldValueBinder
			field  protoreflect.FieldDescriptor
		}{binder: binder, field: field})
	}
	return handler, nil
}

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
func (m MessageBinder) BuildMessage(positionalArgs []string) protoreflect.Message {
	msg := m.messageType.New()
	m.Bind(msg, positionalArgs)
	return msg
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
func (m MessageBinder) Get() protoreflect.Value {
	return protoreflect.ValueOfMessage(m.BuildMessage(nil))
}
