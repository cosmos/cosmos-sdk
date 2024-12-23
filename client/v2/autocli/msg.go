package autocli

import (
	"bufio"
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	"cosmossdk.io/client/v2/autocli/flag"
	"cosmossdk.io/client/v2/autocli/prompt"
	v2context "cosmossdk.io/client/v2/context"
	"cosmossdk.io/client/v2/internal/flags"
	"cosmossdk.io/client/v2/internal/governance"
	"cosmossdk.io/client/v2/internal/print"
	"cosmossdk.io/client/v2/internal/util"
	v2tx "cosmossdk.io/client/v2/tx"
	addresscodec "cosmossdk.io/core/address"
	"cosmossdk.io/core/transaction"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// BuildMsgCommand builds the msg commands for all the provided modules. If a custom command is provided for a
// module, this is used instead of any automatically generated CLI commands. This allows apps to a fully dynamic client
// with a more customized experience if a binary with custom commands is downloaded.
func (b *Builder) BuildMsgCommand(ctx context.Context, appOptions AppOptions, customCmds map[string]*cobra.Command) (*cobra.Command, error) {
	msgCmd := topLevelCmd(ctx, "tx", "Transaction subcommands")

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
			short := subCmdDescriptor.Short
			if short == "" {
				short = fmt.Sprintf("Tx commands for the %s service", subCmdDescriptor.Service)
			}
			subCmd = topLevelCmd(cmd.Context(), cmdName, short)
		}

		// Add recursive sub-commands if there are any. This is used for nested services.
		if err := b.AddMsgServiceCommands(subCmd, subCmdDescriptor); err != nil {
			return err
		}

		if !subCmdDescriptor.EnhanceCustomCommand {
			cmd.AddCommand(subCmd)
		}
	}

	if cmdDescriptor.Service == "" {
		// skip empty command descriptor
		return nil
	}

	descriptor, err := b.FileResolver.FindDescriptorByName(protoreflect.FullName(cmdDescriptor.Service))
	if err != nil {
		return fmt.Errorf("can't find service %s: %w", cmdDescriptor.Service, err)
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

		if !util.IsSupportedVersion(methodDescriptor) {
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

		cmd.AddCommand(methodCmd)
	}

	return nil
}

// BuildMsgMethodCommand returns a command that outputs the JSON representation of the message.
func (b *Builder) BuildMsgMethodCommand(descriptor protoreflect.MethodDescriptor, options *autocliv1.RpcCommandOptions) (*cobra.Command, error) {
	execFunc := func(cmd *cobra.Command, input protoreflect.Message) error {
		fd := input.Descriptor().Fields().ByName(protoreflect.Name(flag.GetSignerFieldName(input.Descriptor())))
		addressCodec := b.Builder.AddressCodec

		ctx, err := b.getContext(cmd)
		if err != nil {
			return err
		}
		// handle gov proposals commands
		skipProposal, _ := cmd.Flags().GetBool(flags.FlagNoProposal)
		if options.GovProposal && !skipProposal {
			return b.handleGovProposal(ctx, cmd, input, addressCodec, fd)
		}

		// set signer to signer field if empty
		if addr := input.Get(fd).String(); addr == "" {
			scalarType, ok := flag.GetScalarType(fd)
			if ok {
				// override address codec if validator or consensus address
				switch scalarType {
				case flag.ValidatorAddressStringScalarType:
					addressCodec = b.Builder.ValidatorAddressCodec
				case flag.ConsensusAddressStringScalarType:
					addressCodec = b.Builder.ConsensusAddressCodec
				}
			}

			signer, _, err := b.getFromAddress(ctx, cmd, addressCodec)
			if err != nil {
				return err
			}

			input.Set(fd, protoreflect.ValueOfString(signer))
		}

		// AutoCLI uses protov2 messages, while the SDK only supports proto v1 messages.
		// Here we use dynamicpb, to create a proto v1 compatible message.
		// The SDK codec will handle protov2 -> protov1 (marshal)
		msg := dynamicpb.NewMessage(input.Descriptor())
		proto.Merge(msg, input.Interface())

		return b.generateOrBroadcastTxWithV2(cmd, msg)
	}

	cmd, err := b.buildMethodCommandCommon(descriptor, options, execFunc)
	if err != nil {
		return nil, err
	}

	if b.AddTxConnFlags != nil {
		b.AddTxConnFlags(cmd)
	}

	// silence usage only for inner txs & queries commands
	cmd.SilenceUsage = true

	// set gov proposal flags if command is a gov proposal
	if options.GovProposal {
		governance.AddGovPropFlagsToCmd(cmd)
		cmd.Flags().Bool(flags.FlagNoProposal, false, "Skip gov proposal and submit a normal transaction")
	}

	return cmd, nil
}

