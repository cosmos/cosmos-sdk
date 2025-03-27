package protocolpool

import (
	"fmt"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	poolv1 "cosmossdk.io/api/cosmos/protocolpool/v1"

	"github.com/cosmos/cosmos-sdk/version"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: poolv1.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "CommunityPool",
					Use:       "community-pool",
					Short:     "Query the amount of coins in the community pool",
					Example:   fmt.Sprintf(`%s query protocolpool community-pool`, version.AppName),
				},
				{
					RpcMethod: "ContinuousFunds",
					Use:       "continuous-funds",
					Short:     "Query all continuous funds",
					Example:   fmt.Sprintf(`%s query protocolpool continuous-funds`, version.AppName),
				},
				{
					RpcMethod:      "ContinuousFund",
					Use:            "continuous-fund <recipient>",
					Short:          "Query a continuous fund by its recipient address",
					Example:        fmt.Sprintf(`%s query protocolpool continuous-fund cosmos1...`, version.AppName),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "recipient"}},
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service: poolv1.Msg_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod:      "FundCommunityPool",
					Use:            "fund-community-pool <amount>",
					Short:          "Funds the community pool with the specified amount",
					Example:        fmt.Sprintf(`%s tx protocolpool fund-community-pool 100uatom --from mykey`, version.AppName),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "amount"}},
				},
				{
					RpcMethod: "CreateContinuousFund",
					Use:       "create-continuous-fund <recipient> <percentage> <expiry>",
					Short:     "Create continuous fund for a recipient with optional expiry",
					Example:   fmt.Sprintf(`%s tx protocolpool create-continuous-fund cosmos1... 0.2 2023-11-31T12:34:56.789Z --from mykey`, version.AppName),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "recipient"},
						{ProtoField: "percentage"},
						{ProtoField: "expiry", Optional: true},
					},
					GovProposal: true,
				},
				{
					RpcMethod: "CancelContinuousFund",
					Use:       "cancel-continuous-fund <recipient>",
					Short:     "Cancel continuous fund for a specific recipient",
					Example:   fmt.Sprintf(`%s tx protocolpool cancel-continuous-fund cosmos1... --from mykey`, version.AppName),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "recipient"},
					},
					GovProposal: true,
				},
				{
					RpcMethod:      "UpdateParams",
					Use:            "update-params-proposal <params>",
					Short:          "Submit a proposal to update protocolpool module params. Note: the entire params must be provided.",
					Example:        fmt.Sprintf(`%s tx protocolpool update-params-proposal '{ "enabled_distribution_denoms": ["stake", "foo"] }'`, version.AppName),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "params"}},
					GovProposal:    true,
				},
			},
		},
	}
}
