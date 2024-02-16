package crisis

import (
	"fmt"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	crisisv1beta1 "cosmossdk.io/api/cosmos/crisis/v1beta1"

	"github.com/cosmos/cosmos-sdk/version"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service: crisisv1beta1.Msg_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "VerifyInvariant",
					Use:       "invariant-broken [module-name] [invariant-route] --from mykey",
					Short:     "Submit proof that an invariant broken",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "invariant_module_name"},
						{ProtoField: "invariant_route"},
					},
				},
				{
					RpcMethod:      "UpdateParams",
					Use:            "update-params-proposal [params]",
					Short:          "Submit a proposal to update crisis module params. Note: the entire params must be provided.",
					Example:        fmt.Sprintf(`%s tx crisis update-params-proposal '{ "constant_fee": {"denom": "stake", "amount": "1000"} }'`, version.AppName),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "params"}},
					GovProposal:    true,
				},
			},
		},
	}
}
