package bankv2

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return nil // Disable AutoCLI until https://github.com/cosmos/cosmos-sdk/issues/21682 is resolved.
	// return &autocliv1.ModuleOptions{
	// 	Query: &autocliv1.ServiceCommandDescriptor{
	// 		Service: "cosmos.bank.v2.Query",
	// 		RpcCommandOptions: []*autocliv1.RpcCommandOptions{
	// 			{
	// 				RpcMethod: "Params",
	// 				Use:       "params",
	// 				Short:     "Query current bank/v2 parameters",
	// 			},
	// 		},
	// 	},
	// 	Tx: &autocliv1.ServiceCommandDescriptor{
	// 		Service:              "cosmos.bank.v2.Msg",
	// 		EnhanceCustomCommand: true,
	// 		RpcCommandOptions: []*autocliv1.RpcCommandOptions{
	// 			{
	// 				RpcMethod:      "UpdateParams",
	// 				Use:            "update-params-proposal <params>",
	// 				Short:          "Submit a proposal to update bank module params. Note: the entire params must be provided.",
	// 				Example:        fmt.Sprintf(`%s tx bank update-params-proposal '{ }'`, version.AppName),
	// 				PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "params"}},
	// 				GovProposal:    true,
	// 			},
	// 		},
	// 	},
	// }
}
