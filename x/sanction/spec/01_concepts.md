<!--
order: 1
-->

# Concepts

## Sanctioned Account

An account becomes sanctioned when a governance proposal is passed with a `MsgSanction` in it containing the account's address.
When an account is sanctioned, funds cannot be removed from it.
The funds cannot be sent to another account. They cannot be spent either (e.g. on `Tx` fees).
Funds can be sent *to* a sanctioned account, but would then be immediately frozen in that account.

Sanctioning is enforced as a restriction injected into the `x/bank` send keeper and prevents removal of funds from an account.
A sanctioned account is otherwise unchanged.

When an attempt is made to remove funds from a sanctioned account, an error is returned indicating that the account is sanctioned.

## Immediate Temporary Sanctions

Immediate Temporary Sanctions (sometimes called just "immediate sanctions" or "temporary sanctions") are possible.
They happen when a `MsgSanction` governance proposal has a large enough deposit.
They also happen if a deposit is added after proposal submittal that puts the proposal's total deposit over the threshold.

This deposit threshold is managed as a module parameter: `ImmediateSanctionMinDeposit`.
If zero or empty, immediate sanctions are not possible.
If set to the governance proposal minimum deposit or less (not recommended), all `MsgSanction` governance proposals put to a vote will create immediate temporary sanctions.
If set to more than the governance proposal minimum deposit (this is recommended), it's possible to submit a `MsgSanction` proposal without the sanctions being immediate.

Immediate temporary sanctions are associated with both the governance proposal and address in question.
They expire once the governance proposal is resolved (e.g. the voting period ends).
If the proposal passes, permanent sanctions are enacted and any temporary entries for each address are removed.
If the proposal does not pass, any temporary entries associated with that proposal are removed.

Note: The phrase "permanent sanction" is used in here as a counterpart to "temporary sanction".
It is "permanent" only in the sense that it isn't temporary.
It is *not* "permanent" in the sense that it is possible to be undone (e.g. with a `MsgUnsanction`).

## Unsanctioning

A `MsgUnsanction` can be used in a governance proposal to unsanction accounts.
Once an account is unsanctioned, it can again send or spend its funds.

## Immediate Temporary Unsanctions

Similar to immediate temporary sanctions, these are created when a `MsgUnsanction` proposal has a large enough deposit (either initially or later).

This deposit threshold is managed as a module parameter: `ImmediateUnsanctionMinDeposit`.
If zero or empty, immediate unsanctions are not possible.
If set to the governance proposal minimum deposit or less (not recommended), all `MsgUnsanction` governance proposals put to a vote will create immediate temporary unsanctions.
If set to more than the governance proposal minimum deposit (this is recommended), it's possible to submit a `MsgUnsanction` proposal without the unsanctions being immediate.

Immediate temporary unsanctions are associated with both the governance proposal and address in question.
They expire once the governance proposal is resolved (e.g. the voting period ends).
If the proposal passes, permanent sanctions are removed and any temporary entries for each address are removed.
If the proposal does not pass, any temporary entries associated with that proposal are removed.

## Unsanctionable Addresses

When creating the sanction keeper, a list of addresses of unsanctionable accounts can be provided.
An attempt to sanction or enact an immediate temporary sanction on an address in that list results in the error: `"address cannot be sanctioned"`.

An example of an account that should not be sanctionable is the fee collector.

## Params

The `x/sanction` module has some params that can be defined in state.

* `ImmediateSanctionMinDeposit` is the minimum deposit required for immediate temporary sanctions to be enacted for addresses in a `MsgSanction`.
* `ImmediateUnsanctionMinDeposit` is the minimum deposit required for immediate temporary unsanctions to be enacted for addresses in a `MsgUnsanction`.

If not defined in state, the following variables are used (defined in `x/sanction/sanction.go`):
* `DefaultImmediateSanctionMinDeposit`
* `DefaultImmediateUnsanctionMinDeposit`

By default, those have a value of `nil` which makes it impossible to enact immediate temporary sanctions or unsanctions.
They are public, though, so consuming chains can change them as desired.

The default variables are only used if the state entry does not exist.
If the entry exists, but is empty, that empty value is used.

It is recommended that both of these minimum deposits be significantly larger than the governance proposal minimum deposit.
This is to prevent malicious use of immediate temporary sanctions or unsanctions.

## Complex Interactions

It's possible to end up with some complex interactions due to multiple `MsgSanction` and `MsgUnsanction` messages with large enough deposits for temporary effects.
In a general sense, the last one takes precedence.

### Conflicting Messages in a Proposal

When a proposal has multiple messages, they are processed in the order they are listed in the proposal.
So if a governance proposal contains both a `MsgSanction` and `MsgUnsanction`, and one or more addresses are listed in both,
then, the last message they're in takes precedence.

For example, say a proposal has, a `MsgSanction` for accounts A, B, and C, then a `MsgUnsanction` for accounts B, C, and D.
And the proposal has enough of a deposit for both immediate sanctions and unsanctions.
There will be 4 temporary entries: account A will have a temporary sanction; and B, C, and D will have temporary unsanctions.
If the proposal passes, a permanent sanction will be placed on A, and accounts B, C, and D will have their sanctions removed.

If a proposal contains both a `MsgSanction` and `MsgUnsanction` and the total deposit is enough for immediate temporary entries of one type, but not the other,
the temporary entries are enacted for the one, but not the other. If later, more deposit is added so there's enoug for both, the others will then be enacted too.

### Conflicting Governance Proposals

If multiple governance proposals have immediate temporary effects, the effect from the proposal with the largest id takes precedence.
Voting period start and end times/heights are not taken into account, only the proposal id.

For example, let's say proposal 3 has a `MsgSanction` for accounts A, B, and C; and proposal 5 has a `MsgUnsanction` for accounts B, C, and D.
Both have large enough deposits for immediate effects.
There will be six temporary entries, A+3, B+3, C+3, B+5, C+5, and D+5.
But the temporary entries for accounts B and C from prop 3 are ignored because of their prop 5 entries.
In effect, account A will have a temporary sanction; and accounts, B, C, and D will have temporary unsanctions.

Scenarios:
* Prop 3 passes while prop 5 is still being voted on:
  All temporary entries for accounts A, B, and C are removed, and those accounts are permanently sanctioned.
  The only temporary entry left will be an unsanction for account D.
* Prop 3 does not pass while prop 5 is still being voted on:
  The temporary sanction entries for accounts A, B, and C are removed leaving temporary unsanction entries for B, C, and D.
  No permanent sanctions are enacted.
* Prop 5 passes while prop 3 is still being voted on:
  All temporary entries for accounts B, C, and D are removed and permanent sanctions are also removed for those accounts.
  The only temporary entry left will be a sanction on account A.
* Prop 5 does not pass while prop 3 is still being voted on:
  The temporary unsanction entries for accounts B, C, and D are removed leaving temporary sanction entries for A, B, and C.
  If accounts B, C, or D were previously permanently sanctioned, those sanctions remain.
