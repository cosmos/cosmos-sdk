package accounts

import (
	accountsv1 "cosmossdk.io/api/cosmos/accounts/v1"
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	"fmt"

	"github.com/cosmos/cosmos-sdk/version"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: accountsv1.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "AccountQuery",
					Use:       "query <account-address> <query-request-type-url> <json-message>",
					Short:     "Query account state",
					Example:   fmt.Sprintf(`%s q accounts query cosmos1uds6tz96dxfllz7tz3s3tm8tlg6x95g0mc2987sx6psjz98qlpss89sheu cosmos.accounts.defaults.multisig.v1.QueryProposal '{"proposal_id":1}`, version.AppName),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "target"},
						{ProtoField: "query_request_type_url"},
						{ProtoField: "json_message"},
					},
				},
			},
			EnhanceCustomCommand: true,
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service:              accountsv1.Msg_ServiceDesc.ServiceName,
			RpcCommandOptions:    []*autocliv1.RpcCommandOptions{},
			EnhanceCustomCommand: true,
		},
	}
}
