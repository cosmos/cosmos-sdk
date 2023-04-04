package autocli

import (
	"context"
	"fmt"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/reflect/protoreflect"
	"sigs.k8s.io/yaml"

	"cosmossdk.io/client/v2/internal/util"
)

func (b *Builder) buildMethodCommandCommon(descriptor protoreflect.MethodDescriptor, options *autocliv1.RpcCommandOptions, exec func(cmd *cobra.Command, input protoreflect.Message) error) (*cobra.Command, error) {
	if options == nil {
		// use the defaults
		options = &autocliv1.RpcCommandOptions{}
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

	cmd.SetContext(context.Background())
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

	return cmd, nil
}

// enhanceCommandCommon enhances the provided query or msg command with either generated commands based on the provided module
// options or the provided custom commands for each module. If the provided query command already contains a command
// for a module, that command is not over-written by this method. This allows a graceful addition of autocli to
// automatically fill in missing commands.
func (b *Builder) enhanceCommandCommon(cmd *cobra.Command, moduleOptions map[string]*autocliv1.ModuleOptions, customCmds map[string]*cobra.Command, buildModuleCommand func(*cobra.Command, *autocliv1.ModuleOptions, string) error) error {
	allModuleNames := map[string]bool{}
	for moduleName := range moduleOptions {
		allModuleNames[moduleName] = true
	}
	for moduleName := range customCmds {
		allModuleNames[moduleName] = true
	}

	for moduleName := range allModuleNames {
		// if we have an existing command skip adding one here
		if cmd.HasSubCommands() {
			if _, _, err := cmd.Find([]string{moduleName}); err == nil {
				// command already exists, skip
				continue
			}
		}

		// if we have a custom command use that instead of generating one
		if custom := customCmds[moduleName]; custom != nil {
			// custom commands get added lower down
			cmd.AddCommand(custom)
			continue
		}

		// check for autocli options
		modOpts := moduleOptions[moduleName]
		if modOpts == nil {
			continue
		}

		err := buildModuleCommand(cmd, modOpts, moduleName)
		if err != nil {
			return err
		}
	}

	return nil
}

// outOrStdoutFormat formats the output based on the output flag and writes it to the command's output stream.
func (b *Builder) outOrStdoutFormat(cmd *cobra.Command, out []byte) error {
	var err error
	outputType := cmd.Flag(flags.FlagOutput)
	// if the output type is text, convert the json to yaml
	// if output type is json or nil, default to json
	if outputType != nil && outputType.Value.String() == "text" {
		out, err = yaml.JSONToYAML(out)
		if err != nil {
			return err
		}
	}
	_, err = fmt.Fprintln(cmd.OutOrStdout(), string(out))

	return err
}
