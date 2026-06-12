package upgrade

import (
	upgradev1beta1 "cosmossdk.io/api/cosmos/upgrade/v1beta1"
	autocli "cosmossdk.io/core/autocli"
)

func (am AppModule) AutoCLIOptions() *autocli.ModuleOptions {
	return &autocli.ModuleOptions{
		Query: &autocli.ServiceCommandDescriptor{
			Service: upgradev1beta1.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocli.RpcCommandOptions{
				{
					RpcMethod: "CurrentPlan",
					Use:       "plan",
					Short:     "Query the upgrade plan (if one exists)",
					Long:      "Gets the currently scheduled upgrade plan, if one exists",
				},
				{
					RpcMethod: "AppliedPlan",
					Use:       "applied [upgrade-name]",
					Short:     "Query the block header for height at which a completed upgrade was applied",
					Long:      "If upgrade-name was previously executed on the chain, this returns the header for the block at which it was applied. This helps a client determine which binary was valid over a given range of blocks, as well as more context to understand past migrations.",
					PositionalArgs: []*autocli.PositionalArgDescriptor{
						{ProtoField: "name"},
					},
				},
				{
					RpcMethod: "ModuleVersions",
					Use:       "module-versions [optional module_name]",
					Alias:     []string{"module_versions"},
					Short:     "Query the list of module versions",
					Long:      "Gets a list of module names and their respective consensus versions. Following the command with a specific module name will return only that module's information.",
					PositionalArgs: []*autocli.PositionalArgDescriptor{
						{ProtoField: "module_name", Optional: true},
					},
				},
				{
					RpcMethod: "Authority",
					Use:       "authority",
					Short:     "Get the upgrade authority address",
				},
				{
					RpcMethod: "UpgradedConsensusState",
					Skip:      true, // Skipping this command as the query is deprecated.
				},
			},
		},
		Tx: &autocli.ServiceCommandDescriptor{
			Service: upgradev1beta1.Msg_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocli.RpcCommandOptions{
				{
					RpcMethod:   "CancelUpgrade",
					Use:         "cancel-upgrade-proposal",
					Short:       "Submit a proposal to cancel a planned chain upgrade.",
					GovProposal: true,
				},
				{
					RpcMethod: "SoftwareUpgrade",
					Skip:      true, // skipped because authority gated
				},
			},
			EnhanceCustomCommand: true,
		},
	}
}
