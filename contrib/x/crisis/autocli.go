package crisis

import (
	autocli "cosmossdk.io/core/autocli"

	crisisv1beta1 "github.com/cosmos/cosmos-sdk/contrib/api/cosmos/crisis/v1beta1"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocli.ModuleOptions {
	return &autocli.ModuleOptions{
		Tx: &autocli.ServiceCommandDescriptor{
			Service: crisisv1beta1.Msg_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocli.RpcCommandOptions{
				{
					RpcMethod: "VerifyInvariant",
					Use:       "invariant-broken [module-name] [invariant-route] --from mykey",
					Short:     "Submit proof that an invariant is broken",
					PositionalArgs: []*autocli.PositionalArgDescriptor{
						{ProtoField: "invariant_module_name"},
						{ProtoField: "invariant_route"},
					},
				},
				{
					RpcMethod: "UpdateParams",
					Skip:      true, // Crisis is deprecated.
				},
			},
		},
	}
}
