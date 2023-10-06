package staking

import (
	"fmt"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	_ "cosmossdk.io/api/cosmos/crypto/ed25519" // register to that it shows up in protoregistry.GlobalTypes
	stakingv1beta "cosmossdk.io/api/cosmos/staking/v1beta1"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
)

func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: stakingv1beta.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Validators",
					Short:     "Query for all validators",
					Long:      "Query details about all validators on a network.",
				},
				{
					RpcMethod: "Validator",
					Use:       "validator [validator-addr]",
					Short:     "Query a validator",
					Long:      "Query details about an individual validator.",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "validator_addr"},
					},
				},
				{
					RpcMethod: "ValidatorDelegations",
					Use:       "delegations-to [validator-addr]",
					Short:     "Query all delegations made to one validator",
					Long:      "Query delegations on an individual validator.",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{
							ProtoField: "validator_addr",
						},
					},
				},
				{
					RpcMethod: "ValidatorUnbondingDelegations",
					Use:       "unbonding-delegations-from [validator-addr]",
					Short:     "Query all unbonding delegatations from a validator",
					Long:      "Query delegations that are unbonding _from_ a validator.",
					Example:   fmt.Sprintf("$ %s query staking unbonding-delegations-from [val-addr]", version.AppName),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "validator_addr"},
					},
				},
				{
					RpcMethod: "Delegation",
					Use:       "delegation [delegator-addr] [validator-addr]",
					Short:     "Query a delegation based on address and validator address",
					Long:      "Query delegations for an individual delegator on an individual validator",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "delegator_addr"},
						{ProtoField: "validator_addr"},
					},
				},
				{
					RpcMethod: "UnbondingDelegation",
					Use:       "unbonding-delegation [delegator-addr] [validator-addr]",
					Short:     "Query an unbonding-delegation record based on delegator and validator address",
					Long:      "Query unbonding delegations for an individual delegator on an individual validator.",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "delegator_addr"},
						{ProtoField: "validator_addr"},
					},
				},
				{
					RpcMethod: "DelegatorDelegations",
					Use:       "delegations [delegator-addr]",
					Short:     "Query all delegations made by one delegator",
					Long:      "Query delegations for an individual delegator on all validators.",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "delegator_addr"},
					},
				},
				{
					RpcMethod: "DelegatorValidators",
					Use:       "delegator-validators [delegator-addr]",
					Short:     "Query all validators info for given delegator address",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "delegator_addr"},
					},
				},
				{
					RpcMethod: "DelegatorValidator",
					Use:       "delegator-validator [delegator-addr] [validator-addr]",
					Short:     "Query validator info for given delegator validator pair",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "delegator_addr"},
						{ProtoField: "validator_addr"},
					},
				},
				{
					RpcMethod: "DelegatorUnbondingDelegations",
					Use:       "unbonding-delegations [delegator-addr]",
					Short:     "Query all unbonding-delegations records for one delegator",
					Long:      "Query unbonding delegations for an individual delegator.",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "delegator_addr"},
					},
				},
				{
					RpcMethod: "Redelegations",
					Use:       "redelegation [delegator-addr] [src-validator-addr] [dst-validator-addr]",
					Short:     "Query a redelegation record based on delegator and a source and destination validator address",
					Long:      "Query a redelegation record for an individual delegator between a source and destination validator.",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "delegator_addr"},
						{ProtoField: "src_validator_addr"},
						{ProtoField: "dst_validator_addr"},
					},
				},
				{
					RpcMethod: "HistoricalInfo",
					Use:       "historical-info [height]",
					Short:     "Query historical info at given height",
					Long:      "Query historical info at given height.",
					Example:   fmt.Sprintf("$ %s query staking historical-info 5", version.AppName),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "height"},
					},
				},
				{
					RpcMethod: "Pool",
					Use:       "pool",
					Short:     "Query the current staking pool values",
					Long:      "Query values for amounts stored in the staking pool.",
				},
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Query the current staking parameters information",
					Long:      "Query values set as staking parameters.",
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service: stakingv1beta.Msg_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod:      "Delegate",
					Use:            "delegate [validator-addr] [amount] --from [delegator_address]",
					Short:          "Delegate liquid tokens to a validator",
					Long:           "Delegate an amount of liquid coins to a validator from your wallet.",
					Example:        fmt.Sprintf("%s tx staking delegate cosmosvalopers1l2rsakp388kuv9k8qzq6lrm9taddae7fpx59wm 1000stake --from mykey", version.AppName),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "validator_address"}, {ProtoField: "amount"}},
				},
				{
					RpcMethod: "BeginRedelegate",
					Use:       "redelegate [src-validator-addr] [dst-validator-addr] [amount] --from [delegator]",
					Alias:     []string{"redelegate"},
					Short:     "Generate multisig signatures for transactions generated offline",
					Long: `Sign transactions created with the --generate-only flag that require multisig signatures.

Read one or more signatures from one or more [signature] file, generate a multisig signature compliant to the
multisig key [name], and attach the key name to the transaction read from [file].

If --signature-only flag is on, output a JSON representation
of only the generated signature.

If the --offline flag is on, the client will not reach out to an external node.
Account number or sequence number lookups are not performed so you must
set these parameters manually.

The current multisig implementation defaults to amino-json sign mode.
The SIGN_MODE_DIRECT sign mode is not supported.`,
					Example:        fmt.Sprintf(`%s tx multisign transaction.json k1k2k3 k1sig.json k2sig.json k3sig.json`, version.AppName),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "validator_src_address"}, {ProtoField: "validator_dst_address"}, {ProtoField: "amount"}},
				},
				{
					RpcMethod:      "Undelegate",
					Use:            "unbond [validator-addr] [amount] --from [delegator_address]",
					Short:          "Unbond shares from a validator",
					Long:           "Unbond an amount of bonded shares from a validator.",
					Example:        fmt.Sprintf(`%s tx staking unbond %s1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj 100stake --from mykey`, version.AppName, sdk.GetConfig().GetBech32ValidatorAddrPrefix()),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "validator_address"}, {ProtoField: "amount"}},
				},
				{
					RpcMethod:      "CancelUnbondingDelegation",
					Use:            "cancel-unbond [validator-addr] [amount] [creation-height]",
					Short:          "Cancel unbonding delegation and delegate back to the validator",
					Example:        fmt.Sprintf(`%s tx staking cancel-unbond %s1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj 100stake 2 --from mykey`, version.AppName, sdk.GetConfig().GetBech32ValidatorAddrPrefix()),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "validator_address"}, {ProtoField: "amount"}, {ProtoField: "creation_height"}},
				},
				{
					RpcMethod: "UpdateParams",
					Skip:      true, // skipped because authority gated
				},
			},
			EnhanceCustomCommand: true,
		},
	}
}
