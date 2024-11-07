package autocli

import (
	"crypto/tls"
	"errors"
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
	"cosmossdk.io/client/v2/autocli/print"
	"cosmossdk.io/client/v2/internal/flags"
	"cosmossdk.io/client/v2/internal/util"
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

	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		err := b.setFlagsFromConfig(cmd, args)
		if err != nil {
			return err
		}

		k, err := keyring.NewKeyringFromFlags(cmd.Flags(), b.AddressCodec, cmd.InOrStdin(), b.Cdc)
		b.SetKeyring(k) // global flag keyring must be set on PreRunE.

		return err
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		ctx = cmd.Context()

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
			if hasModuleOptions { // check if we need to enhance the existing command
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
	output, _ := cmd.Flags().GetString(flags.FlagOutput)
	return print.NewPrinter(output, cmd.OutOrStdout()).PrintBytes(out)
}

func (b *Builder) setFlagsFromConfig(cmd *cobra.Command, args []string) error {
	conf, err := config.CreateClientConfigFromFlags(cmd.Flags())
	if err != nil {
		return err
	}

	if cmd.Flags().Lookup("chain-id") != nil && !cmd.Flags().Changed("chain-id") {
		cmd.Flags().Set("chain-id", conf.ChainID)
	}

	if cmd.Flags().Lookup("keyring-backend") != nil && !cmd.Flags().Changed("keyring-backend") {
		cmd.Flags().Set("keyring-backend", conf.KeyringBackend)
	}

	if cmd.Flags().Lookup("from") != nil && !cmd.Flags().Changed("from") {
		cmd.Flags().Set("from", conf.KeyringDefaultKeyName)
	}

	if cmd.Flags().Lookup("output") != nil && !cmd.Flags().Changed("output") {
		cmd.Flags().Set("output", conf.Output)
	}

	if cmd.Flags().Lookup("node") != nil && !cmd.Flags().Changed("node") {
		cmd.Flags().Set("node", conf.Node)
	}

	if cmd.Flags().Lookup("broadcast-mode") != nil && !cmd.Flags().Changed("broadcast-mode") {
		cmd.Flags().Set("broadcast-mode", conf.BroadcastMode)
	}

	if cmd.Flags().Lookup("grpc-addr") != nil && !cmd.Flags().Changed("grpc-addr") {
		cmd.Flags().Set("grpc-addr", conf.GRPC.Address)
	}

	if cmd.Flags().Lookup("grpc-insecure") != nil && !cmd.Flags().Changed("grpc-insecure") {
		cmd.Flags().Set("grpc-insecure", strconv.FormatBool(conf.GRPC.Insecure))
	}

	return nil
}

// TODO: godoc
func (b *Builder) getQueryClientConn(cmd *cobra.Command) (*grpc.ClientConn, error) {
	if cmd.Flags().Lookup("grpc-insecure") == nil || cmd.Flags().Lookup("grpc-addr") == nil {
		return nil, errors.New("grpc-insecure and grpc-addr flags are required")
	}

	creds := grpcinsecure.NewCredentials()
	insecure, err := cmd.Flags().GetBool("grpc-insecure")
	if err != nil {
		return nil, err
	}
	if !insecure {
		creds = credentials.NewTLS(&tls.Config{MinVersion: tls.VersionTLS12})
	}

	addr, err := cmd.Flags().GetString("grpc-addr")
	if err != nil {
		return nil, err
	}

	if addr == "" {
		// TODO: fall to default by querying state via abci query.
		return nil, errors.New("grpc-addr flag must be set")
	}

	return grpc.NewClient(addr, []grpc.DialOption{grpc.WithTransportCredentials(creds)}...)
}
