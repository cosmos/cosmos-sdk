package autocli

import (
	"fmt"
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
	cmd := &cobra.Command{
		RunE: func(cmd *cobra.Command, args []string) error {
			endpoint, err := remote.SelectGRPCEndpoints(args[0])
			if err != nil {
				return err
			}

			fmt.Printf("Selected: %v", endpoint)
			return nil
		},
	}

	config, err := remote.LoadConfig(options.ConfigDir)
	if err != nil {
		return nil, err
	}

	for chain, chainConfig := range config.Chains {
		chainInfo, err := remote.LoadChainInfo(chain, chainConfig, false)
		if err != nil {
			return nil, err
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
				return chainInfo.GRPCClient, nil
			},
			AddQueryConnFlags: func(command *cobra.Command) {},
		}

		chainCmd := &cobra.Command{Use: chain}
		err = appOpts.EnhanceRootCommandWithBuilder(chainCmd, builder)
		if err != nil {
			return nil, err
		}
	}

	return cmd, nil
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
