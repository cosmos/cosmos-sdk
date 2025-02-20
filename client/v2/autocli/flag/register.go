package flag

import (
	"context"
	"fmt"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func (b *Builder) AddMessageFlags(ctx context.Context, flagSet *pflag.FlagSet, messageType protoreflect.MessageType, commandOptions *autocliv1.RpcCommandOptions) (*MessageBinder, error) {
	return b.addMessageFlags(ctx, flagSet, messageType, commandOptions, namingOptions{})
}

// AddMessageFlags adds flags for each field in the message to the flag set.
func (b *Builder) addMessageFlags(ctx context.Context, flagSet *pflag.FlagSet, messageType protoreflect.MessageType, commandOptions *autocliv1.RpcCommandOptions, options namingOptions) (*MessageBinder, error) {
	fields := messageType.Descriptor().Fields()
	numFields := fields.Len()
	handler := &MessageBinder{
		messageType: messageType,
	}

	isPositional := map[string]bool{}
	hasVarargs := false
	n := len(commandOptions.PositionalArgs)
	// positional args are also parsed using a FlagSet so that we can reuse all the same parsers
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

		_, hasValue, err := b.addFieldFlag(
			ctx,
			handler.positionalFlagSet,
			field,
			&autocliv1.FlagOptions{Name: fmt.Sprintf("%d", i)},
			namingOptions{},
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

	// validate flag options
	for name := range commandOptions.FlagOptions {
		if fields.ByName(protoreflect.Name(name)) == nil {
			return nil, fmt.Errorf("can't find field %s on %s specified as a flag", name, messageType.Descriptor().FullName())
		}
	}

	flagOptsByFlagName := map[string]*autocliv1.FlagOptions{}
	for i := 0; i < numFields; i++ {
		field := fields.Get(i)
		if isPositional[string(field.Name())] {
			continue
		}

		flagOpts := commandOptions.FlagOptions[string(field.Name())]
		name, hasValue, err := b.addFieldFlag(ctx, flagSet, field, flagOpts, options)
		flagOptsByFlagName[name] = flagOpts
		if err != nil {
			return nil, err
		}

		handler.flagBindings = append(handler.flagBindings, fieldBinding{
			hasValue: hasValue,
			field:    field,
		})
	}

	flagSet.VisitAll(func(flag *pflag.Flag) {
		opts := flagOptsByFlagName[flag.Name]
		if opts != nil {
			// This is a bit of hacking around the pflag API, but
			// we need to set these options here using Flag.VisitAll because the flag
			// constructors that pflag gives us (StringP, Int32P, etc.) do not
			// actually return the *Flag instance
			flag.Deprecated = opts.Deprecated
			flag.ShorthandDeprecated = opts.ShorthandDeprecated
			flag.Hidden = opts.Hidden
			flag.NoOptDefVal = opts.NoOptDefaultValue
		}
	})

	return handler, nil
}
