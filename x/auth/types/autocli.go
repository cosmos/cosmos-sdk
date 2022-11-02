package types

import autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

var AutoCLIOptions = &autocliv1.ModuleOptions{
	Query: &autocliv1.ServiceCommandDescriptor{
		Service: _Query_serviceDesc.ServiceName,
		RpcCommandOptions: []*autocliv1.RpcCommandOptions{
			{
				RpcMethod:      "Account",
				Use:            "account [address]",
				Short:          "query account by address",
				PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "address"}},
			},
			{
				RpcMethod:      "AccountAddressByID",
				Use:            "address-by-id [id]",
				Short:          "query account address by account ID",
				PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "id"}},
			},
		},
	},
	Tx: &autocliv1.ServiceCommandDescriptor{
		Service: _Msg_serviceDesc.ServiceName,
	},
}
