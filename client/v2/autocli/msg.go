package autocli

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	"cosmossdk.io/client/v2/internal/util"
	"fmt"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// BuildMsgCommand builds the msg commands for all the provided modules. If a custom command is provided for a
// module, this is used instead of any automatically generated CLI commands. This allows apps to a fully dynamic client
// with a more customized experience if a binary with custom commands is downloaded.
func (b *Builder) BuildMsgCommand(moduleOptions map[string]*autocliv1.ModuleOptions, customCmds map[string]*cobra.Command) (*cobra.Command, error) {
	queryCmd := topLevelCmd("query", "Querying subcommands")
	queryCmd.Aliases = []string{"q"}
	if err := b.EnhanceQueryCommand(queryCmd, moduleOptions, customCmds); err != nil {
		return nil, err
	}

	return queryCmd, nil
}

// BuildModuleMsgCommand builds the msg command for a single module.
func (b *Builder) BuildModuleMsgCommand(moduleName string, cmdDescriptor *autocliv1.ServiceCommandDescriptor) (*cobra.Command, error) {
	cmd := topLevelCmd(moduleName, fmt.Sprintf("Transactions for the %s module", moduleName))

	err := b.AddQueryServiceCommands(cmd, cmdDescriptor)

	return cmd, err
}

// AddMsgServiceCommands adds a sub-command to the provided command for each
// method in the specified service and returns the command. This can be used in
// order to add auto-generated commands to an existing command.
func (b *Builder) AddMsgServiceCommands(cmd *cobra.Command, cmdDescriptor *autocliv1.ServiceCommandDescriptor) error {
	for cmdName, method := range cmdDescriptor.Methods {
		if err := b.AddMsgServiceCommand(cmd, method); err != nil {
			return err
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

	serviceDescriptor := descriptor.Parent().(protoreflect.ServiceDescriptor)

	long := options.Long
	if long == "" {
		long = util.DescriptorDocs(descriptor)
	}

	inputType := util.ResolveMessageType(b.TypeResolver, descriptor.Input())
	outputType := util.ResolveMessageType(b.TypeResolver, descriptor.Output())

	use := options.Use

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

		output := outputType.New()
		bz, err := jsonMarshalOptions.Marshal(output.Interface())
		if err != nil {
			return err
		}

		_, err = fmt.Fprintln(cmd.OutOrStdout(), string(bz))
		return err
	}

	return cmd, nil

}
