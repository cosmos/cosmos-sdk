package distribution

import (
	"fmt"
	"strings"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	distirbuitonv1beta1 "cosmossdk.io/api/cosmos/distribution/v1beta1"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
)

var (
	FlagCommission       = "commission"
	FlagMaxMessagesPerTx = "max-msgs"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	bech32PrefixValAddr := sdk.GetConfig().GetBech32ValidatorAddrPrefix()
	bech32PrefixAccAddr := sdk.GetConfig().GetBech32AccountAddrPrefix()

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
					Example: fmt.Sprintf(`Example: $ %s query distribution validator-distribution-info %s1lwjmdnks33xwnmfayc64ycprww49n33mtm92ne`,
						version.AppName, bech32PrefixValAddr,
					),

					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "validator_address"},
					},
				},
				{
					RpcMethod: "ValidatorOutstandingRewards",
					Use:       "validator-outstanding-rewards [validator]",
					Short:     "Query distribution outstanding (un-withdrawn) rewards for a validator and all their delegations",
					Example:   fmt.Sprintf(`$ %s query distribution validator-outstanding-rewards %s1lwjmdnks33xwnmfayc64ycprww49n33mtm92ne`, version.AppName, bech32PrefixValAddr),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "validator_address"},
					},
				},
				{
					RpcMethod: "ValidatorCommission",
					Use:       "commission [validator]",
					Short:     "Query distribution validator commission",
					Example:   fmt.Sprintf(`$ %s query distribution commission %s1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj`, version.AppName, bech32PrefixValAddr),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "validator_address"},
					},
				},
				{
					RpcMethod: "ValidatorSlashes",
					Use:       "slashes [validator] [start-height] [end-height]",
					Short:     "Query distribution validator slashes",
					Example:   fmt.Sprintf(`$ %s query distribution slashes %svaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj 0 100`, version.AppName, bech32PrefixValAddr),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "validator_address"},
						{ProtoField: "start_height"},
						{ProtoField: "end_height"},
					},
				},
				{
					RpcMethod: "DelegationRewards",
					Use:       "rewards [delegator-addr] [validator-addr]",
					Short:     "Query all distribution delegator rewards or rewards from a particular validator",
					Example: strings.TrimSpace(
						fmt.Sprintf(`Query all rewards earned by a delegator, optionally restrict to rewards from a single validator.

Examples:
$ %s query distribution rewards %s1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p
$ %s query distribution rewards %s1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p %s1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj
`,
							version.AppName, bech32PrefixAccAddr, version.AppName, bech32PrefixAccAddr, bech32PrefixValAddr,
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
