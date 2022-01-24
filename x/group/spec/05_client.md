<!--
order: 5
-->

# Client

## CLI

A user can query and interact with the `group` module using the CLI.

### Query

The `query` commands allow users to query `group` state.

```bash
simd query group --help
```

#### group-info

The `group-info` command allows users to query for group info by given group id.

```bash
simd query group group-info [id] [flags]
```

Example:

```bash
simd query group group-info 1
```

Example Output:

```bash
admin: cosmos1..
group_id: "1"
metadata: AQ==
total_weight: "3"
version: "1"
```

#### group-policy-info

The `group-policy-info` command allows users to query for group policy info by account address of group policy .

```bash
simd query group group-policy-info [group-policy-account] [flags]
```

Example:

```bash
simd query group group-policy-info cosmos1..
```

Example Output:

```bash
address: cosmos1..
admin: cosmos1..
decision_policy:
  '@type': /cosmos.group.v1beta1.ThresholdDecisionPolicy
  threshold: "1"
  timeout: 600s
group_id: "1"
metadata: AQ==
version: "1"
```

#### group-members

The `group-members` command allows users to query for group members by group id with pagination flags.

```bash
simd query group group-members [id] [flags]
```

Example:

```bash
simd query group group-members 1
```

Example Output:

```bash
members:
- group_id: "1"
  member:
    address: cosmos1..
    metadata: AQ==
    weight: "2"
- group_id: "1"
  member:
    address: cosmos1..
    metadata: AQ==
    weight: "1"
pagination:
  next_key: null
  total: "2"
```

#### groups-by-admin

The `groups-by-admin` command allows users to query for groups by admin account address with pagination flags.

```bash
simd query group groups-by-admin [admin] [flags]
```

Example:

```bash
simd query group groups-by-admin cosmos1..
```

Example Output:

```bash
groups:
- admin: cosmos1..
  group_id: "1"
  metadata: AQ==
  total_weight: "3"
  version: "1"
- admin: cosmos1..
  group_id: "2"
  metadata: AQ==
  total_weight: "3"
  version: "1"
pagination:
  next_key: null
  total: "2"
```

#### group-policies-by-group

The `group-policies-by-group` command allows users to query for group policies by group id with pagination flags.

```bash
simd query group group-policies-by-group [group-id] [flags]
```

Example:

```bash
simd query group group-policies-by-group 1 
```

Example Output:

```bash
group_policies:
- address: cosmos1..
  admin: cosmos1..
  decision_policy:
    '@type': /cosmos.group.v1beta1.ThresholdDecisionPolicy
    threshold: "1"
    timeout: 600s
  group_id: "1"
  metadata: AQ==
  version: "1"
- address: cosmos1..
  admin: cosmos1..
  decision_policy:
    '@type': /cosmos.group.v1beta1.ThresholdDecisionPolicy
    threshold: "1"
    timeout: 600s
  group_id: "1"
  metadata: AQ==
  version: "1"
pagination:
  next_key: null
  total: "2"
```

#### group-policies-by-admin

The `group-policies-by-admin` command allows users to query for group policies by admin account address with pagination flags.

```bash
simd query group group-policies-by-admin [admin] [flags]
```

Example:

```bash
simd query group group-policies-by-admin cosmos1..
```

Example Output:

```bash
group_policies:
- address: cosmos1..
  admin: cosmos1..
  decision_policy:
    '@type': /cosmos.group.v1beta1.ThresholdDecisionPolicy
    threshold: "1"
    timeout: 600s
  group_id: "1"
  metadata: AQ==
  version: "1"
- address: cosmos1..
  admin: cosmos1..
  decision_policy:
    '@type': /cosmos.group.v1beta1.ThresholdDecisionPolicy
    threshold: "1"
    timeout: 600s
  group_id: "1"
  metadata: AQ==
  version: "1"
pagination:
  next_key: null
  total: "2"
```

#### proposal

The `proposal` command allows users to query for proposal by id.

```bash
simd query group proposal [id] [flags]
```

Example:

```bash
simd query group proposal 1
```

Example Output:

