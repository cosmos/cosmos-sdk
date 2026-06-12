package slashing

import (
	"fmt"

	slashingv1beta "cosmossdk.io/api/cosmos/slashing/v1beta1"
	autocli "cosmossdk.io/core/autocli"

	"github.com/cosmos/cosmos-sdk/version"
)

func (am AppModule) AutoCLIOptions() *autocli.ModuleOptions {
	return &autocli.ModuleOptions{
		Query: &autocli.ServiceCommandDescriptor{
			Service: slashingv1beta.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocli.RpcCommandOptions{
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Query the current slashing parameters",
				},
				{
					RpcMethod: "SigningInfo",
					Use:       "signing-info [validator-conspub/address]",
					Short:     "Query a validator's signing information",
					Long:      "Query a validator's signing information, with a pubkey ('<appd> comet show-validator') or a validator consensus address",
					Example:   fmt.Sprintf(`%s query slashing signing-info '{"@type":"/cosmos.crypto.ed25519.PubKey","key":"OauFcTKbN5Lx3fJL689cikXBqe+hcp6Y+x0rYUdR9Jk="}'`, version.AppName),
					PositionalArgs: []*autocli.PositionalArgDescriptor{
						{ProtoField: "cons_address"},
					},
				},
				{
					RpcMethod: "SigningInfos",
					Use:       "signing-infos",
					Short:     "Query signing information of all validators",
				},
			},
		},
		Tx: &autocli.ServiceCommandDescriptor{
			Service: slashingv1beta.Msg_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocli.RpcCommandOptions{
				{
					RpcMethod: "Unjail",
					Use:       "unjail",
					Short:     "Unjail a jailed validator",
					Example:   fmt.Sprintf("%s tx slashing unjail --from [validator]", version.AppName),
				},
				{
					RpcMethod:      "UpdateParams",
					Use:            "update-params-proposal [params]",
					Short:          "Submit a proposal to update slashing module params. Note: the entire params must be provided.",
					Long:           fmt.Sprintf("Submit a proposal to update slashing module params. Note: the entire params must be provided.\n See the fields to fill in by running `%s query slashing params --output json`", version.AppName),
					Example:        fmt.Sprintf(`%s tx slashing update-params-proposal '{ "signed_blocks_window": "100", ... }'`, version.AppName),
					PositionalArgs: []*autocli.PositionalArgDescriptor{{ProtoField: "params"}},
					GovProposal:    true,
				},
			},
		},
	}
}
