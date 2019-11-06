# Group Proposal Specification

Groups can submit proposals in order to execute arbitrary transactions.
Authorization to execute messages is performed by the `msg_delegation.Keeper`'s
`DispatchActions` method. The actual protocol that a group uses to determine
whether a certain proposal passes are based on the group accounts's
`DecisionPolicy`.

## Types

### `ProposalID`

```go
type ProposalID uint64
```

An auto-incremented integer ID for each proposal.

### `Vote`

```go
type Vote int

const (
  No Vote = iota
  Yes = 1
  Abstain = 2
  Veto = 3
)
```

## Messages

### `MsgPropose`

```go
type MsgPropose struct {
    Proposer sdk.AccAddress `json:"proposer"`
	GroupAccount    sdk.AccAddress `json:"group_account"`
	Msgs     []sdk.Msg      `json:"msgs"`
	Description string      `json:"Description,omitempty"`
}
```

*Returns:* a new `ProposalID` based on an auto-incremented integer.

### `MsgVote`

Votes are counted as `MsgVote`s are submitted. The final tally of votes should
always be based on the state of the group at the end of the proposal. As a result,
whenever group membership is changed via `MsgUpdateGroupMembers`, existing vote
counts for open proposals must be updated when a member's power is changed.

```go
type MsgVote struct {
	ProposalID ProposalID     `json:"proposal_id"`
	Voter      sdk.AccAddress `json:"voter"`
	Vote       Vote           `json:"vote"`
	Comment string            `json:"comment,omitempty"`
}
```

### `MsgExecProposal`

```go
// MsgExecProposal attempts to execute and close an open proposal based on
// the current Tally of votes and the amount of time that the proposal has
// been open for voting. If the proposal can pass based on those parameters,
// its Msgs will be passed to the router via the msg_delegation.Keeper and the
// proposal will be closed. The signer of this message will pay any gas costs.
// If the proposal cannot pass based on the current parameters, it will remain
// open until the MaxVotingPeriod time is reached and then automatically closed.
type MsgExecProposal struct {
	ProposalID ProposalID     `json:"proposal_id"`
	Signer     sdk.AccAddress `json:"signer"`
}
```
### `MsgWithdrawProposal`

```go
// MsgWithdrawProposal allows the original proposer to withdraw an open proposal.
type MsgWithdrawProposal struct {
	ProposalID ProposalID     `json:"proposal_id"`
	Proposer   sdk.AccAddress `json:"proposer"`
}
```

## Params

The group module contains the following parameters:

| Key | Type | Usage | Example |
|------------------------|-----------------|---------|-----|
| MaxVotingPeriod | string (time ns) | Proposals older than this time will be garbage-collected | "172800000000000" |
| MaxDescriptionCharacters | string (uint64) | | "256" |
| MaxCommentCharacters | string (uint64) | | "256" |

