package gov

import (
	"fmt"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	govv1 "cosmossdk.io/api/cosmos/gov/v1"
	govv1beta1 "cosmossdk.io/api/cosmos/gov/v1beta1"

	"github.com/cosmos/cosmos-sdk/version"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: govv1.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Query the parameters of the governance process",
					Long:      "Query the parameters of the governance process. Specify specific param types (voting|tallying|deposit) to filter results.",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "params_type", Optional: true},
					},
				},
				{
					RpcMethod: "Proposals",
					Use:       "proposals",
					Short:     "Query proposals with optional filters",
					Example:   fmt.Sprintf("%[1]s query gov proposals --depositor cosmos1...\n%[1]s query gov proposals --voter cosmos1...\n%[1]s query gov proposals --proposal-status (PROPOSAL_STATUS_DEPOSIT_PERIOD|PROPOSAL_STATUS_VOTING_PERIOD|PROPOSAL_STATUS_PASSED|PROPOSAL_STATUS_REJECTED|PROPOSAL_STATUS_FAILED)", version.AppName),
				},
				{
					RpcMethod: "Proposal",
					Use:       "proposal [proposal-id]",
					Short:     "Query details of a single proposal",
					Example:   fmt.Sprintf("%s query gov proposal 1", version.AppName),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "proposal_id"},
					},
				},
				{
					RpcMethod: "Vote",
					Use:       "vote [proposal-id] [voter-addr]",
					Short:     "Query details of a single vote",
					Example:   fmt.Sprintf("%s query gov vote 1 cosmos1...", version.AppName),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "proposal_id"},
						{ProtoField: "voter"},
					},
				},
				{
					RpcMethod: "Votes",
					Use:       "votes [proposal-id]",
					Short:     "Query votes of a single proposal",
					Example:   fmt.Sprintf("%s query gov votes 1", version.AppName),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "proposal_id"},
					},
				},
				{
					RpcMethod: "Deposit",
					Use:       "deposit [proposal-id] [depositer-addr]",
					Short:     "Query details of a deposit",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "proposal_id"},
						{ProtoField: "depositor"},
					},
				},
				{
					RpcMethod: "Deposits",
					Use:       "deposits [proposal-id]",
					Short:     "Query deposits on a proposal",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "proposal_id"},
					},
				},
				{
					RpcMethod: "TallyResult",
					Use:       "tally [proposal-id]",
					Short:     "Query the tally of a proposal vote",
					Example:   fmt.Sprintf("%s query gov tally 1", version.AppName),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "proposal_id"},
					},
				},
				{
					RpcMethod: "Constitution",
					Use:       "constitution",
					Short:     "Query the current chain constitution",
				},
			},
			// map v1beta1 as a sub-command
			SubCommands: map[string]*autocliv1.ServiceCommandDescriptor{
				"v1beta1": {
					Service: govv1beta1.Query_ServiceDesc.ServiceName,
				},
			},
			EnhanceCustomCommand: true, // We still have manual commands in gov that we want to keep
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service: govv1.Msg_ServiceDesc.ServiceName,
			// map v1beta1 as a sub-command
			SubCommands: map[string]*autocliv1.ServiceCommandDescriptor{
				"v1beta1": {
					Service: govv1beta1.Msg_ServiceDesc.ServiceName,
				},
			},
		},
	}
}
