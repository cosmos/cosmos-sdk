package autocli

import (
	"fmt"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"

	"cosmossdk.io/client/v2/internal/util"
)

// BuildQueryCommand builds the query commands for all the provided modules. If a custom command is provided for a
// module, this is used instead of any automatically generated CLI commands. This allows apps to a fully dynamic client
// with a more customized experience if a binary with custom commands is downloaded.
func (b *Builder) BuildQueryCommand(appOptions AppOptions, customCmds map[string]*cobra.Command) (*cobra.Command, error) {
	queryCmd := topLevelCmd("query", "Querying subcommands")
	queryCmd.Aliases = []string{"q"}

	if err := b.enhanceCommandCommon(queryCmd, queryCmdType, appOptions, customCmds); err != nil {
		return nil, err
	}

	return queryCmd, nil
}

// AddQueryServiceCommands adds a sub-command to the provided command for each
// method in the specified service and returns the command. This can be used in
// order to add auto-generated commands to an existing command.
func (b *Builder) AddQueryServiceCommands(cmd *cobra.Command, cmdDescriptor *autocliv1.ServiceCommandDescriptor) error {
	for cmdName, subCmdDesc := range cmdDescriptor.SubCommands {
		subCmd := findSubCommand(cmd, cmdName)
		if subCmd == nil {
			subCmd = topLevelCmd(cmdName, fmt.Sprintf("Querying commands for the %s service", subCmdDesc.Service))
		}

		if err := b.AddQueryServiceCommands(subCmd, subCmdDesc); err != nil {
			return err
		}

		cmd.AddCommand(subCmd)
	}

	// skip empty command descriptors
	if cmdDescriptor.Service == "" {
		return nil
	}

	descriptor, err := b.FileResolver.FindDescriptorByName(protoreflect.FullName(cmdDescriptor.Service))
	if err != nil {
		return errors.Errorf("can't find service %s: %v", cmdDescriptor.Service, err)
	}

	service := descriptor.(protoreflect.ServiceDescriptor)
	methods := service.Methods()

	rpcOptMap := map[protoreflect.Name]*autocliv1.RpcCommandOptions{}
	for _, option := range cmdDescriptor.RpcCommandOptions {
		name := protoreflect.Name(option.RpcMethod)
		rpcOptMap[name] = option
		// make sure method exists
		if m := methods.ByName(name); m == nil {
			return fmt.Errorf("rpc method %q not found for service %q", name, service.FullName())
		}
	}

	for i := 0; i < methods.Len(); i++ {
		methodDescriptor := methods.Get(i)
		methodOpts, ok := rpcOptMap[methodDescriptor.Name()]
		if !ok {
			methodOpts = &autocliv1.RpcCommandOptions{}
		}

		if methodOpts.Skip {
			continue
		}

		methodCmd, err := b.BuildQueryMethodCommand(methodDescriptor, methodOpts)
		if err != nil {
			return err
		}

		if findSubCommand(cmd, methodCmd.Name()) != nil {
			// do not overwrite existing commands
			// we do not display a warning because you may want to overwrite an autocli command
			continue
		}

		cmd.AddCommand(methodCmd)
	}

	return nil
}

// BuildQueryMethodCommand creates a gRPC query command for the given service method. This can be used to auto-generate
// just a single command for a single service rpc method.
func (b *Builder) BuildQueryMethodCommand(descriptor protoreflect.MethodDescriptor, options *autocliv1.RpcCommandOptions) (*cobra.Command, error) {
	getClientConn := b.GetClientConn
	serviceDescriptor := descriptor.Parent().(protoreflect.ServiceDescriptor)
	methodName := fmt.Sprintf("/%s/%s", serviceDescriptor.FullName(), descriptor.Name())
	outputType := util.ResolveMessageType(b.TypeResolver, descriptor.Output())
	jsonMarshalOptions := protojson.MarshalOptions{
		Indent:          "  ",
		UseProtoNames:   true,
		UseEnumNumbers:  false,
		EmitUnpopulated: true,
		Resolver:        b.TypeResolver,
	}

	cmd, err := b.buildMethodCommandCommon(descriptor, options, func(cmd *cobra.Command, input protoreflect.Message) error {
		if noIdent, _ := cmd.Flags().GetBool(flagNoIndent); noIdent {
			jsonMarshalOptions.Indent = ""
		}

		clientConn, err := getClientConn(cmd)
		if err != nil {
			return err
		}

		output := outputType.New()
		if err := clientConn.Invoke(cmd.Context(), methodName, input.Interface(), output.Interface()); err != nil {
			return err
		}

		bz, err := jsonMarshalOptions.Marshal(output.Interface())
		if err != nil {
			return fmt.Errorf("cannot marshal response %v: %w", output.Interface(), err)
		}

		return b.outOrStdoutFormat(cmd, bz)
	})
	if err != nil {
		return nil, err
	}

	if b.AddQueryConnFlags != nil {
		b.AddQueryConnFlags(cmd)

		cmd.Flags().BoolP(flagNoIndent, "", false, "Do not indent JSON output")
	}

	return cmd, nil
}