```bash
proposal:
  address: cosmos1..
  executor_result: EXECUTOR_RESULT_NOT_RUN
  group_policy_version: "1"
  group_version: "1"
  metadata: AQ==
  msgs:
  - '@type': /cosmos.bank.v1beta1.MsgSend
    amount:
    - amount: "100000000"
      denom: stake
    from_address: cosmos1..
    to_address: cosmos1..
  proposal_id: "1"
  proposers:
  - cosmos1..
  result: RESULT_UNFINALIZED
  status: STATUS_SUBMITTED
  submitted_at: "2021-12-17T07:06:26.310638964Z"
  timeout: "2021-12-17T07:06:27.310638964Z"
  vote_state:
    abstain_count: "0"
    no_count: "0"
    veto_count: "0"
    yes_count: "0"
```

#### proposals-by-group-policy

The `proposals-by-group-policy` command allows users to query for proposals by account address of group policy with pagination flags.

```bash
simd query group proposals-by-group-policy [group-policy-account] [flags]
```

Example:

```bash
simd query group proposals-by-group-policy cosmos1..
```

Example Output:

```bash
pagination:
  next_key: null
  total: "1"
proposals:
- address: cosmos1..
  executor_result: EXECUTOR_RESULT_NOT_RUN
  group_policy_version: "1"
  group_version: "1"
  metadata: AQ==
  msgs:
  - '@type': /cosmos.bank.v1beta1.MsgSend
    amount:
    - amount: "100000000"
      denom: stake
    from_address: cosmos1..
    to_address: cosmos1..
  proposal_id: "1"
  proposers:
  - cosmos1..
  result: RESULT_UNFINALIZED
  status: STATUS_SUBMITTED
  submitted_at: "2021-12-17T07:06:26.310638964Z"
  timeout: "2021-12-17T07:06:27.310638964Z"
  vote_state:
    abstain_count: "0"
    no_count: "0"
    veto_count: "0"
    yes_count: "0"
```

#### vote

The `vote` command allows users to query for vote by proposal id and voter account address.

```bash
simd query group vote [proposal-id] [voter] [flags]
```

Example:

```bash
simd query group vote 1 cosmos1..
```

Example Output:

```bash
vote:
  choice: CHOICE_YES
  metadata: AQ==
  proposal_id: "1"
  submitted_at: "2021-12-17T08:05:02.490164009Z"
  voter: cosmos1..
```

#### votes-by-proposal

The `votes-by-proposal` command allows users to query for votes by proposal id with pagination flags.

```bash
simd query group votes-by-proposal [proposal-id] [flags]
```

Example:

```bash
simd query group votes-by-proposal 1 
```

Example Output:

```bash
pagination:
  next_key: null
  total: "1"
votes:
- choice: CHOICE_YES
  metadata: AQ==
  proposal_id: "1"
  submitted_at: "2021-12-17T08:05:02.490164009Z"
  voter: cosmos1..
```

#### votes-by-voter

The `votes-by-voter` command allows users to query for votes by voter account address with pagination flags.

```bash
simd query group votes-by-voter [voter] [flags]
```

Example:

```bash
simd query group votes-by-voter cosmos1..
```

Example Output:

```bash
pagination:
  next_key: null
  total: "1"
votes:
- choice: CHOICE_YES
  metadata: AQ==
  proposal_id: "1"
  submitted_at: "2021-12-17T08:05:02.490164009Z"
  voter: cosmos1..
```

### Transactions

The `tx` commands allow users to interact with the `group` module.

```bash
simd tx group --help
```

#### create-group

The `create-group` command allows users to create a group which is an aggregation of member accounts with associated weights and
an administrator account.

```bash
simd tx group create-group [admin] [metadata] [members-json-file]
```

Example:

```bash
simd tx group create-group cosmos1.. "AQ==" members.json 
```

#### update-group-admin

The `update-group-admin` command allows users to update a group's admin.

```bash
simd tx group update-group-admin [admin] [group-id] [new-admin] [flags]
```

Example:

```bash
simd tx group update-group-admin cosmos1.. 1 cosmos1..
```

#### update-group-members

The `update-group-members` command allows users to update a group's members.

```bash
simd tx group update-group-members [admin] [group-id] [members-json-file] [flags]
```

Example:

```bash
simd tx group update-group-members cosmos1.. 1 members.json
```

#### update-group-metadata

The `update-group-metadata` command allows users to update a group's metadata.

```bash
simd tx group update-group-metadata [admin] [group-id] [metadata] [flags]
```

Example:

```bash
simd tx group update-group-metadata cosmos1.. 1 "AQ=="
```

#### create-group-policy

The `create-group-policy` command allows users to create a group policy which is an account associated with a group and a decision policy. 

