package params

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

	paramsv1beta1 "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: paramsv1beta1.Query_serviceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Params",
					Use:       "subspace [subspace] [key]",
					Short:     "Query for raw parameters by subspace and key",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "subspace"},
						{ProtoField: "key"},
					},
				},
				{
					RpcMethod: "Subspaces",
					Use:       "subspaces",
					Short:     "Query for all registered subspaces and all keys for a subspace",
				},
			},
		},
	}
}
