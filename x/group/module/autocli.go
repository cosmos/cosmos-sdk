package module

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	groupv1 "cosmossdk.io/api/cosmos/group/v1"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: groupv1.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "GroupInfo",
					Use:       "group-info [group-id]",
					Short:     "Query for group info by group id",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "group_id"},
					},
				},
				{
					RpcMethod: "GroupPolicyInfo",
					Use:       "group-policy-info [group-policy-account]",
					Short:     "Query for group policy info by account address of group policy",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "address"},
					},
				},
				{
					RpcMethod: "GroupMembers",
					Use:       "group-members [group-id]",
					Short:     "Query for group members by group id",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "group_id"},
					},
				},
				{
					RpcMethod: "GroupsByAdmin",
					Use:       "groups-by-admin [admin]",
					Short:     "Query for groups by admin account address",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "admin"},
					},
				},
				{
					RpcMethod: "GroupPoliciesByGroup",
					Use:       "group-policies-by-group [group-id]",
					Short:     "Query for group policies by group id",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "group_id"},
					},
				},
				{
					RpcMethod: "GroupPoliciesByAdmin",
					Use:       "group-policies-by-admin [admin]",
					Short:     "Query for group policies by admin account address",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "admin"},
					},
				},
				{
					RpcMethod: "Proposal",
					Use:       "proposal [proposal-id]",
					Short:     "Query for proposal by id",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "proposal_id"},
					},
				},
				{
					RpcMethod: "ProposalsByGroupPolicy",
					Use:       "proposals-by-group-policy [group-policy-account]",
					Short:     "Query for proposals by account address of group policy",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "address"},
					},
				},
				{
					RpcMethod: "VoteByProposalVoter",
					Use:       "vote [proposal-id] [voter]",
					Short:     "Query for vote by proposal id and voter account address",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "proposal_id"},
						{ProtoField: "voter"},
					},
				},
				{
					RpcMethod: "VotesByProposal",
					Use:       "votes-by-proposal [proposal-id]",
					Short:     "Query for votes by proposal id",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "proposal_id"},
					},
				},
				{
					RpcMethod: "VotesByVoter",
					Use:       "votes-by-voter [voter]",
					Short:     "Query for votes by voter account address",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "voter"},
					},
				},
				{
					RpcMethod: "GroupsByMember",
					Use:       "groups-by-member [address]",
					Short:     "Query for groups by member address",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "address"},
					},
				},
				{
					RpcMethod: "TallyResult",
					Use:       "tally-result [proposal-id]",
					Short:     "Query tally result of proposal",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
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
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service: groupv1.Query_ServiceDesc.ServiceName,
		},
	}
}