```bash
simd tx group create-group-policy [admin] [group-id] [metadata] [decision-policy] [flags]
```

Example:

```bash
simd tx group create-group-policy cosmos1.. 1 "AQ==" '{"@type":"/cosmos.group.v1beta1.ThresholdDecisionPolicy", "threshold":"1", "timeout":"600s"}' 
```

#### update-group-policy-admin

The `update-group-policy-admin` command allows users to update a group policy admin.

```bash
simd tx group update-group-policy-admin [admin] [group-policy-account] [new-admin] [flags]
```

Example:

```bash
simd tx group update-group-policy-admin cosmos1.. cosmos1.. cosmos1..
```

#### update-group-policy-metadata

The `update-group-policy-metadata` command allows users to update a group policy metadata.

```bash
simd tx group update-group-policy-metadata [admin] [group-policy-account] [new-metadata] [flags]
```

Example:

```bash
simd tx group update-group-policy-metadata cosmos1.. cosmos1.. "AQ=="
```

#### update-group-policy-decision-policy

The `update-group-policy-decision-policy` command allows users to update a group policy's decision policy.

```bash
simd  tx group update-group-policy-decision-policy [admin] [group-policy-account] [decision-policy] [flags]
```

Example:

```bash
simd tx group update-group-policy-decision-policy cosmos1.. cosmos1.. '{"@type":"/cosmos.group.v1beta1.ThresholdDecisionPolicy", "threshold":"2", "timeout":"1000s"}' 
```

#### create-proposal

The `create-proposal` command allows users to submit a new proposal.

```bash
simd tx group create-proposal [group-policy-account] [proposer[,proposer]*] [msg_tx_json_file] [metadata] [flags]
```

Example:

```bash
simd tx group create-proposal cosmos1.. cosmos1.. msg_tx.json "AQ=="
```

#### withdraw-proposal

The `withdraw-proposal` command allows users to withdraw a proposal.

```bash
simd tx group withdraw-proposal [proposal-id] [group-policy-admin-or-proposer]
```

Example:

```bash
simd tx group withdraw-proposal 1 cosmos1..
```

#### vote 

The `vote` command allows users to vote on a proposal.

```bash
simd tx group vote proposal-id] [voter] [choice] [metadata] [flags]
```

Example:

```bash
simd tx group vote 1 cosmos1.. CHOICE_YES "AQ=="
```

#### exec

The `exec` command allows users to execute a proposal.

```bash
simd tx group exec [proposal-id] [flags]
```

Example:

```bash
simd tx group exec 1
```

## gRPC

A user can query the `group` module using gRPC endpoints.

### GroupInfo

The `GroupInfo` endpoint allows users to query for group info by given group id.

```bash
cosmos.group.v1beta1.Query/GroupInfo
```

Example: 

```bash
grpcurl -plaintext \
    -d '{"group_id":1}' localhost:9090 cosmos.group.v1beta1.Query/GroupInfo
```

Example Output:

```bash
{
  "info": {
    "groupId": "1",
    "admin": "cosmos1..",
    "metadata": "AQ==",
    "version": "1",
    "totalWeight": "3"
  }
}
```

### GroupPolicyInfo

The `GroupPolicyInfo` endpoint allows users to query for group policy info by account address of group policy.

```bash
cosmos.group.v1beta1.Query/GroupPolicyInfo
```

Example:

```bash
grpcurl -plaintext \
    -d '{"address":"cosmos1.."}'  localhost:9090 cosmos.group.v1beta1.Query/GroupPolicyInfo
```

Example Output:

```bash
{
  "info": {
    "address": "cosmos1..",
    "groupId": "1",
    "admin": "cosmos1..",
    "version": "1",
    "decisionPolicy": {"@type":"/cosmos.group.v1beta1.ThresholdDecisionPolicy","threshold":"1","timeout":"600s"},
  }
}
```

### GroupMembers

The `GroupMembers` endpoint allows users to query for group members by group id with pagination flags.

```bash
cosmos.group.v1beta1.Query/GroupMembers
```

Example:

```bash
grpcurl -plaintext \
    -d '{"group_id":"1"}'  localhost:9090 cosmos.group.v1beta1.Query/GroupMembers
```

Example Output:

```bash
{
  "members": [
    {
      "groupId": "1",
      "member": {
        "address": "cosmos1..",
        "weight": "1"
      }
    },
    {
      "groupId": "1",
      "member": {
        "address": "cosmos1..",
        "weight": "2"
      }
    }
  ],
  "pagination": {
    "total": "2"
  }
}
```

