# ADR 057: App Wiring

## Changelog

* 2022-05-04: Initial Draft

## Status

PROPOSED Not Implemented

## Abstract

> "If you can't explain it simply, you don't understand it well enough." Provide a simplified and layman-accessible explanation of the ADR.
> A short (~200 word) description of the issue being addressed.

## Context

A number of factors have made the SDK and SDK apps in their current state hard to maintain. A symptom of the current
state of complexity is [`simapp/app.go`](https://github.com/cosmos/cosmos-sdk/blob/c3edbb22cab8678c35e21fe0253919996b780c01/simapp/app.go)
which contains almost 100 lines of imports and is otherwise over 600 lines of mostly boilerplate code that is
generally copied to each new project. (Not to mention the additional boilerplate which gets copied in `simapp/simd`.)

The large amount of boilerplate needed to bootstrap an app has made it hard to release independently versioned go
modules for Cosmos SDK modules as described in [ADR 053: Go Module Refactoring](./adr-053-go-module-refactoring.md).

In addition to being very verbose and repetitive, `app.go` also exposes a large surface area for breaking changes
as most modules instantiate themselves with positional parameters which forces breaking changes anytime a new parameter
(even an optional one) is needed.

Several attempts were made to improve the current situation including [ADR 033: Internal-Module Communication](./adr-033-protobuf-inter-module-comm.md)
and [a proof-of-concept of a new SDK](https://github.com/allinbits/cosmos-sdk-poc). The discussions around these
designs led to the current solution described here.

## Decision

In order to improve the current situation, a new "app wiring" paradigm has been designed to replace `app.go` which
involves:
* declaration configuration of the modules in an app which can be serialized to JSON or YAML
* a dependency-injection (DI) framework for instantiating apps from the that configuration

## Dependency Injection

When examining the code in `app.go` most of the code simply instantiates modules with dependencies provided either
by the framework (such as store keys) or by other modules (such as keepers). It is generally pretty obvious given
the context what the correct dependencies actually should be, so dependency-injection is an obvious solution. Rather
than making developers manually resolve dependencies, a module will tell the DI container what dependency it needs
and the container will figure out how to provide it.

We explored several existing DI solutions in golang and felt that the reflection-based approach in [uber/dig](https://github.com/uber-go/dig)
was closest to what we needed but not quite there. Assessing what we needed for the SDK, we designed and built
the Cosmos SDK [container module](https://pkg.go.dev/github.com/cosmos/cosmos-sdk/container), which has the following
features:
* dependency resolution and provision through functional constructors, ex: `func(need SomeDep) (AnotherDep, error)`
* dependency injection `In` and `Out` structs which support `optional` dependencies
* grouped-dependencies (many-per-container) through the `AutoGroupType` tag interface
* module-scoped dependencies via `ModuleKey`s
* one-per-module dependencies through the `OnePerModuleType` tag interface

Here are some examples of how these would be used in an SDK module:
*

## Consequences

> This section describes the resulting context, after applying the decision. All consequences should be listed here, not just the "positive" ones. A particular decision may have positive, negative, and neutral consequences, but all of them affect the team and project in the future.

### Backwards Compatibility

> All ADRs that introduce backwards incompatibilities must include a section describing these incompatibilities and their severity. The ADR must explain how the author proposes to deal with these incompatibilities. ADR submissions without a sufficient backwards compatibility treatise may be rejected outright.

### Positive

{positive consequences}

### Negative

{negative consequences}

### Neutral

{neutral consequences}

## Further Discussions

While an ADR is in the DRAFT or PROPOSED stage, this section should contain a summary of issues to be solved in future iterations (usually referencing comments from a pull-request discussion).
Later, this section can optionally list ideas or improvements the author or reviewers found during the analysis of this ADR.

## Test Cases [optional]

Test cases for an implementation are mandatory for ADRs that are affecting consensus changes. Other ADRs can choose to include links to test cases if applicable.

## References

* {reference link}
