package internal

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/dynamicpb"

	"cosmossdk.io/client/v2/autocli"
	"cosmossdk.io/client/v2/autocli/flag"
	"cosmossdk.io/tools/hubl/internal/config"
	"cosmossdk.io/tools/hubl/internal/flags"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
)

func InitCmd(config *config.Config, configDir string) *cobra.Command {
	var insecure bool

	cmd := &cobra.Command{
		Use:   "init <foochain>",
		Short: "Initialize a new chain",
		Long: `To configure a new chain, run this command using the --init flag and the name of the chain as it's listed in the chain registry (https://github.com/cosmos/chain-registry).
If the chain is not listed in the chain registry, you can use any unique name.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			chainName := strings.ToLower(args[0])
			return reconfigure(cmd, config, configDir, chainName)
		},
	}

	cmd.Flags().BoolVar(&insecure, flags.FlagInsecure, false, "allow setting up insecure gRPC connection")

	return cmd
}

func RemoteCommand(config *config.Config, configDir string) ([]*cobra.Command, error) {
	commands := []*cobra.Command{}

	for chain, chainConfig := range config.Chains {

		// load chain info
		chainInfo := NewChainInfo(configDir, chain, chainConfig)
		if err := chainInfo.Load(false); err != nil {
			commands = append(commands, RemoteErrorCommand(config, configDir, chain, chainConfig, err))
			continue
		}

		// add comet commands
		cometCmds := cmtservice.NewCometBFTCommands()
		chainInfo.ModuleOptions[cometCmds.Name()] = cometCmds.AutoCLIOptions()

		appOpts := autocli.AppOptions{
			ModuleOptions: chainInfo.ModuleOptions,
		}

		addressCodec, validatorAddressCodec, consensusAddressCodec, err := getAddressCodecFromConfig(config, chain)
		if err != nil {
			return nil, err
		}

		kr, err := getKeyring(chain)
		if err != nil {
			return nil, err
		}

		autoCLIKeyring, err := keyring.NewAutoCLIKeyring(kr)
		if err != nil {
			return nil, err
		}

		builder := &autocli.Builder{
			Builder: flag.Builder{
				TypeResolver:          &dynamicTypeResolver{chainInfo},
				FileResolver:          chainInfo.ProtoFiles,
				AddressCodec:          addressCodec,
				ValidatorAddressCodec: validatorAddressCodec,
				ConsensusAddressCodec: consensusAddressCodec,
				Keyring:               autoCLIKeyring,
			},
			GetClientConn: func(command *cobra.Command) (grpc.ClientConnInterface, error) {
				return chainInfo.OpenClient()
			},
			AddQueryConnFlags: func(command *cobra.Command) {},
		}

		var (
			update   bool
			reconfig bool
			insecure bool
			output   string
		)

		chainCmd := &cobra.Command{
			Use:   chain,
			Short: fmt.Sprintf("Commands for the %s chain", chain),
			RunE: func(cmd *cobra.Command, args []string) error {
				switch {
				case reconfig:
					return reconfigure(cmd, config, configDir, chain)
				case update:
					cmd.Printf("Updating AutoCLI data for %s\n", chain)
					return chainInfo.Load(true)
				default:
					return cmd.Help()
				}
			},
		}
		chainCmd.Flags().BoolVar(&update, flags.FlagUpdate, false, "update the CLI commands for the selected chain (should be used after every chain upgrade)")
		chainCmd.Flags().BoolVar(&reconfig, flags.FlagConfig, false, "re-configure the selected chain (allows choosing a new gRPC endpoint and refreshes data")
		chainCmd.Flags().BoolVar(&insecure, flags.FlagInsecure, false, "allow re-configuring the selected chain using an insecure gRPC connection")
		chainCmd.PersistentFlags().StringVar(&output, flags.FlagOutput, flags.OutputFormatJSON, fmt.Sprintf("output format (%s|%s)", flags.OutputFormatText, flags.OutputFormatJSON))

		// add chain specific keyring
		chainCmd.AddCommand(KeyringCmd(chainInfo.Chain))

		// add client context
		clientCtx := client.Context{}.WithKeyring(kr)
		chainCmd.SetContext(context.WithValue(context.Background(), client.ClientContextKey, &clientCtx))

		if err := appOpts.EnhanceRootCommandWithBuilder(chainCmd, builder); err != nil {
			// when enriching the command with autocli fails, we add a command that
			// will print the error and allow the user to reconfigure the chain instead
			chainCmd.RunE = func(cmd *cobra.Command, args []string) error {
				cmd.Printf("Error while loading AutoCLI data for %s: %+v\n", chain, err)
				cmd.Printf("Attempt to reconfigure the chain using the %s flag\n", flags.FlagConfig)
				if cmd.Flags().Changed(flags.FlagConfig) {
					return reconfigure(cmd, config, configDir, chain)
				}

				return nil
			}
		}

		commands = append(commands, chainCmd)
	}

	return commands, nil
}

func RemoteErrorCommand(cfg *config.Config, configDir, chain string, chainConfig *config.ChainConfig, err error) *cobra.Command {
	cmd := &cobra.Command{
		Use:   chain,
		Short: fmt.Sprintf("Unable to load %s data", chain),
		Long:  fmt.Sprintf("Unable to load %s data, reconfiguration needed.", chain),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Printf("Error loading chain data for %s: %+v\n", chain, err)
			return reconfigure(cmd, cfg, configDir, chain)
		},
	}

	cmd.Flags().Bool(flags.FlagInsecure, chainConfig.GRPCEndpoints[0].Insecure, "allow setting up insecure gRPC connection")

	return cmd
}

func reconfigure(cmd *cobra.Command, cfg *config.Config, configDir, chain string) error {
	insecure, _ := cmd.Flags().GetBool(flags.FlagInsecure)

	cmd.Printf("Configuring %s\n", chain)
	endpoint, err := SelectGRPCEndpoints(chain)
	if err != nil {
		return err
	}

	cmd.Printf("%s endpoint selected\n", endpoint)
	chainConfig := &config.ChainConfig{
		GRPCEndpoints: []config.GRPCEndpoint{
			{
				Endpoint: endpoint,
				Insecure: insecure,
			},
		},
	}

	chainInfo := NewChainInfo(configDir, chain, chainConfig)
	if err = chainInfo.Load(true); err != nil {
		return err
	}

	client, err := chainInfo.OpenClient()
	if err != nil {
		return err
	}

	addressPrefix, err := getAddressPrefix(context.Background(), client)
	if err != nil {
		return err
	}

	chainConfig.KeyringBackend = flags.DefaultKeyringBackend
	chainConfig.AddressPrefix = addressPrefix
	cfg.Chains[chain] = chainConfig

	if err := config.Save(configDir, cfg); err != nil {
		return err
	}

	cmd.Printf("Configuration saved to %s\n", configDir)
	return nil
}

type dynamicTypeResolver struct {
	*ChainInfo
}

var (
	_ protoregistry.MessageTypeResolver   = dynamicTypeResolver{}
	_ protoregistry.ExtensionTypeResolver = dynamicTypeResolver{}
)

func (d dynamicTypeResolver) FindMessageByName(message protoreflect.FullName) (protoreflect.MessageType, error) {
	desc, err := d.ProtoFiles.FindDescriptorByName(message)
	if err != nil {
		return nil, err
	}

	return dynamicpb.NewMessageType(desc.(protoreflect.MessageDescriptor)), nil
}

func (d dynamicTypeResolver) FindMessageByURL(url string) (protoreflect.MessageType, error) {
	if i := strings.LastIndexByte(url, '/'); i >= 0 {
		url = url[i+len("/"):]
	}

	return d.FindMessageByName(protoreflect.FullName(url))
}

func (d dynamicTypeResolver) FindExtensionByName(field protoreflect.FullName) (protoreflect.ExtensionType, error) {
	desc, err := d.ProtoFiles.FindDescriptorByName(field)
	if err != nil {
		return nil, err
	}

	return dynamicpb.NewExtensionType(desc.(protoreflect.ExtensionTypeDescriptor)), nil
}

func (d dynamicTypeResolver) FindExtensionByNumber(message protoreflect.FullName, field protoreflect.FieldNumber) (protoreflect.ExtensionType, error) {
	desc, err := d.ProtoFiles.FindDescriptorByName(message)
	if err != nil {
		return nil, err
	}

	messageDesc := desc.(protoreflect.MessageDescriptor)
	exts := messageDesc.Extensions()
	n := exts.Len()
	for i := 0; i < n; i++ {
		ext := exts.Get(i)
		if ext.Number() == field {
			return dynamicpb.NewExtensionType(ext), nil
		}
	}

	return nil, protoregistry.NotFound
}