### GroupsByAdmin

The `GroupsByAdmin` endpoint allows users to query for groups by admin account address with pagination flags.

```bash
cosmos.group.v1beta1.Query/GroupsByAdmin
```

Example:

```bash
grpcurl -plaintext \
    -d '{"admin":"cosmos1.."}'  localhost:9090 cosmos.group.v1beta1.Query/GroupsByAdmin
```

Example Output:

```bash
{
  "groups": [
    {
      "groupId": "1",
      "admin": "cosmos1..",
      "metadata": "AQ==",
      "version": "1",
      "totalWeight": "3"
    },
    {
      "groupId": "2",
      "admin": "cosmos1..",
      "metadata": "AQ==",
      "version": "1",
      "totalWeight": "3"
    }
  ],
  "pagination": {
    "total": "2"
  }
}
```

### GroupPoliciesByGroup

The `GroupPoliciesByGroup` endpoint allows users to query for group policies by group id with pagination flags.

```bash
cosmos.group.v1beta1.Query/GroupPoliciesByGroup
```

Example:

```bash
grpcurl -plaintext \
    -d '{"group_id":"1"}'  localhost:9090 cosmos.group.v1beta1.Query/GroupPoliciesByGroup
```

Example Output:

```bash
{
  "GroupPolicies": [
    {
      "address": "cosmos1..",
      "groupId": "1",
      "admin": "cosmos1..",
      "version": "1",
      "decisionPolicy": {"@type":"/cosmos.group.v1beta1.ThresholdDecisionPolicy","threshold":"1","timeout":"600s"},
    },
    {
      "address": "cosmos1..",
      "groupId": "1",
      "admin": "cosmos1..",
      "version": "1",
      "decisionPolicy": {"@type":"/cosmos.group.v1beta1.ThresholdDecisionPolicy","threshold":"1","timeout":"600s"},
    }
  ],
  "pagination": {
    "total": "2"
  }
}
```

### GroupPoliciesByAdmin

The `GroupPoliciesByAdmin` endpoint allows users to query for group policies by admin account address with pagination flags.

```bash
cosmos.group.v1beta1.Query/GroupPoliciesByAdmin
```

Example:

```bash
grpcurl -plaintext \
    -d '{"admin":"cosmos1.."}'  localhost:9090 cosmos.group.v1beta1.Query/GroupPoliciesByAdmin
```

Example Output:

```bash
{
  "GroupPolicies": [
    {
      "address": "cosmos1..",
      "groupId": "1",
      "admin": "cosmos1..",
      "version": "1",
      "decisionPolicy": {"@type":"/cosmos.group.v1beta1.ThresholdDecisionPolicy","threshold":"1","timeout":"600s"},
    },
    {
      "address": "cosmos1..",
      "groupId": "1",
      "admin": "cosmos1..",
      "version": "1",
      "decisionPolicy": {"@type":"/cosmos.group.v1beta1.ThresholdDecisionPolicy","threshold":"1","timeout":"600s"},
    }
  ],
  "pagination": {
    "total": "2"
  }
}
```

### Proposal

The `Proposal` endpoint allows users to query for proposal by id.

```bash
cosmos.group.v1beta1.Query/Proposal
```

Example:

```bash
grpcurl -plaintext \
    -d '{"proposal_id":"1"}'  localhost:9090 cosmos.group.v1beta1.Query/Proposal
```

Example Output:

```bash
{
  "proposal": {
    "proposalId": "1",
    "address": "cosmos1..",
    "proposers": [
      "cosmos1.."
    ],
    "submittedAt": "2021-12-17T07:06:26.310638964Z",
    "groupVersion": "1",
    "GroupPolicyVersion": "1",
    "status": "STATUS_SUBMITTED",
    "result": "RESULT_UNFINALIZED",
    "voteState": {
      "yesCount": "0",
      "noCount": "0",
      "abstainCount": "0",
      "vetoCount": "0"
    },
    "timeout": "2021-12-17T07:06:27.310638964Z",
    "executorResult": "EXECUTOR_RESULT_NOT_RUN",
    "msgs": [
      {"@type":"/cosmos.bank.v1beta1.MsgSend","amount":[{"denom":"stake","amount":"100000000"}],"fromAddress":"cosmos1..","toAddress":"cosmos1.."}
    ]
  }
}
```

