# ADR 067: Simulator v2

## Changelog

* June 01, 2023: Initial Draft (@alexanderbez)

## Status

DRAFT

## Abstract

The Cosmos SDK simulator is a tool that allows developers to test the entirety
of their application's state machine through the use of pseudo-randomized "operations",
which represent transactions. The simulator also provides primitives that ensures
there are no non-determinism issues and that the application's state machine can
be successfully exported and imported using randomized state.

The simulator has played an absolutely critical role in the development and testing
of the Cosmos Hub and all the releases of the Cosmos SDK after the launch of the
Cosmos Hub. Since the Hub, the simulator has relatively not changed much, so it's
overdue for a revamp.

## Context

The simulator today acts as a semi-fuzz testing suite that takes in an integer
that represents a seed into a PRNG. The PRNG is used to generate a sequence of
"operations" that are meant to reflect transactions that an application's state
machine can process. Through the use of the PRNG, all aspects of block production
and consumption are randomized. This includes a block's proposer, the validators
who both sign and miss the block, along with the transaction operations themselves.

Each Cosmos SDK module defines a set of simulation operations that _attempt_ to
produce valid transactions, e.g. `x/distribution/simulation/operations.go`. These
operations can sometimes fail depending on the accumulated state of the application
within that simulation run. The simulator will continue to generate operations
until it has reached a certain number of operations or until it has reached a
fatal state, reporting results. This gives the ability for application developers
to reliably execute full range application simulation and fuzz testing against
their application.

However, there are a few major drawbacks. Namely, with the advent of ABCI++, specifically
`FinalizeBlock`, the internal workings of the simulator no longer comply with how
an application would actually perform. Specifically, operations are executed
_after_ `FinalizeBlock`, whereas they should be executed _within_ `FinalizeBlock`.

Additionally, the simulator is not very extensible. Developers should be able to
easily define and extend the following:

* Consistency or validity predicates (what are known as invariants today)
* Property tests of state before and after a block is simulated

In addition, we also want to achieve the following:

* Consolidated weight management, i.e. define weights within the simulator itself
  via a config and not defined in each module
* Observability of the simulator's execution, i.e. have easy to understand output/logs
  with the ability to pipe those logs into some external sink
* Smart replay, i.e. the ability to not only rerun a simulation from a seed, but
  also the ability to replay from an arbitrary breakpoint
* Run a simulation based off of real network state

## Decision

> This section describes our response to these forces. It is stated in full
> sentences, with active voice. "We will ..."
> {decision body}

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

* {reference link}
