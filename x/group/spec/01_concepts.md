<!--
order: 1
-->

# Concepts

## Group

A group is simply an aggregation of accounts with associated weights. It is not
an account and doesn't have a balance. It doesn't in and of itself have any
sort of voting or decision weight. It does have an "administrator" which has
the ability to add, remove and update members in the group. Note that a
group policy account could be an administrator of a group, and that the
administrator doesn't necessarily have to be a member of the group.

## Group Policy

A group policy is an account associated with a group and a decision policy.
Group policies are abstracted from groups because a single group may have
multiple decision policies for different types of actions. Managing group
membership separately from decision policies results in the least overhead
and keeps membership consistent across different policies. The pattern that
is recommended is to have a single master group policy for a given group,
and then to create separate group policies with different decision policies
and delegate the desired permissions from the master account to
those "sub-accounts" using the `x/authz` module.

## Decision Policy

A decision policy is the mechanism by which members of a group can vote on
proposals, as well as the rules that dictate whether a proposal should pass
or not based on its tally outcome.

All decision policies generally would have a mininum execution period and a
maximum voting window. The minimum execution period is the minimum amount of time
that must pass after submission in order for a proposal to potentially be executed, and it may
be set to 0. The maximum voting window is the maximum time after submission that a proposal may
be voted on before it is tallied.

The chain developer also defines an app-wide maximum execution period, which is
the maximum amount of time after a proposal's voting period end where users are
allowed to execute a proposal.

The current group module comes shipped with two decision policies: threshold
and percentage. Any chain developer can extend upon these two, by creating
custom decision policies, as long as they adhere to the `DecisionPolicy`
interface:

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.46.0-rc1/x/group/types.go#L27-L41

### Threshold decision policy

A threshold decision policy defines a threshold of yes votes (based on a tally
of voter weights) that must be achieved in order for a proposal to pass. For
this decision policy, abstain and veto are simply treated as no's.

This decision policy also has a VotingPeriod window and a MinExecutionPeriod
window. The former defines the duration after proposal submission where members
are allowed to vote, after which tallying is performed. The latter specifies
the minimum duration after proposal submission where the proposal can be
executed. If set to 0, then the proposal is allowed to be executed immediately
on submission (using the `TRY_EXEC` option). Obviously, MinExecutionPeriod
cannot be greater than VotingPeriod+MaxExecutionPeriod (where MaxExecution is
the app-defined duration that specifies the window after voting ended where a
proposal can be executed).

### Percentage decision policy

A percentage decision policy is similar to a threshold decision policy, except
that the threshold is not defined as a constant weight, but as a percentage.
It's more suited for groups where the group members' weights can be updated, as
the percentage threshold stays the same, and doesn't depend on how those member
weights get updated.

Same as the Threshold decision policy, the percentage decision policy has the
two VotingPeriod and MinExecutionPeriod parameters.

## Proposal

Any member(s) of a group can submit a proposal for a group policy account to decide upon.
A proposal consists of a set of messages that will be executed if the proposal
passes as well as any metadata associated with the proposal.

### Voting

There are four choices to choose while voting - yes, no, abstain and veto. Not
all decision policies will take the four choices into account. Votes can contain some optional metadata.
In the current implementation, the voting window begins as soon as a proposal
is submitted, and the end is defined by the group policy's decision policy.

### Withdrawing Proposals

Proposals can be withdrawn any time before the voting period end, either by the
admin of the group policy or by one of the proposers. Once withdrawn, it is
marked as `PROPOSAL_STATUS_WITHDRAWN`, and no more voting or execution is
allowed on it.

### Aborted Proposals

If the group policy is updated during the voting period of the proposal, then
the proposal is marked as `PROPOSAL_STATUS_ABORTED`, and no more voting or
execution is allowed on it. This is because the group policy defines the rules
of proposal voting and execution, so if those rules change during the lifecycle
of a proposal, then the proposal should be marked as stale.

### Tallying

Tallying is the counting of all votes on a proposal. It happens only once in
the lifecycle of a proposal, but can be triggered by two factors, whichever
happens first:

* either someone tries to execute the proposal (see next section), which can
  happen on a `Msg/Exec` transaction, or a `Msg/{SubmitProposal,Vote}`
  transaction with the `Exec` field set. When a proposal execution is attempted,
  a tally is done first to make sure the proposal passes.
* or on `EndBlock` when the proposal's voting period end just passed.

If the tally result passes the decision policy's rules, then the proposal is
marked as `PROPOSAL_STATUS_ACCEPTED`, or else it is marked as
`PROPOSAL_STATUS_REJECTED`. In any case, no more voting is allowed anymore, and the tally
result is persisted to state in the proposal's `FinalTallyResult`.

### Executing Proposals

Proposals are executed only when the tallying is done, and the group account's
decision policy allows the proposal to pass based on the tally outcome. They
are marked by the status `PROPOSAL_STATUS_ACCEPTED`. Execution must happen
before a duration of `MaxExecutionPeriod` (set by the chain developer) after
each proposal's voting period end.

Proposals will not be automatically executed by the chain in this current design,
but rather a user must submit a `Msg/Exec` transaction to attempt to execute the
proposal based on the current votes and decision policy. Any user (not only the
group members) can execute proposals that have been accepted, and execution fees are
paid by the proposal executor.
It's also possible to try to execute a proposal immediately on creation or on
new votes using the `Exec` field of `Msg/SubmitProposal` and `Msg/Vote` requests.
In the former case, proposers signatures are considered as yes votes.
In these cases, if the proposal can't be executed (i.e. it didn't pass the
decision policy's rules), it will still be opened for new votes and
could be tallied and executed later on.

A successful proposal execution will have its `ExecutorResult` marked as
`PROPOSAL_EXECUTOR_RESULT_SUCCESS`. The proposal will be automatically pruned
after execution. On the other hand, a failed proposal execution will be marked
as `PROPOSAL_EXECUTOR_RESULT_FAILURE`. Such a proposal can be re-executed
multiple times, until it expires after `MaxExecutionPeriod` after voting period
end.

## Pruning

Proposals and votes are automatically pruned to avoid state bloat.

Votes are pruned:

* either after a successful tally, i.e. a tally whose result passes the decision
  policy's rules, which can be trigged by a `Msg/Exec` or a
  `Msg/{SubmitProposal,Vote}` with the `Exec` field set,
* or on `EndBlock` right after the proposal's voting period end. This applies to proposals with status `aborted` or `withdrawn` too.

whichever happens first.

Proposals are pruned:

* on `EndBlock` whose proposal status is `withdrawn` or `aborted` on proposal's voting period end before tallying,
* and either after a successful proposal execution,
* or on `EndBlock` right after the proposal's `voting_period_end` +
  `max_execution_period` (defined as an app-wide configuration) is passed,

whichever happens first.