### ProposalsByGroupPolicy

The `ProposalsByGroupPolicy` endpoint allows users to query for proposals by account address of group policy with pagination flags.

```bash
cosmos.group.v1beta1.Query/ProposalsByGroupPolicy
```

Example:

```bash
grpcurl -plaintext \
    -d '{"address":"cosmos1.."}'  localhost:9090 cosmos.group.v1beta1.Query/ProposalsByGroupPolicy
```

Example Output:

```bash
{
  "proposals": [
    {
      "proposalId": "1",
      "address": "cosmos1..",
      "proposers": [
        "cosmos1.."
      ],
      "submittedAt": "2021-12-17T08:03:27.099649352Z",
      "groupVersion": "1",
      "GroupPolicyVersion": "1",
      "status": "STATUS_CLOSED",
      "result": "RESULT_ACCEPTED",
      "voteState": {
        "yesCount": "1",
        "noCount": "0",
        "abstainCount": "0",
        "vetoCount": "0"
      },
      "timeout": "2021-12-17T08:13:27.099649352Z",
      "executorResult": "EXECUTOR_RESULT_NOT_RUN",
      "msgs": [
        {"@type":"/cosmos.bank.v1beta1.MsgSend","amount":[{"denom":"stake","amount":"100000000"}],"fromAddress":"cosmos1..","toAddress":"cosmos1.."}
      ]
    }
  ],
  "pagination": {
    "total": "1"
  }
}
```

### VoteByProposalVoter

The `VoteByProposalVoter` endpoint allows users to query for vote by proposal id and voter account address.

```bash
cosmos.group.v1beta1.Query/VoteByProposalVoter
```

Example:

```bash
grpcurl -plaintext \
    -d '{"proposal_id":"1","voter":"cosmos1.."}'  localhost:9090 cosmos.group.v1beta1.Query/VoteByProposalVoter
```

Example Output:

```bash
{
  "vote": {
    "proposalId": "1",
    "voter": "cosmos1..",
    "choice": "CHOICE_YES",
    "submittedAt": "2021-12-17T08:05:02.490164009Z"
  }
}
```

### VotesByProposal

The `VotesByProposal` endpoint allows users to query for votes by proposal id with pagination flags.

```bash
cosmos.group.v1beta1.Query/VotesByProposal
```

Example:

```bash
grpcurl -plaintext \
    -d '{"proposal_id":"1"}'  localhost:9090 cosmos.group.v1beta1.Query/VotesByProposal
```

Example Output:

```bash
{
  "votes": [
    {
      "proposalId": "1",
      "voter": "cosmos1..",
      "choice": "CHOICE_YES",
      "submittedAt": "2021-12-17T08:05:02.490164009Z"
    }
  ],
  "pagination": {
    "total": "1"
  }
}
```

### VotesByVoter

The `VotesByVoter` endpoint allows users to query for votes by voter account address with pagination flags.

```bash
cosmos.group.v1beta1.Query/VotesByVoter
```

Example:

```bash
grpcurl -plaintext \
    -d '{"voter":"cosmos1.."}'  localhost:9090 cosmos.group.v1beta1.Query/VotesByVoter
```

Example Output:

```bash
{
  "votes": [
    {
      "proposalId": "1",
      "voter": "cosmos1..",
      "choice": "CHOICE_YES",
      "submittedAt": "2021-12-17T08:05:02.490164009Z"
    }
  ],
  "pagination": {
    "total": "1"
  }
}
```

## REST

A user can query the `group` module using REST endpoints.

### GroupInfo

The `GroupInfo` endpoint allows users to query for group info by given group id.

```bash
/cosmos/group/v1beta1/group_info/{group_id}
```

Example:

```bash
curl localhost:1317/cosmos/group/v1beta1/group_info/1
```

Example Output:

```bash
{
  "info": {
    "group_id": "1",
    "admin": "cosmos1..",
    "metadata": "AQ==",
    "version": "1",
    "total_weight": "3"
  }
}
```

### GroupPolicyInfo

The `GroupPolicyInfo` endpoint allows users to query for group policy info by account address of group policy.

```bash
/cosmos/group/v1beta1/group_policy_info/{address}
```

Example:

```bash
curl localhost:1317/cosmos/group/v1beta1/group_policy_info/cosmos1..
```

Example Output:

