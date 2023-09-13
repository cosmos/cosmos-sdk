package autocli

import (
	"fmt"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// BuildMsgCommand builds the msg commands for all the provided modules. If a custom command is provided for a
// module, this is used instead of any automatically generated CLI commands. This allows apps to a fully dynamic client
// with a more customized experience if a binary with custom commands is downloaded.
func (b *Builder) BuildMsgCommand(appOptions AppOptions, customCmds map[string]*cobra.Command) (*cobra.Command, error) {
	msgCmd := topLevelCmd("tx", "Transaction subcommands")
	if err := b.enhanceCommandCommon(msgCmd, msgCmdType, appOptions, customCmds); err != nil {
		return nil, err
	}

	return msgCmd, nil
}

// AddMsgServiceCommands adds a sub-command to the provided command for each
// method in the specified service and returns the command. This can be used in
// order to add auto-generated commands to an existing command.
func (b *Builder) AddMsgServiceCommands(cmd *cobra.Command, cmdDescriptor *autocliv1.ServiceCommandDescriptor) error {
	for cmdName, subCmdDescriptor := range cmdDescriptor.SubCommands {
		subCmd := findSubCommand(cmd, cmdName)
		if subCmd == nil {
			subCmd = topLevelCmd(cmdName, fmt.Sprintf("Tx commands for the %s service", subCmdDescriptor.Service))
		}

		// Add recursive sub-commands if there are any. This is used for nested services.
		if err := b.AddMsgServiceCommands(subCmd, subCmdDescriptor); err != nil {
			return err
		}

		cmd.AddCommand(subCmd)
	}

	if cmdDescriptor.Service == "" {
		// skip empty command descriptor
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
		methodName := protoreflect.Name(option.RpcMethod)
		// validate that methods exist
		if m := methods.ByName(methodName); m == nil {
			return fmt.Errorf("rpc method %q not found for service %q", methodName, service.FullName())
		}
		rpcOptMap[methodName] = option

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

		methodCmd, err := b.BuildMsgMethodCommand(methodDescriptor, methodOpts)
		if err != nil {
			return err
		}

		if findSubCommand(cmd, methodCmd.Name()) != nil {
			// do not overwrite existing commands
			// we do not display a warning because you may want to overwrite an autocli command
			continue
		}

		if methodCmd != nil {
			cmd.AddCommand(methodCmd)
		}
	}

	return nil
}

// BuildMsgMethodCommand returns a command that outputs the JSON representation of the message.
func (b *Builder) BuildMsgMethodCommand(descriptor protoreflect.MethodDescriptor, options *autocliv1.RpcCommandOptions) (*cobra.Command, error) {
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

		bz, err := jsonMarshalOptions.Marshal(input.Interface())
		if err != nil {
			return err
		}

		return b.outOrStdoutFormat(cmd, bz)
	})

	if b.AddTxConnFlags != nil {
		b.AddTxConnFlags(cmd)

		cmd.Flags().BoolP(flagNoIndent, "", false, "Do not indent JSON output")
	}

	return cmd, err
}
