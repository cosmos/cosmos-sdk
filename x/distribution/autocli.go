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
					Long: strings.TrimSpace(
						fmt.Sprintf(`Query validator distribution info.
Example:
$ %s query distribution validator-distribution-info %s1lwjmdnks33xwnmfayc64ycprww49n33mtm92ne
`,
							version.AppName, bech32PrefixValAddr,
						),
					),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "validator_address"},
					},
				},
				{
					RpcMethod: "ValidatorOutstandingRewards",
					Use:       "validator-outstanding-rewards [validator]",
					Short:     "Query distribution outstanding (un-withdrawn) rewards for a validator and all their delegations",
					Long: strings.TrimSpace(
						fmt.Sprintf(`Query distribution outstanding (un-withdrawn) rewards for a validator and all their delegations.

Example:
$ %s query distribution validator-outstanding-rewards %s1lwjmdnks33xwnmfayc64ycprww49n33mtm92ne
`,
							version.AppName, bech32PrefixValAddr,
						),
					),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "validator_address"},
					},
				},
				{
					RpcMethod: "ValidatorCommission",
					Use:       "commission [validator]",
					Short:     "Query distribution validator commission",
					Long: strings.TrimSpace(
						fmt.Sprintf(`Query validator commission rewards from delegators to that validator.

Example:
$ %s query distribution commission %s1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj
`,
							version.AppName, bech32PrefixValAddr,
						),
					),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "validator_address"},
					},
				},
				{
					RpcMethod: "ValidatorSlashes",
					Use:       "slashes [validator] [start-height] [end-height]",
					Short:     "Query distribution validator slashes",
					Long: strings.TrimSpace(
						fmt.Sprintf(`Query all slashes of a validator for a given block range.

Example:
$ %s query distribution slashes %svaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj 0 100
`,
							version.AppName, bech32PrefixValAddr,
						),
					),
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
					Long: strings.TrimSpace(
						fmt.Sprintf(`Query all rewards earned by a delegator, optionally restrict to rewards from a single validator.

Example:
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
					Long: strings.TrimSpace(
						fmt.Sprintf(`Query all coins in the community pool which is under Governance control.

Example:
$ %s query distribution community-pool
`,
							version.AppName,
						),
					),
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service: distirbuitonv1beta1.Msg_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "WithdrawDelegatorReward",
					Use:       "withdraw-rewards [validator-addr]",
					Short:     "Withdraw rewards from a given delegation address, and optionally withdraw validator commission if the delegation address given is a validator operator",
					Long: strings.TrimSpace(
						fmt.Sprintf(`Withdraw rewards from a given delegation address,
and optionally withdraw validator commission if the delegation address given is a validator operator.

Example:
$ %s tx distribution withdraw-rewards %s1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj --from mykey
$ %s tx distribution withdraw-rewards %s1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj --from mykey --commission
`,
							version.AppName, bech32PrefixValAddr, version.AppName, bech32PrefixValAddr,
						),
					),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						// client_address gets from context delegator_address
						{ProtoField: "validator_address"},
					},
					FlagOptions: map[string]*autocliv1.FlagOptions{
						// TODO: FlagComission is not in the proto file and just adds messages to the broadcas
						FlagCommission: {
							Name:         FlagCommission,
							Usage:        "Withdraw the validator's commission in addition to the rewards",
							DefaultValue: "false",
						},
					},
				},
				// TODO: Missing withdraw-all-rewards due to it being custom logic around WithdrawDelegatorReward
				{
					RpcMethod: "SetWithdrawAddress",
					Use:       "set-withdraw-addr [withdraw-addr]",
					Short:     "change the default withdraw address for rewards associated with an address",
					Long: strings.TrimSpace(
						fmt.Sprintf(`Set the withdraw address for rewards associated with a delegator address.

Example:
$ %s tx distribution set-withdraw-addr %s1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p --from mykey
`,
							version.AppName, bech32PrefixAccAddr,
						),
					),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						// TODO: delegator_address gets from context client_address
						{ProtoField: "withdraw_address"},
					},
				},
				{
					RpcMethod: "FundCommunityPool",
					Use:       "fund-community-pool [amount]",
					Short:     "Funds the community pool with the specified amount",
					Long: strings.TrimSpace(
						fmt.Sprintf(`Funds the community pool with the specified amount

Example:
$ %s tx distribution fund-community-pool 100uatom --from mykey
`,
							version.AppName,
						),
					),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						// TODO: depositor gets from context client_address
						{ProtoField: "amount", Varargs: true},
					},
				},
				{
					RpcMethod: "DepositValidatorRewardsPool",
					Use:       "fund-validator-rewards-pool [val_addr] [amount]",
					Short:     "Fund the validator rewards pool with the specified amount",
					Example: fmt.Sprintf(
						"%s tx distribution fund-validator-rewards-pool cosmosvaloper1x20lytyf6zkcrv5edpkfkn8sz578qg5sqfyqnp 100uatom --from mykey",
						version.AppName,
					),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						// TODO: depositor gets from context client_address
						{ProtoField: "validator_address"},
						{ProtoField: "amount", Varargs: true},
					},
				},
			},
		},
	}
}