// handleGovProposal sets the authority field of the message to the gov module address and creates a gov proposal.
func (b *Builder) handleGovProposal(
	ctx context.Context,
	cmd *cobra.Command,
	input protoreflect.Message,
	addressCodec addresscodec.Codec,
	fd protoreflect.FieldDescriptor,
) error {
	govAuthority := authtypes.NewModuleAddress(governance.ModuleName)
	authority, err := addressCodec.BytesToString(govAuthority.Bytes())
	if err != nil {
		return fmt.Errorf("failed to convert gov authority: %w", err)
	}
	input.Set(fd, protoreflect.ValueOfString(authority))

	signer, _, err := b.getFromAddress(ctx, cmd, addressCodec)
	if err != nil {
		return err
	}
	proposal, err := governance.ReadGovPropCmdFlags(signer, cmd.Flags())
	if err != nil {
		return err
	}

	// AutoCLI uses protov2 messages, while the SDK only supports proto v1 messages.
	// Here we use dynamicpb, to create a proto v1 compatible message.
	// The SDK codec will handle protov2 -> protov1 (marshal)
	msg := dynamicpb.NewMessage(input.Descriptor())
	proto.Merge(msg, input.Interface())

	if err := governance.SetGovMsgs(proposal, msg); err != nil {
		return fmt.Errorf("failed to set msg in proposal %w", err)
	}

	return b.generateOrBroadcastTxWithV2(cmd, proposal)
}

// getFromAddress retrieves the sender's address from the command flags and keyring.
// It returns the address string, raw address bytes, and any error encountered.
// The address is obtained from the --from flag and looked up in the keyring.
func (b *Builder) getFromAddress(ctx context.Context, cmd *cobra.Command, ac addresscodec.Codec) (string, []byte, error) {
	clientCtx, err := v2context.ClientContextFromGoContext(ctx)
	if err != nil {
		return "", nil, err
	}

	from, err := cmd.Flags().GetString(flags.FlagFrom)
	if err != nil {
		return "", nil, err
	}

	addr, err := ac.StringToBytes(from)
	if err == nil {
		return from, addr, nil
	}

	_, addrStr, addr, _, err := clientCtx.Keyring.KeyInfo(from)
	if err != nil {
		return "", nil, err
	}

	return addrStr, addr, nil
}

// generateOrBroadcastTxWithV2 generates or broadcasts a transaction with the provided messages using v2 transaction handling.
func (b *Builder) generateOrBroadcastTxWithV2(cmd *cobra.Command, msgs ...transaction.Msg) error {
	ctx, err := b.getContext(cmd)
	if err != nil {
		return err
	}

	cConn, err := b.GetClientConn(cmd)
	if err != nil {
		return err
	}

	var bz []byte
	genOnly, _ := cmd.Flags().GetBool(v2tx.FlagGenerateOnly)
	isDryRun, _ := cmd.Flags().GetBool(v2tx.FlagDryRun)
	if genOnly {
		bz, err = v2tx.GenerateOnly(ctx, cConn, msgs...)
	} else if isDryRun {
		bz, err = v2tx.DryRun(ctx, cConn, msgs...)
	} else {
		skipConfirm, _ := cmd.Flags().GetBool("yes")
		if skipConfirm {
			bz, err = v2tx.GenerateAndBroadcastTxCLI(ctx, cConn, msgs...)
		} else {
			bz, err = v2tx.GenerateAndBroadcastTxCLIWithPrompt(ctx, cConn, b.userConfirmation(cmd), msgs...)
		}
	}
	if err != nil {
		return err
	}

	output, _ := cmd.Flags().GetString(flags.FlagOutput)
	p := print.Printer{
		Output:       cmd.OutOrStdout(),
		OutputFormat: output,
	}

	return p.PrintBytes(bz)
}

// userConfirmation returns a function that prompts the user for confirmation
// before signing and broadcasting a transaction.
func (b *Builder) userConfirmation(cmd *cobra.Command) func([]byte) (bool, error) {
	format, _ := cmd.Flags().GetString(flags.FlagOutput)
	printer := print.Printer{
		Output:       cmd.OutOrStdout(),
		OutputFormat: format,
	}

	return func(bz []byte) (bool, error) {
		err := printer.PrintBytes(bz)
		if err != nil {
			return false, err
		}
		buf := bufio.NewReader(cmd.InOrStdin())

		err = printer.PrintString("confirm transaction before signing and broadcasting y/N: ")
		if err != nil {
			return false, err
		}

		ok, err := prompt.UserConfirmation(buf)
		if err != nil {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "error: %v\ncanceled transaction\n", err)
			return false, err
		}
		if !ok {
			_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "canceled transaction")
			return false, nil
		}

		return true, nil
	}
}
