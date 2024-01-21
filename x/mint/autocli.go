package mint

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	mintv1beta1 "cosmossdk.io/api/cosmos/mint/v1beta1"
)

func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: mintv1beta1.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Inflation",
					Use:       "inflation",
					Short:     "Query the current minting inflation value",
				},
				{
					RpcMethod: "AnnualProvisions",
					Use:       "annual-provisions",
					Short:     "Query the current minting annual provisions value",
				},
				{
					RpcMethod: "GenesisTime",
					Use:       "genesis-time",
					Short:     "Query the genesis time",
				},
			},
		},
	}
}
