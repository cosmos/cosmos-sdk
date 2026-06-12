package epochs

import (
	epochsv1beta1 "cosmossdk.io/api/cosmos/epochs/v1beta1"
	autocli "cosmossdk.io/core/autocli"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocli.ModuleOptions {
	return &autocli.ModuleOptions{
		Query: &autocli.ServiceCommandDescriptor{
			Service: epochsv1beta1.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocli.RpcCommandOptions{
				{
					RpcMethod: "EpochInfos",
					Use:       "epoch-infos",
					Short:     "Query running epoch infos",
				},
				{
					RpcMethod: "CurrentEpoch",
					Use:       "current-epoch [identifier]",
					Short:     "Query current epoch by specified identifier",
					PositionalArgs: []*autocli.PositionalArgDescriptor{
						{ProtoField: "identifier"},
					},
				},
			},
		},
	}
}
