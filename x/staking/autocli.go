package staking

import (
	"fmt"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	_ "cosmossdk.io/api/cosmos/crypto/secp256k1" // register to that it shows up in protoregistry.GlobalTypes
	_ "cosmossdk.io/api/cosmos/crypto/secp256r1" // register to that it shows up in protoregistry.GlobalTypes
	stakingv1beta "cosmossdk.io/api/cosmos/staking/v1beta1"

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
					Example:   fmt.Sprintf("$ %s query staking validators", version.AppName),
				},
				{
					RpcMethod: "Validator",
					Use:       "validator [validator-addr]",
					Short:     "Query a validator",
					Long:      "Query details about an individual validator.",
					Example:   fmt.Sprintf("$ %s query staking validator [validator-adr]", version.AppName),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "validator_addr"},
					},
				},
				{
					RpcMethod: "ValidatorDelegations",
					Use:       "delegations-to [validator-addr]",
					Short:     "Query all delegations made to one validator",
					Long:      "Query delegations on an individual validator.",
					Example:   fmt.Sprintf("$ %s query staking delegations-to [validator-adr]", version.AppName),
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
					Example: fmt.Sprintf(`%s query staking delegation [delegator-address] [validator-address]`,
						version.AppName),
					Long: "Query delegations for an individual delegator on an individual validator",
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
					Example:   fmt.Sprintf("$ %s query staking unbonding-delegation [delegator-adr] [validator-adr]", version.AppName),
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
					Example:   fmt.Sprintf("$ %s query staking delegations [validator-addr]", version.AppName),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "delegator_addr"},
					},
				},
				{
					RpcMethod: "DelegatorUnbondingDelegations",
					Use:       "unbonding-delegations [delegator-addr]",
					Short:     "Query all unbonding-delegations records for one delegator",
					Long:      "Query unbonding delegations for an individual delegator.",
					Example:   fmt.Sprintf("$ %s query staking unbonding-delegations [delegator-addr]", version.AppName),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "delegator_addr"},
					},
				},
				{
					RpcMethod: "Redelegations",
					Use:       "redelegation [delegator-addr] [src-validator-addr] [dst-validator-addr]",
					Short:     "Query a redelegation record based on delegator and a source and destination validator address",
					Long:      "Query a redelegation record for an individual delegator between a source and destination validator.",
					Example:   fmt.Sprintf("$ %s query staking redelegation [delegator-addr] [src-validator-addr] [dst-validator-addr]", version.AppName),
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
					Example:   fmt.Sprintf("$ %s query staking pool", version.AppName),
				},
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Query the current staking parameters information",
					Long:      "Query values set as staking parameters.",
					Example:   fmt.Sprintf("$ %s query staking params", version.AppName),
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service: stakingv1beta.Msg_ServiceDesc.ServiceName,
		},
	}
}
