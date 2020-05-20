# ADR 024: Coin Metadata

## Changelog

- 05/19/2020: Initial draft

## Status

Proposed

## Context

Assets in the Cosmos SDK are represented via a `Coins` type that consists of an
`amount` and a `denom`, where the `amount` can be any arbitrarily large or small
value. In addition, the Cosmos SDK uses an account-based model where there are
two types of primary accounts -- basic accounts and module accounts. All account
types have a set of balances that are composed of `Coins`. The `x/bank` module
keeps track of all balances for all accounts and also keeps track of the total
supply of balances in an application.

With regards to a balance `amount`, the Cosmos SDK assumes a static and fixed
unit of denomination, regardless of the denomination itself. In other words,
clients and apps built atop a Cosmos-SDK-based chain may chose to define and use
arbitrary units of denomination to provide a richer UX, however, by the time a tx
or operation reaches the Cosmos SDK state machine, the `amount` is treated as a
single unit. For example, for the Cosmos Hub (Gaia), clients assume 1 ATOM = 10^6 uatom,
and so all txs and operations in the Cosmos SDK work off of units of 10^6.

This clearly provides a poor and limited UX especially as interoperability of
networks increases and as a result the total amount of asset types increases. We
propose to have `x/bank` additionally keep track of metadata per `denom` in order
to help clients, wallet providers, and explorers improve their UX and remove the
requirement for making any assumptions on the unit of denomination.

## Decision

> This section describes our response to these forces. It is stated in full sentences, with active voice. "We will ..."
> {decision body}

## Consequences

> This section describes the resulting context, after applying the decision. All consequences should be listed here, not just the "positive" ones. A particular decision may have positive, negative, and neutral consequences, but all of them affect the team and project in the future.

### Positive

{positive consequences}

### Negative

{negative consequences}

### Neutral

{neutral consequences}

## References

- {reference link}