```bash
{
  "info": {
    "address": "cosmos1..",
    "group_id": "1",
    "admin": "cosmos1..",
    "metadata": "AQ==",
    "version": "1",
    "decision_policy": {
      "@type": "/cosmos.group.v1beta1.ThresholdDecisionPolicy",
      "threshold": "1",
      "timeout": "600s"
    },
  }
}
```

### GroupMembers

The `GroupMembers` endpoint allows users to query for group members by group id with pagination flags.

```bash
/cosmos/group/v1beta1/group_members/{group_id}
```

Example:

```bash
curl localhost:1317/cosmos/group/v1beta1/group_members/1
```

Example Output:

```bash
{
  "members": [
    {
      "group_id": "1",
      "member": {
        "address": "cosmos1..",
        "weight": "1",
        "metadata": "AQ=="
      }
    },
    {
      "group_id": "1",
      "member": {
        "address": "cosmos1..",
        "weight": "2",
        "metadata": "AQ=="
    }
  ],
  "pagination": {
    "next_key": null,
    "total": "2"
  }
}
```

### GroupsByAdmin

The `GroupsByAdmin` endpoint allows users to query for groups by admin account address with pagination flags.

```bash
/cosmos/group/v1beta1/groups_by_admin/{admin}
```

Example:

```bash
curl localhost:1317/cosmos/group/v1beta1/groups_by_admin/cosmos1..
```

Example Output:

```bash
{
  "groups": [
    {
      "group_id": "1",
      "admin": "cosmos1..",
      "metadata": "AQ==",
      "version": "1",
      "total_weight": "3"
    },
    {
      "group_id": "2",
      "admin": "cosmos1..",
      "metadata": "AQ==",
      "version": "1",
      "total_weight": "3"
    }
  ],
  "pagination": {
    "next_key": null,
    "total": "2"
  }
}
```

### GroupPoliciesByGroup

The `GroupPoliciesByGroup` endpoint allows users to query for group policies by group id with pagination flags.

```bash
/cosmos/group/v1beta1/group_policies_by_group/{group_id}
```

Example:

```bash
curl localhost:1317/cosmos/group/v1beta1/group_policies_by_group/1
```

Example Output:

```bash
{
  "group_policies": [
    {
      "address": "cosmos1..",
      "group_id": "1",
      "admin": "cosmos1..",
      "metadata": "AQ==",
      "version": "1",
      "decision_policy": {
        "@type": "/cosmos.group.v1beta1.ThresholdDecisionPolicy",
        "threshold": "1",
        "timeout": "600s"
      },
    },
    {
      "address": "cosmos1..",
      "group_id": "1",
      "admin": "cosmos1..",
      "metadata": "AQ==",
      "version": "1",
      "decision_policy": {
        "@type": "/cosmos.group.v1beta1.ThresholdDecisionPolicy",
        "threshold": "1",
        "timeout": "600s"
      },
    }
  ],
  "pagination": {
    "next_key": null,
    "total": "2"
  }
}
```

### GroupPoliciesByAdmin

The `GroupPoliciesByAdmin` endpoint allows users to query for group policies by admin account address with pagination flags.

```bash
/cosmos/group/v1beta1/group_policies_by_admin/{admin}
```

Example:

```bash
curl localhost:1317/cosmos/group/v1beta1/group_policies_by_admin/cosmos1..
```

Example Output:

```bash
{
  "group_policies": [
    {
      "address": "cosmos1..",
      "group_id": "1",
      "admin": "cosmos1..",
      "metadata": "AQ==",
      "version": "1",
      "decision_policy": {
        "@type": "/cosmos.group.v1beta1.ThresholdDecisionPolicy",
        "threshold": "1",
        "timeout": "600s"
      },
    },
    {
      "address": "cosmos1..",
      "group_id": "1",
      "admin": "cosmos1..",
      "metadata": "AQ==",
      "version": "1",
      "decision_policy": {
        "@type": "/cosmos.group.v1beta1.ThresholdDecisionPolicy",
        "threshold": "1",
        "timeout": "600s"
      },
    }
  ],
  "pagination": {
    "next_key": null,
    "total": "2"
  }
```

### Proposal

The `Proposal` endpoint allows users to query for proposal by id.

```bash
/cosmos/group/v1beta1/proposal/{proposal_id}
```

Example:

```bash
curl localhost:1317/cosmos/group/v1beta1/proposal/1
```

Example Output:

