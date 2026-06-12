package runtime

import (
	appv1alpha1 "cosmossdk.io/api/cosmos/app/v1alpha1"
	"cosmossdk.io/core/autocli"

	reflectionpkg "github.com/cosmos/cosmos-sdk/runtime/services/reflection"
)

func (m appModule) AutoCLIOptions() *autocli.ModuleOptions {
	return &autocli.ModuleOptions{
		Query: &autocli.ServiceCommandDescriptor{
			Service: appv1alpha1.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocli.RpcCommandOptions{
				{
					RpcMethod: "Config",
					Short:     "Query the current app config",
				},
			},
			SubCommands: map[string]*autocli.ServiceCommandDescriptor{
				"autocli": {
					Service: "cosmos.autocli.v1.Query",
					RpcCommandOptions: []*autocli.RpcCommandOptions{
						{
							RpcMethod: "AppOptions",
							Short:     "Query the custom autocli options",
						},
					},
				},
				"reflection": {
					Service: reflectionpkg.ServiceName,
					RpcCommandOptions: []*autocli.RpcCommandOptions{
						{
							RpcMethod: "FileDescriptors",
							Short:     "Query the app's protobuf file descriptors",
						},
					},
				},
			},
		},
	}
}
