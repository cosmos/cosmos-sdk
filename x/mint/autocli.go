package mint

import (
	"fmt"

	mintv1beta1 "cosmossdk.io/api/cosmos/mint/v1beta1"
	autocli "cosmossdk.io/core/autocli"

	"github.com/cosmos/cosmos-sdk/version"
)

func (am AppModule) AutoCLIOptions() *autocli.ModuleOptions {
	return &autocli.ModuleOptions{
		Query: &autocli.ServiceCommandDescriptor{
			Service: mintv1beta1.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocli.RpcCommandOptions{
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Query the current minting parameters",
				},
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
			},
		},
		Tx: &autocli.ServiceCommandDescriptor{
			Service: mintv1beta1.Msg_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocli.RpcCommandOptions{
				{
					RpcMethod:      "UpdateParams",
					Use:            "update-params-proposal [params]",
					Short:          "Submit a proposal to update mint module params. Note: the entire params must be provided.",
					Long:           fmt.Sprintf("Submit a proposal to update mint module params. Note: the entire params must be provided.\n See the fields to fill in by running `%s query mint params --output json`", version.AppName),
					Example:        fmt.Sprintf(`%s tx mint update-params-proposal '{ "mint_denom": "stake", ... }'`, version.AppName),
					PositionalArgs: []*autocli.PositionalArgDescriptor{{ProtoField: "params"}},
					GovProposal:    true,
				},
			},
		},
	}
}
