package module

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	feegrantv1beta1 "cosmossdk.io/api/cosmos/feegrant/v1beta1"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: feegrantv1beta1.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod:      "Allowance",
					Use:            "grant [granter] [grantee]",
					Short:          "Query details of a single grant",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "Granter"}, {ProtoField: "Grantee"}},
				},
				{
					RpcMethod:      "Allowances",
					Use:            "grants-by-grantee [grantee]",
					Short:          "Query all grants of a grantee",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "Grantee"}, {ProtoField: "Pagination"}},
				},
				{
					RpcMethod:      "AllowancesByGranter",
					Use:            "grants-by-granter [granter]",
					Short:          "Query all grants by a granter",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "Granter"}, {ProtoField: "Pagination"}},
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service: feegrantv1beta1.Msg_ServiceDesc.ServiceName,
		},
	}
}
