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
// See https://github.com/cosmos/cosmos-sdk/blob/main/enterprise/group/LICENSE for full terms.
// Copyright (c) 2026 Cosmos Labs US Inc.

package module

import (
	autocli "cosmossdk.io/core/autocli"

	groupv1 "github.com/cosmos/cosmos-sdk/enterprise/group/x/group"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocli.ModuleOptions {
	return &autocli.ModuleOptions{
		Query: &autocli.ServiceCommandDescriptor{
			Service: groupv1.Query_serviceDesc.ServiceName,
			RpcCommandOptions: []*autocli.RpcCommandOptions{
				{
					RpcMethod: "GroupInfo",
					Use:       "group-info [group-id]",
					Short:     "Query for group info by group id",
					PositionalArgs: []*autocli.PositionalArgDescriptor{
						{ProtoField: "group_id"},
					},
				},
				{
					RpcMethod: "GroupPolicyInfo",
					Use:       "group-policy-info [group-policy-account]",
					Short:     "Query for group policy info by account address of group policy",
					PositionalArgs: []*autocli.PositionalArgDescriptor{
						{ProtoField: "address"},
					},
				},
				{
					RpcMethod: "GroupMembers",
					Use:       "group-members [group-id]",
					Short:     "Query for group members by group id",
					PositionalArgs: []*autocli.PositionalArgDescriptor{
						{ProtoField: "group_id"},
					},
				},
				{
					RpcMethod: "GroupsByAdmin",
					Use:       "groups-by-admin [admin]",
					Short:     "Query for groups by admin account address",
					PositionalArgs: []*autocli.PositionalArgDescriptor{
						{ProtoField: "admin"},
					},
				},
				{
					RpcMethod: "GroupPoliciesByGroup",
					Use:       "group-policies-by-group [group-id]",
					Short:     "Query for group policies by group id",
					PositionalArgs: []*autocli.PositionalArgDescriptor{
						{ProtoField: "group_id"},
					},
				},
				{
					RpcMethod: "GroupPoliciesByAdmin",
					Use:       "group-policies-by-admin [admin]",
					Short:     "Query for group policies by admin account address",
					PositionalArgs: []*autocli.PositionalArgDescriptor{
						{ProtoField: "admin"},
					},
				},
				{
					RpcMethod: "Proposal",
					Use:       "proposal [proposal-id]",
					Short:     "Query for proposal by id",
					PositionalArgs: []*autocli.PositionalArgDescriptor{
						{ProtoField: "proposal_id"},
					},
				},
				{
					RpcMethod: "ProposalsByGroupPolicy",
					Use:       "proposals-by-group-policy [group-policy-account]",
					Short:     "Query for proposals by account address of group policy",
					PositionalArgs: []*autocli.PositionalArgDescriptor{
						{ProtoField: "address"},
					},
				},
				{
					RpcMethod: "VoteByProposalVoter",
					Use:       "vote [proposal-id] [voter]",
					Short:     "Query for vote by proposal id and voter account address",
					PositionalArgs: []*autocli.PositionalArgDescriptor{
						{ProtoField: "proposal_id"},
						{ProtoField: "voter"},
					},
				},
				{
					RpcMethod: "VotesByProposal",
					Use:       "votes-by-proposal [proposal-id]",
					Short:     "Query for votes by proposal id",
					PositionalArgs: []*autocli.PositionalArgDescriptor{
						{ProtoField: "proposal_id"},
					},
				},
				{
					RpcMethod: "VotesByVoter",
					Use:       "votes-by-voter [voter]",
					Short:     "Query for votes by voter account address",
					PositionalArgs: []*autocli.PositionalArgDescriptor{
						{ProtoField: "voter"},
					},
				},
				{
					RpcMethod: "GroupsByMember",
					Use:       "groups-by-member [address]",
					Short:     "Query for groups by member address",
					PositionalArgs: []*autocli.PositionalArgDescriptor{
						{ProtoField: "address"},
					},
				},
				{
					RpcMethod: "TallyResult",
					Use:       "tally-result [proposal-id]",
					Short:     "Query tally result of proposal",
					PositionalArgs: []*autocli.PositionalArgDescriptor{
						{ProtoField: "proposal_id"},
					},
				},
				{
					RpcMethod: "Groups",
					Use:       "groups",
					Short:     "Query for all groups on chain",
				},
			},
		},
		Tx: &autocli.ServiceCommandDescriptor{
			Service:              groupv1.Msg_serviceDesc.ServiceName,
			EnhanceCustomCommand: false, // use custom commands only until v0.51
			RpcCommandOptions: []*autocli.RpcCommandOptions{
				{
					RpcMethod: "UpdateGroupAdmin",
					Use:       "update-group-admin [admin] [group-id] [new-admin]",
					Short:     "Update a group's admin",
					PositionalArgs: []*autocli.PositionalArgDescriptor{
						{ProtoField: "admin"}, {ProtoField: "group_id"}, {ProtoField: "new_admin"},
					},
				},
				{
					RpcMethod: "UpdateGroupMetadata",
					Use:       "update-group-metadata [admin] [group-id] [metadata]",
					Short:     "Update a group's metadata",
					PositionalArgs: []*autocli.PositionalArgDescriptor{
						{ProtoField: "admin"}, {ProtoField: "group_id"}, {ProtoField: "metadata"},
					},
				},
				{
					RpcMethod: "UpdateGroupPolicyAdmin",
					Use:       "update-group-policy-admin [admin] [group-policy-account] [new-admin]",
					Short:     "Update a group policy admin",
					PositionalArgs: []*autocli.PositionalArgDescriptor{
						{ProtoField: "admin"}, {ProtoField: "group_policy_address"}, {ProtoField: "new_admin"},
					},
				},
				{
					RpcMethod: "UpdateGroupPolicyMetadata",
					Use:       "update-group-policy-metadata [admin] [group-policy-account] [new-metadata]",
					Short:     "Update a group policy metadata",
					PositionalArgs: []*autocli.PositionalArgDescriptor{
						{ProtoField: "admin"}, {ProtoField: "group_policy_address"}, {ProtoField: "metadata"},
					},
				},
				{
					RpcMethod: "WithdrawProposal",
					Use:       "withdraw-proposal [proposal-id] [group-policy-admin-or-proposer]",
					Short:     "Withdraw a submitted proposal",
					PositionalArgs: []*autocli.PositionalArgDescriptor{
						{ProtoField: "proposal_id"}, {ProtoField: "address"},
					},
				},
				{
					RpcMethod: "Vote",
					Use:       "vote [proposal-id] [voter] [vote-option] [metadata]",
					Long: `Vote on a proposal.
Parameters:
	proposal-id: unique ID of the proposal
	voter: voter account addresses.
	vote-option: choice of the voter(s)
		VOTE_OPTION_UNSPECIFIED: no-op
		VOTE_OPTION_NO: no
		VOTE_OPTION_YES: yes
		VOTE_OPTION_ABSTAIN: abstain
		VOTE_OPTION_NO_WITH_VETO: no-with-veto
	Metadata: metadata for the vote
`,
					PositionalArgs: []*autocli.PositionalArgDescriptor{
						{ProtoField: "proposal_id"}, {ProtoField: "voter"}, {ProtoField: "option"}, {ProtoField: "metadata"},
					},
					FlagOptions: map[string]*autocli.FlagOptions{
						"exec": {Name: "exec", DefaultValue: "", Usage: "Set to 'try' for trying to execute proposal immediately after voting"},
					},
				},
				{
					RpcMethod: "Exec",
					Use:       "exec [proposal-id]",
					Short:     "Execute a proposal",
					PositionalArgs: []*autocli.PositionalArgDescriptor{
						{ProtoField: "proposal_id"},
					},
				},
				{
					RpcMethod: "LeaveGroup",
					Use:       "leave-group [member-address] [group-id]",
					Short:     "Remove member from the group",
					PositionalArgs: []*autocli.PositionalArgDescriptor{
						{ProtoField: "address"}, {ProtoField: "group_id"},
					},
				},
			},
		},
	}
}
