package autocli

import (
	"context"
	"crypto/tls"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	grpcinsecure "google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/reflect/protoreflect"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	"cosmossdk.io/client/v2/autocli/config"
	"cosmossdk.io/client/v2/autocli/keyring"
	"cosmossdk.io/client/v2/broadcast/comet"
	clientcontext "cosmossdk.io/client/v2/context"
	"cosmossdk.io/client/v2/internal/flags"
	"cosmossdk.io/client/v2/internal/print"
	"cosmossdk.io/client/v2/internal/util"

	"github.com/cosmos/cosmos-sdk/codec"
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

	short := options.Short
	if short == "" {
		short = fmt.Sprintf("Execute the %s RPC method", descriptor.Name())
	}

	inputDesc := descriptor.Input()
	inputType := util.ResolveMessageType(b.TypeResolver, inputDesc)

	use := options.Use
	if use == "" {
		use = protoNameToCliName(descriptor.Name())
	}

	cmd := &cobra.Command{
		SilenceUsage: false,
		Use:          use,
		Long:         options.Long,
		Short:        short,
		Example:      options.Example,
		Aliases:      options.Alias,
		SuggestFor:   options.SuggestFor,
		Deprecated:   options.Deprecated,
		Version:      options.Version,
	}

	// we need to use a pointer to the context as the correct context is set in the RunE function
	// however we need to set the flags before the RunE function is called
	ctx := cmd.Context()
	binder, err := b.AddMessageFlags(&ctx, cmd.Flags(), inputType, options)
	if err != nil {
		return nil, err
	}
	cmd.Args = binder.CobraArgs

	cmd.PreRunE = b.preRunE()

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		ctx, err = b.getContext(cmd)
		if err != nil {
			return err
		}

		input, err := binder.BuildMessage(args)
		if err != nil {
			return err
		}

		// signer related logic, triggers only when there is a signer defined
		if binder.SignerInfo.FieldName != "" {
			if binder.SignerInfo.IsFlag {
				// the client context uses the from flag to determine the signer.
				// this sets the signer flags to the from flag value if a custom signer flag is set.
				// marks the custom flag as required.
				if binder.SignerInfo.FlagName != flags.FlagFrom {
					if err := cmd.MarkFlagRequired(binder.SignerInfo.FlagName); err != nil {
						return err
					}

					if err := cmd.Flags().Set(flags.FlagFrom, cmd.Flag(binder.SignerInfo.FlagName).Value.String()); err != nil {
						return err
					}
				}
			} else {
				// if the signer is not a flag, it is a positional argument
				// we need to get the correct positional arguments
				if err := cmd.Flags().Set(flags.FlagFrom, args[binder.SignerInfo.PositionalArgIndex]); err != nil {
					return err
				}
			}
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
		moduleOptions = make(map[string]*autocliv1.ModuleOptions)
	}
	for name, module := range appOptions.Modules {
		if _, ok := moduleOptions[name]; !ok {
			if module, ok := module.(HasAutoCLIConfig); ok {
				moduleOptions[name] = module.AutoCLIOptions()
			} else {
				moduleOptions[name] = nil
			}
		}
	}

	for moduleName, modOpts := range moduleOptions {
		hasModuleOptions := modOpts != nil

		// if we have an existing command skip adding one here
		if subCmd := findSubCommand(cmd, moduleName); subCmd != nil {
			if hasModuleOptions { // check if we need to enhance the existing command
				if err := enhanceCustomCmd(b, subCmd, cmdType, modOpts); err != nil {
					return err
				}
			}

			continue
		}

		// if we have a custom command use that instead of generating one
		if custom, ok := customCmds[moduleName]; ok {
			// Custom may not be called the same as its module, so we need to have a separate check here
			if subCmd := findSubCommand(cmd, custom.Name()); subCmd != nil {
				if hasModuleOptions { // check if we need to enhance the existing command
					if err := enhanceCustomCmd(b, subCmd, cmdType, modOpts); err != nil {
						return err
					}
				}
				continue
			}
			if hasModuleOptions { // check if we need to enhance the new command
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
		short := queryCmdDesc.Short
		if short == "" {
			short = fmt.Sprintf("Querying commands for the %s module", moduleName)
		}
		subCmd := topLevelCmd(cmd.Context(), moduleName, short)
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
		short := txCmdDesc.Short
		if short == "" {
			short = fmt.Sprintf("Transactions commands for the %s module", moduleName)
		}
		subCmd := topLevelCmd(cmd.Context(), moduleName, short)
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
	p, err := print.NewPrinter(cmd)
	if err != nil {
		return err
	}
	return p.PrintBytes(out)
}

// getContext creates and returns a new context.Context with an autocli.Context value.
// It initializes a printer and, if necessary, a keyring based on command flags.
func (b *Builder) getContext(cmd *cobra.Command) (context.Context, error) {
	// if the command uses the keyring this must be set
	var (
		k   keyring.Keyring
		err error
	)
	if cmd.Flags().Lookup(flags.FlagKeyringDir) != nil && cmd.Flags().Lookup(flags.FlagKeyringBackend) != nil {
		k, err = keyring.NewKeyringFromFlags(cmd.Flags(), b.AddressCodec, cmd.InOrStdin(), b.Cdc)
		if err != nil {
			return nil, err
		}
	} else {
		k = keyring.NoKeyring{}
	}

	clientCtx := clientcontext.Context{
		Flags:                 cmd.Flags(),
		AddressCodec:          b.AddressCodec,
		ValidatorAddressCodec: b.ValidatorAddressCodec,
		ConsensusAddressCodec: b.ConsensusAddressCodec,
		Cdc:                   b.Cdc,
		Keyring:               k,
		EnabledSignModes:      b.EnabledSignModes,
	}

	return clientcontext.SetInContext(cmd.Context(), clientCtx), nil
}

// preRunE returns a function that sets flags from the configuration before running a command.
// It is used as a PreRunE hook for cobra commands to ensure flags are properly initialized
// from the configuration before command execution.
func (b *Builder) preRunE() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		err := b.setFlagsFromConfig(cmd)
		if err != nil {
			return err
		}

		return nil
	}
}

// setFlagsFromConfig sets command flags from the provided configuration.
// It only sets flags that haven't been explicitly changed by the user.
func (b *Builder) setFlagsFromConfig(cmd *cobra.Command) error {
	conf, err := config.CreateClientConfigFromFlags(cmd.Flags())
	if err != nil {
		return err
	}

	flagsToSet := map[string]string{
		flags.FlagChainID:        conf.ChainID,
		flags.FlagKeyringBackend: conf.KeyringBackend,
		flags.FlagFrom:           conf.KeyringDefaultKeyName,
		flags.FlagOutput:         conf.Output,
		flags.FlagNode:           conf.Node,
		flags.FlagBroadcastMode:  conf.BroadcastMode,
		flags.FlagGrpcAddress:    conf.GRPC.Address,
		flags.FlagGrpcInsecure:   strconv.FormatBool(conf.GRPC.Insecure),
	}

	for flagName, value := range flagsToSet {
		if flag := cmd.Flags().Lookup(flagName); flag != nil && !cmd.Flags().Changed(flagName) {
			if err := cmd.Flags().Set(flagName, value); err != nil {
				return err
			}
		}
	}

	return nil
}

// getQueryClientConn returns a function that creates a gRPC client connection based on command flags.
// It handles the creation of secure or insecure connections and falls back to a CometBFT broadcaster
// if no gRPC address is specified.
func getQueryClientConn(cdc codec.Codec) func(cmd *cobra.Command) (grpc.ClientConnInterface, error) {
	return func(cmd *cobra.Command) (grpc.ClientConnInterface, error) {
		var err error
		creds := grpcinsecure.NewCredentials()

		insecure := true
		if cmd.Flags().Lookup(flags.FlagGrpcInsecure) != nil {
			insecure, err = cmd.Flags().GetBool(flags.FlagGrpcInsecure)
			if err != nil {
				return nil, err
			}
		}
		if !insecure {
			creds = credentials.NewTLS(&tls.Config{MinVersion: tls.VersionTLS12})
		}

		var addr string
		if cmd.Flags().Lookup(flags.FlagGrpcAddress) != nil {
			addr, err = cmd.Flags().GetString(flags.FlagGrpcAddress)
			if err != nil {
				return nil, err
			}
		}
		if addr == "" {
			// if grpc-addr has not been set, use the default clientConn
			// TODO: default is comet
			node, err := cmd.Flags().GetString(flags.FlagNode)
			if err != nil {
				return nil, err
			}
			return comet.NewCometBFTBroadcaster(node, comet.BroadcastSync, cdc)
		}

		return grpc.NewClient(addr, []grpc.DialOption{grpc.WithTransportCredentials(creds)}...)
	}
}
