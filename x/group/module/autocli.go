package module

import (
	"fmt"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	groupv1 "cosmossdk.io/api/cosmos/group/v1"

	"github.com/cosmos/cosmos-sdk/version"
)

const (
	FlagExec               = "exec"
	FlagGroupPolicyAsAdmin = "group-policy-as-admin"
)

func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: groupv1.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod:      "GroupByMember",
					Use:            "group-by-member [member-address]",
					Short:          "Query for groups by member address with pagination flags",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "member_address"}},
				},
				{
					RpcMethod:      "GroupInfo",
					Use:            "group-info [id]",
					Short:          "Query for group info by group id",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "id"}},
				},
				{
					RpcMethod:      "GroupPolicyInfo",
					Use:            "group-policy-info [group-policy-account]",
					Short:          "Query for group policy info by group policy account",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "group_policy_account"}},
				},
				{
					RpcMethod:      "GroupMembers",
					Use:            "group-members [group-id]",
					Short:          "Query for group members by group id",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "group_id"}},
				},
				{
					RpcMethod:      "GroupByAdmin",
					Use:            "group-by-admin [admin-address]",
					Short:          "Query for groups by admin address with pagination flags",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "admin_address"}},
				},
				{
					RpcMethod:      "GroupPoliciesByGroup",
					Use:            "group-policies-by-group [group-id]",
					Short:          "Query for group policies by group id",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "group_id"}},
				},
				{
					RpcMethod:      "GroupPoliciesByAdmin",
					Use:            "group-policies-by-admin [admin-address]",
					Short:          "Query for group policies by admin address",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "admin_address"}},
				},
				{
					RpcMethod:      "Proposal",
					Use:            "proposal [proposal-id]",
					Short:          "Query for proposal by proposal id",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "proposal_id"}},
				},
				{
					RpcMethod:      "ProposalsByGroupPolicy",
					Use:            "proposals-by-group-policy [group-policy-account]",
					Short:          "Query for proposals by group policy account",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "group_policy_account"}},
				},
				{
					RpcMethod:      "Vote",
					Use:            "vote [proposal-id] [voter-address]",
					Short:          "Query for vote by proposal id and voter address",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "proposal_id"}, {ProtoField: "voter_address"}},
				},
				{
					RpcMethod:      "VotesByProposal",
					Use:            "votes-by-proposal [proposal-id]",
					Short:          "Query for votes by proposal id",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "proposal_id"}},
				},
				{
					RpcMethod:      "TallyResult",
					Use:            "tally-result [proposal-id]",
					Short:          "Query for tally result by proposal id",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "proposal_id"}},
				},
				{
					RpcMethod:      "VotesByVoter",
					Use:            "votes-by-voter [voter-address]",
					Short:          "Query for votes by voter address",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "voter_address"}},
				},
				{
					RpcMethod: "Groups",
					Use:       "groups",
					Short:     "Query for all groups present in the state",
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service: groupv1.Msg_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "CreateGroup",
					Use:       "create-group [admin-address] [metadata] [members-json-file]",
					Short:     "Create a group which is an aggregation of member accounts with associated weights and an administrator account.",
					Long: `Create a group which is an aggregation of member accounts with associated weights and an administrator account.
Note, the '--from' flag is ignored as it is implied from [admin]. Members accounts can be given through a members JSON file that contains an array of members.`,
					Example: fmt.Sprintf(`
%s tx group create-group [admin] [metadata] [members-json-file]

Where members.json contains:

{
	"members": [
		{
			"address": "addr1",
			"weight": "1",
			"metadata": "some metadata"
		},
		{
			"address": "addr2",
			"weight": "1",
			"metadata": "some metadata"
		}
	]
}`, version.AppName),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "admin_address"},
						{ProtoField: "metadata"},
						{ProtoField: "members_json_file"},
					},
				},
				{
					RpcMethod: "UpdateGroupMembers",
					Use:       "update-group-members [admin-address] [group-id] [members-json-file]",
					Short:     "Update a group's members. Set a member's weight to \"0\" to delete it.",
					Example: fmt.Sprintf(`
%s tx group update-group-members [admin] [group-id] [members-json-file]

Where members.json contains:

{
	"members": [
		{
			"address": "addr1",
			"weight": "1",
			"metadata": "some new metadata"
		},
		{
			"address": "addr2",
			"weight": "0",
			"metadata": "some metadata"
		}
	]
}

Set a member's weight to "0" to delete it.
`, version.AppName),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "admin_address"},
						{ProtoField: "group_id"},
						{ProtoField: "members_json_file"},
					},
				},
				{
					RpcMethod: "UpdateGroupAdmin",
					Use:       "update-group-admin [admin-address] [group-id] [new-admin-address]",
					Short:     "Update a group's admin.",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "admin_address"},
						{ProtoField: "group_id"},
						{ProtoField: "new_admin_address"},
					},
				},
				{
					RpcMethod: "UpdateGroupMetadata",
					Use:       "update-group-metadata [admin-address] [group-id] [metadata]",
					Short:     "Update a group's metadata.",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "admin_address"},
						{ProtoField: "group_id"},
						{ProtoField: "metadata"},
					},
				},
				{
					RpcMethod: "CreateGroupPolicy",
					Use:       "create-group-policy [admin-address] [group-metadata] [group-policy-metadata] [members-json-file] [decision-policy-json-file]",
					Short:     "Create a group with policy which is an aggregation of member accounts with associated weights, an administrator account and decision policy.",
					Long: `Create a group with policy which is an aggregation of member accounts with associated weights,
an administrator account and decision policy. Note, the '--from' flag is ignored as it is implied from [admin].
Members accounts can be given through a members JSON file that contains an array of members.
If group-policy-as-admin flag is set to true, the admin of the newly created group and group policy is set with the group policy address itself.`,
					Example: fmt.Sprintf(`
%s tx group create-group-with-policy [admin] [group-metadata] [group-policy-metadata] members.json policy.json

where members.json contains:

{
	"members": [
		{
			"address": "addr1",
			"weight": "1",
			"metadata": "some metadata"
		},
		{
			"address": "addr2",
			"weight": "1",
			"metadata": "some metadata"
		}
	]
}

and policy.json contains:

{
    "@type": "/cosmos.group.v1.ThresholdDecisionPolicy",
    "threshold": "1",
    "windows": {
        "voting_period": "120h",
        "min_execution_period": "0s"
    }
}
`, version.AppName),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "admin_address"},
						{ProtoField: "group_metadata"},
						{ProtoField: "group_policy_metadata"},
						{ProtoField: "members_json_file"},
						{ProtoField: "decision_policy_json_file"},
					},
					FlagOptions: map[string]*autocliv1.FlagOptions{
						FlagGroupPolicyAsAdmin: {
							Name:  FlagGroupPolicyAsAdmin,
							Usage: "Sets admin of the newly created group and group policy with group policy address itself when true",
						},
					},
				},
				{
					RpcMethod: "CreateGroupPolicy",
					Use:       "create-group-policy [admin-address] [group-id] [metadata] [decision-policy-json-file]",
					Short:     `Create a group policy which is an account associated with a group and a decision policy. Note, the '--from' flag is ignored as it is implied from [admin].`,
					Example: fmt.Sprintf(`
%s tx group create-group-policy [admin] [group-id] [metadata] policy.json

where policy.json contains:

{
    "@type": "/cosmos.group.v1.ThresholdDecisionPolicy",
    "threshold": "1",
    "windows": {
        "voting_period": "120h",
        "min_execution_period": "0s"
    }
}

Here, we can use percentage decision policy when needed, where 0 < percentage <= 1:

{
    "@type": "/cosmos.group.v1.PercentageDecisionPolicy",
    "percentage": "0.5",
    "windows": {
        "voting_period": "120h",
        "min_execution_period": "0s"
    }
}`, version.AppName),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "admin_address"},
						{ProtoField: "group_id"},
						{ProtoField: "metadata"},
						{ProtoField: "decision_policy_json_file"},
					},
				},
				{
					RpcMethod: "UpdateGroupPolicyAdmin",
					Use:       "update-group-policy-admin [admin-address] [group-policy-account] [new-admin-address]",
					Short:     "Update a group policy admin",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "admin_address"},
						{ProtoField: "group_policy_account"},
						{ProtoField: "new_admin_address"},
					},
				},
				{
					RpcMethod: "UpdateGroupPolicyDecisionPolicy",
					Use:       "update-group-policy-decision-policy [admin-address] [group-policy-account] [decision-policy-json-file]",
					Short:     "Update a group policy's decision policy",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "admin_address"},
						{ProtoField: "group_policy_account"},
						{ProtoField: "decision_policy_json_file"},
					},
				},
				{
					RpcMethod: "SubmitProposal",
					Use:       "submit-proposal [proposal-json-file]",
					Short:     "Submit a proposal to a group policy",
					Long: `Submit a new proposal.
Parameters:
			msg_tx_json_file: path to json file with messages that will be executed if the proposal is accepted.`,
					Example: fmt.Sprintf(`
%s tx group submit-proposal path/to/proposal.json
	
	Where proposal.json contains:

{
	"group_policy_address": "cosmos1...",
	// array of proto-JSON-encoded sdk.Msgs
	"messages": [
	{
		"@type": "/cosmos.bank.v1beta1.MsgSend",
		"from_address": "cosmos1...",
		"to_address": "cosmos1...",
		"amount":[{"denom": "stake","amount": "10"}]
		"title": "My proposal",
		"summary": "This is a proposal to send 10 stake to cosmos1...",
	}
	],
	// metadata can be any of base64 encoded, raw text, stringified json, IPFS link to json
	// see below for example metadata
	"metadata": "4pIMOgIGx1vZGU=", // base64-encoded metadata
	"proposers": ["cosmos1...", "cosmos1..."],
}

metadata example: 
{
	"title": "",
	"authors": [""],
	"summary": "",
	"details": "", 
	"proposal_forum_url": "",
	"vote_option_context": "",
} 
`, version.AppName),

					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "proposal_json_file"},
					},
					FlagOptions: map[string]*autocliv1.FlagOptions{
						FlagExec: {
							Name:  FlagExec,
							Usage: "Set to 1 to try to execute proposal immediately after creation (proposers signatures are considered as Yes votes)",
						},
					},
				},
				{
					RpcMethod: "WithdrawProposal",
					Use:       "withdraw-proposal [proposal-id] [group-policy-admin-or-proposer]",
					Short:     "Withdraw a submitted proposal",
					Long: `Withdraw a submitted proposal.

Parameters:
			proposal-id: unique ID of the proposal.
			group-policy-admin-or-proposer: either admin of the group policy or one the proposer of the proposal.
			Note: --from flag will be ignored here.
`,
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "proposal_id"},
						{ProtoField: "group_policy_admin_or_proposer"},
					},
				},
				{
					RpcMethod: "Vote",
					Use:       "vote [proposal-id] [voter-address] [vote-option] [metadata]",
					Short:     "Vote on a proposal",
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
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "proposal_id"},
						{ProtoField: "voter_address"},
						{ProtoField: "vote_option"},
						{ProtoField: "metadata"},
					},
					FlagOptions: map[string]*autocliv1.FlagOptions{
						FlagExec: {
							Name:  FlagExec,
							Usage: "Set to 1 to try to execute proposal immediately after voting",
						},
					},
				},
				{
					RpcMethod: "Exec",
					Use:       "exec [proposal-id]",
					Short:     "Execute a proposal",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "proposal_id"},
					},
				},
				{
					RpcMethod: "LeaveGroup",
					Use:       "leave-group [member-address] [group-id] ",
					Short:     "Remove member from the group",
					Long: `Remove member from the group

Parameters:
		   group-id: unique id of the group
		   member-address: account address of the group member
		   Note, the '--from' flag is ignored as it is implied from [member-address]
		`,
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "member_address"},
						{ProtoField: "group_id"},
					},
				},
			},
		},
	}
}
