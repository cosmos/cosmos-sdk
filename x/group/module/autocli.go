package module

import (
	"fmt"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	groupv1 "cosmossdk.io/api/cosmos/group/v1"

	"github.com/cosmos/cosmos-sdk/version"
)

const (
	FlagGroupPolicyAsAdmin = "group-policy-as-admin"
	FlaExec                = "exec"
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
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "CreateGroup",
					Use:       "create-group [admin] [metadata] [members-json-file]",
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
						{ProtoField: "admin"},
						{ProtoField: "metadata"},
						{ProtoField: "members"},
					},
				},
				{
					RpcMethod: "UpdateGroupMembers",
					Use:       "update-group-members [admin] [group-id] [members-json-file]",
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
						{ProtoField: "admin"},
						{ProtoField: "group_id"},
						{ProtoField: "member_updates"},
					},
				},
				{
					RpcMethod: "UpdateGroupAdmin",
					Use:       "update-group-admin [admin] [group-id] [new-admin]",
					Short:     "Update a group's admin",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "admin"},
						{ProtoField: "group_id"},
						{ProtoField: "new_admin"},
					},
				},
				{
					RpcMethod: "UpdateGroupMetadata",
					Use:       "update-group-metadata [admin] [group-id] [metadata]",
					Short:     "Update a group's metadata",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "admin"},
						{ProtoField: "group_id"},
						{ProtoField: "metadata"},
					},
				},
				{
					RpcMethod: "CreateGroupPolicy",
					Use:       "create-group-policy [admin] [group-id] [metadata] [decision-policy-json-file]",
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
}`, version.AppName), PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "admin"},
						{ProtoField: "group_id"},
						{ProtoField: "metadata"},
						{ProtoField: "decision_policy"},
					},
				},
				{
					RpcMethod: "CreateGroupWithPolicy",
					Use:       "create-group-with-policy [admin] [group-metadata] [group-policy-metadata] [members-json-file] [decision-policy-json-file]",
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
						{ProtoField: "admin"},
						{ProtoField: "members"},
						{ProtoField: "group_metadata"},
						{ProtoField: "group_policy_metadata"},
					},
					FlagOptions: map[string]*autocliv1.FlagOptions{
						FlagGroupPolicyAsAdmin: {
							Name:         FlagGroupPolicyAsAdmin,
							Usage:        "Sets admin of the newly created group and group policy with group policy address itself when true",
							DefaultValue: "false",
						},
					},
				},
				{
					RpcMethod: "UpdateGroupPolicyAdmin",
					Use:       "update-group-policy-admin [admin] [group-policy-account] [new-admin]",
					Short:     "Update a group policy admin",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "admin"},
						{ProtoField: "group_policy_address"},
						{ProtoField: "new_admin"},
					},
				},
				{
					RpcMethod: "UpdateGroupPolicyDecisionPolicy",
					Use:       "update-group-policy-decision-policy [admin] [group-policy-account] [decision-policy-json-file]",
					Short:     "Update a group policy's decision policy",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "admin"},
						{ProtoField: "group_policy_address"},
						{ProtoField: "decision_policy"},
					},
				},
				{
					RpcMethod: "UpdateGroupPolicyMetadata",
					Use:       "update-group-policy-metadata [admin] [group-policy-account] [new-metadata]",
					Short:     "Update a group policy metadata",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "admin"},
						{ProtoField: "group_policy_address"},
						{ProtoField: "metadata"},
					},
				},
				{
					RpcMethod: "SubmitProposal",
					Use:       "submit-proposal [proposal_json_file]",
					Short:     "Submit a new proposal",
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
						{ProtoField: "proposal"},
					},
					FlagOptions: map[string]*autocliv1.FlagOptions{
						FlaExec: {
							Name:  FlaExec,
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
						{ProtoField: "address"},
					},
				},
				{
					RpcMethod: "Vote",
					Use:       "vote [proposal-id] [voter] [vote-option] [metadata]",
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
						{ProtoField: "voter"},
						{ProtoField: "option"},
						{ProtoField: "metadata"},
					},
					FlagOptions: map[string]*autocliv1.FlagOptions{
						FlaExec: {
							Name:  FlaExec,
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
						{ProtoField: "executor"},
					},
				},
				{
					RpcMethod: "LeaveGroup",
					Use:       "leave-group [member-address] [group-id]",
					Short:     "Remove member from the group",
					Long: `Remove member from the group

Parameters:
		   group-id: unique id of the group
		   member-address: account address of the group member
		   Note, the '--from' flag is ignored as it is implied from [member-address]
		`,
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "address"},
						{ProtoField: "group_id"},
					},
				},
			},
		},
	}
}
