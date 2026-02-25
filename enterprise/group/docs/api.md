# Group Module API Reference

## gRPC Queries

| Service | RPC | Description |
|---------|-----|-------------|
| Query | GroupInfo | Query group info by group ID |
| Query | GroupPolicyInfo | Query group policy info by address |
| Query | GroupMembers | Query group members by group ID |
| Query | GroupsByAdmin | Query groups by admin address |
| Query | GroupPoliciesByGroup | Query group policies by group ID |
| Query | GroupPoliciesByAdmin | Query group policies by admin address |
| Query | Proposal | Query proposal by ID |
| Query | ProposalsByGroupPolicy | Query proposals by group policy |
| Query | VoteByProposalVoter | Query vote by proposal and voter |
| Query | VotesByProposal | Query votes by proposal |
| Query | VotesByVoter | Query votes by voter |

## Msg Service

| RPC | Description |
|-----|-------------|
| CreateGroup | Create a new group |
| UpdateGroupMembers | Update group members |
| UpdateGroupAdmin | Update group admin |
| UpdateGroupMetadata | Update group metadata |
| CreateGroupPolicy | Create a group policy |
| CreateGroupWithPolicy | Create group with policy |
| UpdateGroupPolicyAdmin | Update group policy admin |
| UpdateGroupPolicyDecisionPolicy | Update decision policy |
| UpdateGroupPolicyMetadata | Update group policy metadata |
| SubmitProposal | Submit a proposal |
| WithdrawProposal | Withdraw a proposal |
| Vote | Vote on a proposal |
| Exec | Execute a proposal |
| LeaveGroup | Leave a group |

## CLI

```bash
# Query commands
simd query group --help

# Transaction commands  
simd tx group --help
```
