package internal

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/dynamicpb"

	"cosmossdk.io/client/v2/autocli"
	"cosmossdk.io/client/v2/autocli/flag"
)

var (
	flagInsecure = "insecure"
	flagUpdate   = "update"
	flagConfig   = "config"
)

func RootCommand() (*cobra.Command, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configDir := path.Join(homeDir, DefaultConfigDirName)
	config, err := LoadConfig(configDir)
	if err != nil {
		return nil, err
	}

	cmd := &cobra.Command{
		Use:   "hubl",
		Short: "Hubl is a CLI for interacting with Cosmos SDK chains",
		Long:  "Hubl is a CLI for interacting with Cosmos SDK chains",
	}

	// add commands
	commands, err := RemoteCommand(config, configDir)
	if err != nil {
		return nil, err
	}
	commands = append(commands, InitCommand(config, configDir))

	cmd.AddCommand(commands...)
	return cmd, nil
}

func InitCommand(config *Config, configDir string) *cobra.Command {
	var insecure bool

	cmd := &cobra.Command{
		Use:   "init [foochain]",
		Short: "Initialize a new chain",
		Long: `To configure a new chain just run this command using the --init flag and the name of the chain as it's listed in the chain registry (https://github.com/cosmos/chain-registry).
If the chain is not listed in the chain registry, you can use any unique name.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			chainName := strings.ToLower(args[0])

			return reconfigure(cmd, config, configDir, chainName)
		},
	}

	cmd.Flags().BoolVar(&insecure, flagInsecure, false, "allow setting up insecure gRPC connection")

	return cmd
}

func RemoteCommand(config *Config, configDir string) ([]*cobra.Command, error) {
	commands := []*cobra.Command{}

	for chain, chainConfig := range config.Chains {
		chain, chainConfig := chain, chainConfig

		// load chain info
		chainInfo := NewChainInfo(configDir, chain, chainConfig)
		if err := chainInfo.Load(false); err != nil {
			commands = append(commands, RemoteErrorCommand(config, configDir, chain, chainConfig, err))
			continue
		}

		appOpts := autocli.AppOptions{
			ModuleOptions: chainInfo.ModuleOptions,
		}

		builder := &autocli.Builder{
			Builder: flag.Builder{
				TypeResolver: &dynamicTypeResolver{chainInfo},
				FileResolver: chainInfo.ProtoFiles,
				GetClientConn: func() (grpc.ClientConnInterface, error) {
					return chainInfo.OpenClient()
				},
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
		)
		chainCmd := &cobra.Command{
			Use:   chain,
			Short: fmt.Sprintf("Commands for the %s chain", chain),
			RunE: func(cmd *cobra.Command, args []string) error {
				switch {
				case reconfig:
					return reconfigure(cmd, config, configDir, chain)
				case update:
					cmd.Printf("Updating autocli data for %s\n", chain)
					return chainInfo.Load(true)
				default:
					return cmd.Help()
				}
			},
		}
		chainCmd.Flags().BoolVar(&update, flagUpdate, false, "update the CLI commands for the selected chain (should be used after every chain upgrade)")
		chainCmd.Flags().BoolVar(&reconfig, flagConfig, false, "re-configure the selected chain (allows choosing a new gRPC endpoint and refreshes data")
		chainCmd.Flags().BoolVar(&insecure, flagInsecure, false, "allow re-configuring the selected chain using an insecure gRPC connection")

		if err := appOpts.EnhanceRootCommandWithBuilder(chainCmd, builder); err != nil {
			return nil, err
		}

		commands = append(commands, chainCmd)
	}

	return commands, nil
}

func RemoteErrorCommand(config *Config, configDir, chain string, chainConfig *ChainConfig, err error) *cobra.Command {
	cmd := &cobra.Command{
		Use:   chain,
		Short: "Unable to load data",
		Long:  "Unable to load data, reconfiguration needed.",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Printf("Error loading chain data for %s: %+v\n", chain, err)

			return reconfigure(cmd, config, configDir, chain)
		},
	}

	cmd.Flags().Bool(flagInsecure, chainConfig.GRPCEndpoints[0].Insecure, "allow setting up insecure gRPC connection")

	return cmd
}

func reconfigure(cmd *cobra.Command, config *Config, configDir, chain string) error {
	insecure, _ := cmd.Flags().GetBool(flagInsecure)

	cmd.Printf("Configuring %s\n", chain)
	endpoint, err := SelectGRPCEndpoints(chain)
	if err != nil {
		return err
	}

	cmd.Printf("%s endpoint selected\n", endpoint)
	chainConfig := &ChainConfig{
		GRPCEndpoints: []GRPCEndpoint{
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

	config.Chains[chain] = chainConfig
	if err := SaveConfig(configDir, config); err != nil {
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
