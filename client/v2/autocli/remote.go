package autocli

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

	"cosmossdk.io/client/v2/autocli/flag"
	"cosmossdk.io/client/v2/autocli/internal/remote"
)

type RemoteCommandOptions struct {
	ConfigDir string
}

func (options RemoteCommandOptions) Command() (*cobra.Command, error) {
	configDir := options.ConfigDir
	if configDir == "" {
		userCfgDir, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}

		configDir = path.Join(userCfgDir, remote.DefaultDirName)
	}

	config, err := remote.LoadConfig(configDir)
	if err != nil {
		return nil, err
	}

	var initChain string
	cmd := &cobra.Command{
		Long: `To configure a new chain just run this command using the --init flag and the name of the chain as it's listed in the chain registry (https://github.com/cosmos/chain-registry).
If the chain is not listed in the chain registry, you can use any unique name.`,
		Example: "cosmcli --init cosmoshub",
		RunE: func(cmd *cobra.Command, args []string) error {
			if initChain != "" {
				return options.reconfigure(configDir, initChain, config)
			}

			return cmd.Help()
		},
	}

	cmd.Flags().StringVar(&initChain, "init", "", "Initialize a new chain with the specified name")

	for chain, chainConfig := range config.Chains {
		chainInfo, err := remote.LoadChainInfo(configDir, chain, chainConfig, false)
		if err != nil {
			fmt.Printf("Unable to load data for %s\n", chain)
			continue
		}

		appOpts := AppOptions{
			ModuleOptions: chainInfo.ModuleOptions,
		}

		builder := &Builder{
			Builder: flag.Builder{
				TypeResolver: &dynamicTypeResolver{
					files: chainInfo.FileDescriptorSet,
				},
				FileResolver: chainInfo.FileDescriptorSet,
			},
			GetClientConn: func(command *cobra.Command) (grpc.ClientConnInterface, error) {
				return chainInfo.OpenClient()
			},
			AddQueryConnFlags: func(command *cobra.Command) {},
		}

		var update bool
		var reconfig bool
		chainCmd := &cobra.Command{
			Use: chain,
			RunE: func(cmd *cobra.Command, args []string) error {
				if reconfig {
					return options.reconfigure(configDir, chain, config)
				} else if update {
					fmt.Printf("Updating autocli data for %s\n", chain)
					_, err := remote.LoadChainInfo(configDir, chain, chainConfig, true)
					return err
				} else {
					return cmd.Help()
				}
			},
		}
		chainCmd.Flags().BoolVar(&update, "update", false, "update the autocli data for the selected chain")
		chainCmd.Flags().BoolVar(&reconfig, "config", false, "re-configure the selected chain")

		err = appOpts.EnhanceRootCommandWithBuilder(chainCmd, builder)
		if err != nil {
			return nil, err
		}

		cmd.AddCommand(chainCmd)
	}

	return cmd, nil
}

func (options RemoteCommandOptions) reconfigure(configDir, chain string, config *remote.Config) error {
	fmt.Printf("Configuring %s\n", chain)
	endpoint, err := remote.SelectGRPCEndpoints(chain)
	if err != nil {
		return err
	}

	fmt.Printf("Selected: %s\n", endpoint)
	chainConfig := &remote.ChainConfig{
		GRPCEndpoints: []remote.GRPCEndpoint{
			{
				Endpoint: endpoint,
			},
		},
	}
	config.Chains[chain] = chainConfig

	err = remote.SaveConfig(configDir, config)
	if err != nil {
		return err
	}

	_, err = remote.LoadChainInfo(configDir, chain, chainConfig, true)
	return err
}

type dynamicTypeResolver struct {
	files *protoregistry.Files
}

var _ protoregistry.MessageTypeResolver = dynamicTypeResolver{}
var _ protoregistry.ExtensionTypeResolver = dynamicTypeResolver{}

func (d dynamicTypeResolver) FindMessageByName(message protoreflect.FullName) (protoreflect.MessageType, error) {
	desc, err := d.files.FindDescriptorByName(message)
	if err != nil {
		return nil, err
	}

	return dynamicpb.NewMessageType(desc.(protoreflect.MessageDescriptor)), nil
}

func (d dynamicTypeResolver) FindMessageByURL(url string) (protoreflect.MessageType, error) {
	if i := strings.LastIndexByte(url, '/'); i >= 0 {
		url = url[i+len("/"):]
	}

	desc, err := d.files.FindDescriptorByName(protoreflect.FullName(url))
	if err != nil {
		return nil, err
	}

	return dynamicpb.NewMessageType(desc.(protoreflect.MessageDescriptor)), nil
}

func (d dynamicTypeResolver) FindExtensionByName(field protoreflect.FullName) (protoreflect.ExtensionType, error) {
	desc, err := d.files.FindDescriptorByName(field)
	if err != nil {
		return nil, err
	}

	return dynamicpb.NewExtensionType(desc.(protoreflect.ExtensionTypeDescriptor)), nil
}

func (d dynamicTypeResolver) FindExtensionByNumber(message protoreflect.FullName, field protoreflect.FieldNumber) (protoreflect.ExtensionType, error) {
	panic("TODO")
}