```bash
{
  "proposal": {
    "proposal_id": "1",
    "address": "cosmos1..",
    "metadata": "AQ==",
    "proposers": [
      "cosmos1.."
    ],
    "submitted_at": "2021-12-17T07:06:26.310638964Z",
    "group_version": "1",
    "group_policy_version": "1",
    "status": "STATUS_SUBMITTED",
    "result": "RESULT_UNFINALIZED",
    "vote_state": {
      "yes_count": "0",
      "no_count": "0",
      "abstain_count": "0",
      "veto_count": "0"
    },
    "timeout": "2021-12-17T07:06:27.310638964Z",
    "executor_result": "EXECUTOR_RESULT_NOT_RUN",
    "msgs": [
      {
        "@type": "/cosmos.bank.v1beta1.MsgSend",
        "from_address": "cosmos1..",
        "to_address": "cosmos1..",
        "amount": [
          {
            "denom": "stake",
            "amount": "100000000"
          }
        ]
      }
    ]
  }
}
```

### ProposalsByGroupPolicy

The `ProposalsByGroupPolicy` endpoint allows users to query for proposals by account address of group policy with pagination flags.

```bash
/cosmos/group/v1beta1/proposals_by_group_policy/{address}
```

Example:

```bash
curl localhost:1317/cosmos/group/v1beta1/proposals_by_group_policy/cosmos1..
```

Example Output:

```bash
{
  "proposals": [
    {
      "proposal_id": "1",
      "address": "cosmos1..",
      "metadata": "AQ==",
      "proposers": [
        "cosmos1.."
      ],
      "submitted_at": "2021-12-17T08:03:27.099649352Z",
      "group_version": "1",
      "group_policy_version": "1",
      "status": "STATUS_CLOSED",
      "result": "RESULT_ACCEPTED",
      "vote_state": {
        "yes_count": "1",
        "no_count": "0",
        "abstain_count": "0",
        "veto_count": "0"
      },
      "timeout": "2021-12-17T08:13:27.099649352Z",
      "executor_result": "EXECUTOR_RESULT_NOT_RUN",
      "msgs": [
        {
          "@type": "/cosmos.bank.v1beta1.MsgSend",
          "from_address": "cosmos1..",
          "to_address": "cosmos1..",
          "amount": [
            {
              "denom": "stake",
              "amount": "100000000"
            }
          ]
        }
      ]
    }
  ],
  "pagination": {
    "next_key": null,
    "total": "1"
  }
}
```

### VoteByProposalVoter

The `VoteByProposalVoter` endpoint allows users to query for vote by proposal id and voter account address.

```bash
/cosmos/group/v1beta1/vote_by_proposal_voter/{proposal_id}/{voter}
```

Example:

```bash
curl localhost:1317/cosmos/group/v1beta1/vote_by_proposal_voter/1/cosmos1..
```

Example Output:

```bash
{
  "vote": {
    "proposal_id": "1",
    "voter": "cosmos1..",
    "choice": "CHOICE_YES",
    "metadata": "AQ==",
    "submitted_at": "2021-12-17T08:05:02.490164009Z"
  }
}
```

### VotesByProposal

The `VotesByProposal` endpoint allows users to query for votes by proposal id with pagination flags.

```bash
/cosmos/group/v1beta1/votes_by_proposal/{proposal_id}
```

Example:

```bash
curl localhost:1317/cosmos/group/v1beta1/votes_by_proposal/1
```

Example Output:

```bash
{
  "votes": [
    {
      "proposal_id": "1",
      "voter": "cosmos1..",
      "choice": "CHOICE_YES",
      "metadata": "AQ==",
      "submitted_at": "2021-12-17T08:05:02.490164009Z"
    }
  ],
  "pagination": {
    "next_key": null,
    "total": "1"
  }
}
```

### VotesByVoter

The `VotesByVoter` endpoint allows users to query for votes by voter account address with pagination flags.

```bash
/cosmos/group/v1beta1/votes_by_voter/{voter}
```

Example:

```bash
curl localhost:1317/cosmos/group/v1beta1/votes_by_voter/cosmos1..
```

Example Output:

```bash
{
  "votes": [
    {
      "proposal_id": "1",
      "voter": "cosmos1..",
      "choice": "CHOICE_YES",
      "metadata": "AQ==",
      "submitted_at": "2021-12-17T08:05:02.490164009Z"
    }
  ],
  "pagination": {
    "next_key": null,
    "total": "1"
  }
}
```
