package autocli

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/reflect/protoreflect"

	"cosmossdk.io/client/v2/internal/util"
)

func (b *Builder) buildMethodCommandCommon(descriptor protoreflect.MethodDescriptor, options *autocliv1.RpcCommandOptions, exec func(cmd *cobra.Command, input protoreflect.Message) error) (*cobra.Command, error) {
	if options == nil {
		// use the defaults
		options = &autocliv1.RpcCommandOptions{}
	}

	if options.Skip {
		return nil, nil
	}

	long := options.Long
	if long == "" {
		long = util.DescriptorDocs(descriptor)
	}

	inputDesc := descriptor.Input()
	inputType := util.ResolveMessageType(b.TypeResolver, inputDesc)

	use := options.Use
	if use == "" {
		use = protoNameToCliName(descriptor.Name())
	}

	cmd := &cobra.Command{
		Use:        use,
		Long:       long,
		Short:      options.Short,
		Example:    options.Example,
		Aliases:    options.Alias,
		SuggestFor: options.SuggestFor,
		Deprecated: options.Deprecated,
		Version:    options.Version,
	}

	binder, err := b.AddMessageFlags(cmd.Context(), cmd.Flags(), inputType, options)
	if err != nil {
		return nil, err
	}

	cmd.Args = binder.CobraArgs

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		input, err := binder.BuildMessage(args)
		if err != nil {
			return err
		}

		return exec(cmd, input)
	}

	if b.AddQueryConnFlags != nil {
		b.AddQueryConnFlags(cmd)
	}

	return cmd, nil
}
