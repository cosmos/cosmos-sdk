package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"cosmossdk.io/client/v2/cli/internal/remote"
)

type RemoteCommandOptions struct {
	ConfigDir string
}

func RemoteCommand(options RemoteCommandOptions) (*cobra.Command, error) {
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

	//config, err := remote.LoadConfig(options.ConfigDir)
	//if err != nil {
	//	return nil, err
	//}
	//
	//for chainName, chainConfig := range config.Chains {
	//	info, err := remote.LoadChainInfo(chainConfig)
	//	if err != nil {
	//		return nil, err
	//	}
	//
	//	builder := &Builder{
	//		Builder: flag.Builder{
	//			FileResolver: info.FileDescriptorSet,
	//		},
	//		GetClientConn: func(ctx context.Context) grpc.ClientConnInterface {
	//			return info.GRPCClient
	//		},
	//	}
	//
	//	appCmd, err := builder.BuildAppCommand(AppCommandOptions{
	//		Name:            chainName,
	//		ModuleOptions:   info.ModuleOptions,
	//		CustomQueryCmds: nil,
	//		CustomTxCmds:    nil,
	//	})
	//	if err != nil {
	//		return nil, err
	//	}
	//
	//	cmd.AddCommand(appCmd)
	//}

	return cmd, nil
}
