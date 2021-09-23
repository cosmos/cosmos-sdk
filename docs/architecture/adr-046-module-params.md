# ADR 046: Module Params

## Changelog

- Sep 22, 2021: Initial Draft

## Status

DRAFT

## Abstract

This ADR describes an alternative approach to how Cosmos SDK modules use, interact,
and store their respective parameters.

## Context

Currently, in the Cosmos SDK, modules that require the use of parameters use the
`x/params` module. The `x/params` works by having modules define parameters,
typically via a simple `Params` structure, and registering that structure in
the `x/params` module via a unique `Subspace` that belongs to the respective
registering module. The registering module then has unique access to its respective
`Subspace`. Through this `Subspace`, the module can get and set its `Params`
structure.

In addition, the Cosmos SDK's `x/gov` module has direct support for changing
parameters on-chain via a `ParamChangeProposal` governance proposal type, where
stakeholders can vote on suggested parameter changes.

There are various tradeoffs to using the `x/params` module to manage individual
module parameters. Namely, managing parameters essentially comes for "free" in
that developers only need to define the `Params` struct, the `Subspace`, and the
various auxiliary functions, e.g. `ParamSetPairs`, on the `Params` type. However,
there are some notable drawbacks. These drawbacks include the fact that parameters
are serialized in state via JSON which is extremely slow. In addition, parameter
changes via `ParamChangeProposal` governance proposals are _stateless_. In other
words, it is currently not possible to have any state transitions in the
application during an attempt to change param(s).

## Decision

We will build off of the alignment of `x/gov` and `x/authz` per [#9810](https://github.com/cosmos/cosmos-sdk/pull/9810). Namely, module developers will create
one or more unique parameter structures

## Consequences

> This section describes the resulting context, after applying the decision. All consequences should be listed here, not just the "positive" ones. A particular decision may have positive, negative, and neutral consequences, but all of them affect the team and project in the future.

### Backwards Compatibility

> All ADRs that introduce backwards incompatibilities must include a section describing these incompatibilities and their severity. The ADR must explain how the author proposes to deal with these incompatibilities. ADR submissions without a sufficient backwards compatibility treatise may be rejected outright.

### Positive

- Module parameters are serialized more efficiently
- Module parameters changes are able to be stateful

### Negative

- Module parameter UX becomes slightly more burdensome for developers:
    - Modules are now responsible for persisting and retrieving parameter state
    - Modules are now required to have unique message handlers to handle parameter
      changes per unique parameter data structure.

### Neutral

- Requires [#9810](https://github.com/cosmos/cosmos-sdk/pull/9810) to be reviewed
  and merged.

<!-- ## Further Discussions

While an ADR is in the DRAFT or PROPOSED stage, this section should contain a summary of issues to be solved in future iterations (usually referencing comments from a pull-request discussion).
Later, this section can optionally list ideas or improvements the author or reviewers found during the analysis of this ADR. -->

## References

- https://github.com/cosmos/cosmos-sdk/pull/9810
- https://github.com/cosmos/cosmos-sdk/issues/9438
- https://github.com/cosmos/cosmos-sdk/discussions/9913
