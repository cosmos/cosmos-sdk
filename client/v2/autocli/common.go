package autocli

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/exp/maps"
	"google.golang.org/protobuf/reflect/protoreflect"
	"sigs.k8s.io/yaml"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	"cosmossdk.io/client/v2/internal/util"

	"github.com/cosmos/cosmos-sdk/client/flags"
)

type cmdType int

const (
	queryCmdType cmdType = iota
	msgCmdType
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
		SilenceUsage: true,
		Use:          use,
		Long:         long,
		Short:        options.Short,
		Example:      options.Example,
		Aliases:      options.Alias,
		SuggestFor:   options.SuggestFor,
		Deprecated:   options.Deprecated,
		Version:      options.Version,
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
func (b *Builder) enhanceCommandCommon(
	cmd *cobra.Command,
	cmdType cmdType,
	appOptions AppOptions,
	customCmds map[string]*cobra.Command,
) error {
	moduleOptions := appOptions.ModuleOptions
	if len(moduleOptions) == 0 {
		moduleOptions = map[string]*autocliv1.ModuleOptions{}
		for name, module := range appOptions.Modules {
			if module, ok := module.(HasAutoCLIConfig); ok {
				moduleOptions[name] = module.AutoCLIOptions()
			}
		}
	}

	modules := append(maps.Keys(appOptions.Modules), maps.Keys(moduleOptions)...)
	for _, moduleName := range modules {
		modOpts, hasModuleOptions := moduleOptions[moduleName]

		// if we have an existing command skip adding one here
		if subCmd := findSubCommand(cmd, moduleName); subCmd != nil {
			if hasModuleOptions {
				if err := enhanceCustomCmd(b, subCmd, cmdType, modOpts); err != nil {
					return err
				}
			}

			continue
		}

		// if we have a custom command use that instead of generating one
		if custom, ok := customCmds[moduleName]; ok {
			if hasModuleOptions {
				if err := enhanceCustomCmd(b, custom, cmdType, modOpts); err != nil {
					return err
				}
			}

			cmd.AddCommand(custom)
			continue
		}

		// if we don't have module options, skip adding a command as we don't have anything to add
		if !hasModuleOptions {
			continue
		}

		switch cmdType {
		case queryCmdType:
			if err := enhanceQuery(b, moduleName, cmd, modOpts); err != nil {
				return err
			}
		case msgCmdType:
			if err := enhanceMsg(b, moduleName, cmd, modOpts); err != nil {
				return err
			}
		}
	}

	return nil
}

// enhanceQuery enhances the provided query command with the autocli commands for a module.
func enhanceQuery(builder *Builder, moduleName string, cmd *cobra.Command, modOpts *autocliv1.ModuleOptions) error {
	if queryCmdDesc := modOpts.Query; queryCmdDesc != nil {
		subCmd := topLevelCmd(moduleName, fmt.Sprintf("Querying commands for the %s module", moduleName))
		if err := builder.AddQueryServiceCommands(subCmd, queryCmdDesc); err != nil {
			return err
		}

		cmd.AddCommand(subCmd)
	}

	return nil
}

// enhanceMsg enhances the provided msg command with the autocli commands for a module.
func enhanceMsg(builder *Builder, moduleName string, cmd *cobra.Command, modOpts *autocliv1.ModuleOptions) error {
	if txCmdDesc := modOpts.Tx; txCmdDesc != nil {
		subCmd := topLevelCmd(moduleName, fmt.Sprintf("Transactions commands for the %s module", moduleName))
		if err := builder.AddMsgServiceCommands(subCmd, txCmdDesc); err != nil {
			return err
		}

		cmd.AddCommand(subCmd)
	}

	return nil
}

// enhanceCustomCmd enhances the provided custom query or msg command autocli commands for a module.
func enhanceCustomCmd(builder *Builder, cmd *cobra.Command, cmdType cmdType, modOpts *autocliv1.ModuleOptions) error {
	switch cmdType {
	case queryCmdType:
		if modOpts.Query != nil && modOpts.Query.EnhanceCustomCommand {
			if err := builder.AddQueryServiceCommands(cmd, modOpts.Query); err != nil {
				return err
			}
		}
	case msgCmdType:
		if modOpts.Tx != nil && modOpts.Tx.EnhanceCustomCommand {
			if err := builder.AddMsgServiceCommands(cmd, modOpts.Tx); err != nil {
				return err
			}
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
	if outputType != nil && outputType.Value.String() == flags.OutputFormatText {
		out, err = yaml.JSONToYAML(out)
		if err != nil {
			return err
		}
	}

	cmd.Println(strings.TrimSpace(string(out)))
	return nil
}
