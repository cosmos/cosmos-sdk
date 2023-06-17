package distribution

import (
	"fmt"
	"strings"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	distirbuitonv1beta1 "cosmossdk.io/api/cosmos/distribution/v1beta1"

	"github.com/cosmos/cosmos-sdk/version"
)

var (
	FlagCommission       = "commission"
	FlagMaxMessagesPerTx = "max-msgs"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	exAccAddress, err := am.ac.BytesToString([]byte("A58856F0FD53BF058B4909A21AEC019107BA6A58856F0FD53BF058B4909A21AEC019107BA6"))
	if err != nil {
		panic(err)
	}

	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: distirbuitonv1beta1.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Query the current distribution parameters.",
				},
				{
					RpcMethod: "ValidatorDistributionInfo",
					Use:       "validator-distribution-info [validator]",
					Short:     "Query validator distribution info",
					Example: fmt.Sprintf(`Example: $ %s query distribution validator-distribution-info %s`,
						version.AppName, exAccAddress,
					),

					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "validator_address"},
					},
				},
				{
					RpcMethod: "ValidatorOutstandingRewards",
					Use:       "validator-outstanding-rewards [validator]",
					Short:     "Query distribution outstanding (un-withdrawn) rewards for a validator and all their delegations",
					Example:   fmt.Sprintf(`$ %s query distribution validator-outstanding-rewards [validator-address]`, version.AppName),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "validator_address"},
					},
				},
				{
					RpcMethod: "ValidatorCommission",
					Use:       "commission [validator]",
					Short:     "Query distribution validator commission",
					Example:   fmt.Sprintf(`$ %s query distribution commission [validator-address]`, version.AppName),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "validator_address"},
					},
				},
				{
					RpcMethod: "ValidatorSlashes",
					Use:       "slashes [validator] [start-height] [end-height]",
					Short:     "Query distribution validator slashes",
					Example:   fmt.Sprintf(`$ %s query distribution slashes [validator-address] 0 100`, version.AppName),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "validator_address"},
						{ProtoField: "starting_height"},
						{ProtoField: "ending_height"},
					},
				},
				{
					RpcMethod: "DelegationRewards",
					Use:       "rewards [delegator-addr] [validator-addr]",
					Short:     "Query all distribution delegator rewards or rewards from a particular validator",
					Long:      "Query all rewards earned by a delegator, optionally restrict to rewards from a single validator.",
					Example: strings.TrimSpace(
						fmt.Sprintf(`
$ %s query distribution rewards %s
$ %s query distribution rewards %s [validator-address]
`,
							version.AppName, exAccAddress, version.AppName, exAccAddress,
						),
					),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "delegator_address"},
						{ProtoField: "validator_address"},
					},
				},
				{
					RpcMethod: "CommunityPool",
					Use:       "community-pool",
					Short:     "Query the amount of coins in the community pool",
					Example:   fmt.Sprintf(`$ %s query distribution community-pool`, version.AppName),
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service: distirbuitonv1beta1.Msg_ServiceDesc.ServiceName,
		},
	}
}
