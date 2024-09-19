# ADR 76: Valset Manager

## Changelog

* 19-09-2024: Created

## Status

PROPOSED: Not Implemented


## Abstract

The downfalls of the current staking module is due to the complexity and extra work that it does to provide features that many do not use. The ValsetManager aims to separate the current module into two. First the valset management module and then the sybil resistance mechanism as a plugin into the valset management module. This separation allows chains to have a generalized valset manager which handles the logic of providing data to the consensus engine. The sybil resistance plugins can be adapted, modified or rewritten to suit the needs of the chain.

## Context

Many chains use the staking module out of the box because it is available. The staking module was completed in 2019 for the Cosmos Hub, it was designed as one of the earliest Proof of Stake systems, since then there have been many learnings about the Proof of Stake experiment. Today, if a user would like to implement a sybil resistance mechanism they use the staking module as inspiration or wrap the staking module due to the feature set of the module. This complexity comes at a price, the valset manager is a small module that is meant to provide simple primitives for users to use in order to develop simple or complex sybil resistance mechanisms.

The ValsetManager is a generalized staking module that allows teams to build sybil resistance mechanisms without having to implement the entire staking module.

## Alternatives

* Implement new sybil resistance mechanisms every time a team would like to use one, from the ground up.
* Iterate on the current staking module without allowing for custom sybil resistance mechanisms out of the box.

## Decision

The decision is to separate the current staking module into two. First the valset management module and then the sybil resistance mechanism as a plugin into the valset management module.

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

## Test Cases [optional]

Test cases for an implementation are mandatory for ADRs that are affecting consensus
changes. Other ADRs can choose to include links to test cases if applicable.

## References

* {reference link}
