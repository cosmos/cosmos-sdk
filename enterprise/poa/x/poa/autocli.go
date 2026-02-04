// IMPORTANT LICENSE NOTICE
//
// SPDX-License-Identifier: CosmosLabs-Evaluation-Only
//
// This file is NOT licensed under the Apache License 2.0.
//
// Licensed under the Cosmos Labs Source Available Evaluation License, which forbids:
// - commercial use,
// - production use, and
// - redistribution.
//
// See https://github.com/cosmos/cosmos-sdk/blob/main/enterprise/poa/LICENSE for full terms.
// Copyright (c) 2026 Cosmos Labs US Inc.

package poa

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

	poav1 "github.com/cosmos/cosmos-sdk/enterprise/poa/api/cosmos/poa/v1"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: poav1.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Query the current poa module parameters",
				},
				{
					RpcMethod: "Validator",
					Use:       "validator [address]",
					Short:     "Query a validator by consensus or operator address",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "address"},
					},
				},
				{
					RpcMethod: "Validators",
					Use:       "validators",
					Short:     "Query all validators",
				},
				{
					RpcMethod: "WithdrawableFees",
					Use:       "withdrawable-fees [operator-address]",
					Short:     "Query withdrawable fees for a validator operator",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "operator_address"},
					},
				},
				{
					RpcMethod: "TotalPower",
					Use:       "total-power",
					Short:     "Query the total voting power",
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service:              poav1.Msg_ServiceDesc.ServiceName,
			EnhanceCustomCommand: true, // Preserve custom tx commands (CreateValidator has complex pubkey handling)
		},
	}
}
