package circuit

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	circuitv1 "cosmossdk.io/api/cosmos/circuit/v1"
)

func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: circuitv1.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "DisabledList",
					Use:       "disabled-list",
					Short:     "Query for all disabled message types",
				},
				{
					RpcMethod:      "Account",
					Use:            "account [address]",
					Short:          "Query for account permissions",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "address"}},
				},
				{
					RpcMethod: "Accounts",
					Use:       "accounts",
					Short:     "Query for all account permissions",
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service: circuitv1.Msg_ServiceDesc.ServiceName,
		},
	}
}
