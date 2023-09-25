package crisis

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	crisisv1beta1 "cosmossdk.io/api/cosmos/crisis/v1beta1"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service:              crisisv1beta1.Msg_ServiceDesc.ServiceName,
			EnhanceCustomCommand: true,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod:      "VerifyInvariant",
					Use:            "invariant-broken [module-name] [invariant-route]",
					Short:          "Submit proof that an invariant broken to halt the chain",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "to"}, {ProtoField: "amount"}},
					FlagOptions: map[string]*autocliv1.FlagOptions{
						"sender": {Name: "from", Shorthand: "f", Usage: "Address of the sender"},
					},
				},
			},
		},
	}
}
