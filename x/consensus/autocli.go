package consensus

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	consensusv1 "cosmossdk.io/api/cosmos/consensus/v1"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: consensusv1.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Query the current consensus parameters",
				},
			},
		},
		// Tx is purposely left empty, as the only tx is MsgUpdateParams which is gov gated.
	}
}
