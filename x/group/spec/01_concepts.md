<!--
order: 1
-->

# Concepts

## Group

A group is simply an aggregation of accounts with associated weights. It is not
an account and doesn't have a balance. It doesn't in and of itself have any
sort of voting or decision weight. It does have an "administrator" which has
the ability to add, remove and update members in the group. Note that a
group policy account could be an administrator of a group.

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
proposals.

All decision policies generally would have a minimum and maximum voting window.
The minimum voting window is the minimum amount of time that must pass in order
for a proposal to potentially pass, and it may be set to 0. The maximum voting
window is the maximum time that a proposal may be voted on before it is closed.
Both of these values must be less than a chain-wide max voting window parameter.

### Threshold decision policy

A threshold decision policy defines a threshold of yes votes (based on a tally
of voter weights) that must be achieved in order for a proposal to pass. For
this decision policy, abstain and veto are simply treated as no's.

## Proposal

Any member of a group can submit a proposal for a group policy account to decide upon.
A proposal consists of a set of messages that will be executed if the proposal
passes as well as any metadata associated with the proposal.

## Voting

There are four choices to choose while voting - yes, no, abstain and veto. Not
all decision policies will support them. Votes can contain some optional metadata.
During the voting window, accounts that have already voted may change their vote.
In the current implementation, the voting window begins as soon as a proposal
is submitted.

## Executing Proposals

Proposals will not be automatically executed by the chain in this current design,
but rather a user must submit a `Msg/Exec` transaction to attempt to execute the
proposal based on the current votes and decision policy.
It's also possible to try to execute a proposal immediately on creation or on
new votes using the `Exec` field of `Msg/CreateProposal` and `Msg/Vote` requests.
In the former case, proposers signatures are considered as yes votes.
For now, if the proposal can't be executed, it'll still be opened for new votes and
could be executed later on.

### Changing Group Membership

In the current implementation, changing a group's membership (adding or removing members or changing their weight)
will cause all existing proposals for group policy accounts linked to this group
to be invalidated. They will simply fail if someone calls `Msg/Exec` and will
eventually be garbage collected.