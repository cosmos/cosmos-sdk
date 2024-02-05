# ADR 071: State Rent for Unclaimed Accounts

## Changelog

* Feb 5, 2024: Initial Draft (@alexanderbez)

## Status

DRAFT

## Abstract

We propose a mechanism for state rent to be applied to unclaimed accounts/balances.
This mechanism can be extended to existing accounts/balances as well with some minor
revisions.

## Context

As of SDK 0.50.x (Eden), including all previous versions, when an existing account
sends tokens to a new account, i.e. an account that does not exist on-chain yet,
the new account is created automatically via `SetAccount` as a byproduct of sending
tokens.

This might seem inconsequential, but it has a few major implications. The primary
being that users can frontrun account creation by possibly knowing the address
ahead of time and creating the account before the sender sends tokens. Another
implication is that this can drastically bloat state depending on the parameters
of the chain (e.g. min fees).

With the advent of [#19188](https://github.com/cosmos/cosmos-sdk/pull/19188), the
execution flow is changed such that the account is not created automatically. Instead,
the balance is set on the recipient's address. Then as soon as the recipient sends
their first transaction, the account is created.

However, even with this improvement in execution flow, there are still DoS and
state bloat implications. E.g. iterating over a large balance array can drastically
degrade the performance of a chain in `BeginBlock` and `EndBlock`.

Thus, we propose a mechanism for charging state rent on yet-to-be-claimed accounts/balances.
This mechanism can further be extended and adapted to charge state rent for existing
objects on-chain.

## Decision

We propose to adopt a similar mechanism to Solana's [state rent algorithm](https://docs.solanalabs.com/implemented-proposals/rent).

## Consequences

> This section describes the resulting context, after applying the decision. All
> consequences should be listed here, not just the "positive" ones. A particular
> decision may have positive, negative, and neutral consequences, but all of them
> affect the team and project in the future.

### Backwards Compatibility

> All ADRs that introduce backwards incompatibilities must include a section
> describing these incompatibilities and their severity. The ADR must explain
> how the author proposes to deal with these incompatibilities. ADR submissions
> without a sufficient backwards compatibility treatise may be rejected outright.

## Alternatives

> This section describes alternative designs to the chosen design. This section
> is important and if an adr does not have any alternatives then it should be
> considered that the ADR was not thought through.

### Positive

> {positive consequences}

### Negative

> {negative consequences}

### Neutral

> {neutral consequences}

## Further Discussions

> While an ADR is in the DRAFT or PROPOSED stage, this section should contain a
> summary of issues to be solved in future iterations (usually referencing comments
> from a pull-request discussion).
>
> Later, this section can optionally list ideas or improvements the author or
> reviewers found during the analysis of this ADR.

## References

* https://github.com/cosmos/cosmos-sdk/pull/19188
* https://docs.solanalabs.com/implemented-proposals/rent
