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
	handler.positionalFlagSet = pflag.NewFlagSet("positional", pflag.ContinueOnError)
	for i, arg := range commandOptions.PositionalArgs {
		isPositional[arg.ProtoField] = true

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

		hasValue, err := b.AddFieldFlag(
			ctx,
			handler.positionalFlagSet,
			field,
			&autocliv1.FlagOptions{Name: fmt.Sprintf("%d", i)},
			Options{},
		)
		if err != nil {
			return nil, err
		}

		handler.positionalArgs = append(handler.positionalArgs, fieldBinding{
			field:    field,
			hasValue: hasValue,
		})
	}

	if hasVarargs {
		handler.CobraArgs = cobra.MinimumNArgs(n)
		handler.hasVarargs = true
	} else {
		handler.CobraArgs = cobra.ExactArgs(n)
	}

	for i := 0; i < numFields; i++ {
		field := fields.Get(i)
		if isPositional[string(field.Name())] {
			continue
		}

		flagOpts := commandOptions.FlagOptions[string(field.Name())]
		hasValue, err := b.AddFieldFlag(ctx, set, field, flagOpts, options)
		if err != nil {
			return nil, err
		}

		handler.flagBindings = append(handler.flagBindings, fieldBinding{
			hasValue: hasValue,
			field:    field,
		})
	}
	return handler, nil
}
