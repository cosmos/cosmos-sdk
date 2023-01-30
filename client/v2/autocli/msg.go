package autocli

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	"cosmossdk.io/client/v2/internal/util"
	"fmt"
	"github.com/cockroachdb/errors"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

// BuildMsgCommand builds the msg commands for all the provided modules. If a custom command is provided for a
// module, this is used instead of any automatically generated CLI commands. This allows apps to a fully dynamic client
// with a more customized experience if a binary with custom commands is downloaded.
func (b *Builder) BuildMsgCommand(moduleOptions map[string]*autocliv1.ModuleOptions, customCmds map[string]*cobra.Command) (*cobra.Command, error) {
	msgCmd := topLevelCmd("tx", "Transaction subcommands")
	if err := b.EnhanceQueryCommand(msgCmd, moduleOptions, customCmds); err != nil {
		return nil, err
	}

	return msgCmd, nil
}

// EnhanceMsgCommand enhances the provided msg command with either generated commands based on the provided module
// options or the provided custom commands for each module. If the provided msg command already contains a command
// for a module, that command is not over-written by this method. This allows a graceful addition of autocli to
// automatically fill in missing commands.
func (b *Builder) EnhanceMsgCommand(msgCmd *cobra.Command, moduleOptions map[string]*autocliv1.ModuleOptions, customCmds map[string]*cobra.Command) error {
	allModuleNames := map[string]bool{}
	for moduleName := range moduleOptions {
		allModuleNames[moduleName] = true
	}
	for moduleName := range customCmds {
		allModuleNames[moduleName] = true
	}

	for moduleName := range allModuleNames {
		if existing := findSubCommand(msgCmd, moduleName); existing != nil {
			continue
		}

		if customCmd, ok := customCmds[moduleName]; ok {
			msgCmd.AddCommand(customCmd)
			continue
		}

		moduleOpt, ok := moduleOptions[moduleName]
		if !ok {
			continue
		}
		txCmdDesc := moduleOpt.Tx

		// if descriptor is nil, then there are no commands to add
		if txCmdDesc == nil {
			continue
		}
		cmd, err := b.BuildModuleMsgCommand(moduleName, moduleOpt.Tx)
		if err != nil {
			return err
		}
		msgCmd.AddCommand(cmd)
	}
	return nil
}

// BuildModuleMsgCommand builds the msg command for a single module.
func (b *Builder) BuildModuleMsgCommand(moduleName string, cmdDescriptor *autocliv1.ServiceCommandDescriptor) (*cobra.Command, error) {
	cmd := topLevelCmd(moduleName, fmt.Sprintf("Transations commands for the %s module", moduleName))

	err := b.AddMsgServiceCommands(cmd, cmdDescriptor)

	return cmd, err
}

// AddMsgServiceCommands adds a sub-command to the provided command for each
// method in the specified service and returns the command. This can be used in
// order to add auto-generated commands to an existing command.
func (b *Builder) AddMsgServiceCommands(cmd *cobra.Command, cmdDescriptor *autocliv1.ServiceCommandDescriptor) error {
	for cmdName, subCmdDescriptor := range cmdDescriptor.SubCommands {
		subCmd := topLevelCmd(cmdName, fmt.Sprintf("Querying commands for the %s service", subCmdDescriptor.Service))
		// Add recursive sub-commands if there are any. This is used for nested services.
		err := b.AddMsgServiceCommands(subCmd, subCmdDescriptor)
		if err != nil {
			return err
		}
		cmd.AddCommand(subCmd)
	}
	// skip empty command descriptor
	if cmdDescriptor.Service == "" {
		return nil
	}

	resolver := b.FileResolver
	if b.FileResolver == nil {
		resolver = protoregistry.GlobalFiles
	}

	descriptor, err := resolver.FindDescriptorByName(protoreflect.FullName(cmdDescriptor.Service))
	if err != nil {
		return errors.Errorf("can't find service %s: %v", cmdDescriptor.Service, err)
	}
	service := descriptor.(protoreflect.ServiceDescriptor)
	methods := service.Methods()

	rpcOptMap := map[protoreflect.Name]*autocliv1.RpcCommandOptions{}
	for _, option := range cmdDescriptor.RpcCommandOptions {
		methodName := protoreflect.Name(option.RpcMethod)
		// validate that methods exist
		if m := methods.ByName(methodName); m == nil {
			return fmt.Errorf("rpc method %q not found for service %q", methodName, service.FullName())

		}
		rpcOptMap[methodName] = option

	}

	methodsLength := methods.Len()
	for i := 0; i < methodsLength; i++ {
		methodDescriptor := methods.Get(i)
		methodOpts := rpcOptMap[methodDescriptor.Name()]
		methodCmd, err := b.BuildMsgMethodCommand(methodDescriptor, methodOpts)
		if err != nil {
			return err
		}
		if methodCmd != nil {
			cmd.AddCommand(methodCmd)
		}
	}

	return nil
}

func (b *Builder) BuildMsgMethodCommand(descriptor protoreflect.MethodDescriptor, options *autocliv1.RpcCommandOptions) (*cobra.Command, error) {
	if options == nil {
		options = &autocliv1.RpcCommandOptions{}
	}

	if options.Skip {
		return nil, nil
	}

	long := options.Long
	if long == "" {
		long = util.DescriptorDocs(descriptor)
	}

	inputType := util.ResolveMessageType(b.TypeResolver, descriptor.Input())
	outputType := util.ResolveMessageType(b.TypeResolver, descriptor.Output())

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

	jsonMarshalOptions := protojson.MarshalOptions{
		Indent:          "  ",
		UseProtoNames:   true,
		UseEnumNumbers:  false,
		EmitUnpopulated: true,
		Resolver:        b.TypeResolver,
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		input, err := binder.BuildMessage(args)
		if err != nil {
			return err
		}
		msg := input.Interface().(sdk.Msg)
		clientCtx, err := client.GetClientTxContext(cmd)
		if err != nil {
			return err
		}
		output := outputType.New()
		bz, err := jsonMarshalOptions.Marshal(output.Interface())
		if err != nil {
			return err
		}

		_, err = fmt.Fprintln(cmd.OutOrStdout(), string(bz))
		return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
	}

	return cmd, nil

}
