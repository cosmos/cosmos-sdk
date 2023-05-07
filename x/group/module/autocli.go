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
					RpcMethod:      "GroupsByMember",
					Use:            "groups-by-member [address]",
					Short:          "Query for groups by member address with pagination flags",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "address"}},
				},
				{
					RpcMethod:      "GroupInfo",
					Use:            "group-info [id]",
					Short:          "Query for group info by group id",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "group_id"}},
				},
				{
					RpcMethod:      "GroupPolicyInfo",
					Use:            "group-policy-info [group-policy-account]",
					Short:          "Query for group policy info by account address of group policy",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "address"}},
				},
				{
					RpcMethod:      "GroupMembers",
					Use:            "group-members [id]",
					Short:          "Query for group members by group id with pagination flags",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "group_id"}},
				},
				{
					RpcMethod:      "GroupPoliciesByAdmin",
					Use:            "group-policies-by-admin [admin]",
					Short:          "Query for groups by admin account address with pagination flags",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "admin"}},
				},
				{
					RpcMethod:      "Proposal",
					Use:            "proposal [id]",
					Short:          "Query for proposal by id",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "proposal_id"}},
				},
				{
					RpcMethod:      "ProposalsByGroupPolicy",
					Use:            "proposals-by-group-policy [group-policy-account]",
					Short:          "Query for proposals by account address of group policy with pagination flags",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "address"}},
				},
				{
					RpcMethod:      "VoteByProposalVoter",
					Use:            "vote [proposal-id] [voter]",
					Short:          "Query for vote by proposal id and voter account address",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "proposal_id"}, {ProtoField: "voter"}},
				},
				{
					RpcMethod:      "VotesByProposal",
					Use:            "votes-by-proposal [proposal-id]",
					Short:          "Query for votes by proposal id with pagination flags",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "proposal_id"}},
				},
				{
					RpcMethod:      "TallyResult",
					Use:            "tally-result [proposal-id]",
					Short:          "Query tally result of proposal",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "proposal_id"}},
				},
				{
					RpcMethod:      "VotesByVoter",
					Use:            "votes-by-voter [voter]",
					Short:          "Query for votes by voter account address with pagination flags",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "voter"}},
				},
				{
					RpcMethod: "Groups",
					Use:       "groups",
					Short:     "Query for groups present in the state",
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service: groupv1.Msg_ServiceDesc.ServiceName,
		},
	}
}
