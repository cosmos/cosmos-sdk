package params

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	paramsv1beta1 "cosmossdk.io/api/cosmos/params/v1beta1"
)

func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: paramsv1beta1.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Subspace",
					Use:       "subspace [subspace] [key]",
					Short:     "Query for raw parameters by subspace and key",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "subspace"},
						{ProtoField: "key"},
					},
				},
			},
		},
	}
}
