package auth

import (
	authv1beta1 "cosmossdk.io/api/cosmos/auth/v1beta1"
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: authv1beta1.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod:      "Account",
					Use:            "account [address]",
					Short:          "query account by address",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "address"}},
				},
				{
					RpcMethod:      "AccountAddressByID",
					Use:            "address-by-acc-num [acc-num]",
					Short:          "query account address by account number",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "id"}},
				},
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Query the current auth parameters",
				},
			},
		},
		// Tx is purposely left empty, as the only tx is MsgUpdateParams which is gov gated.
	}
}
