# ADR 057: App Wiring

## Changelog

* 2022-05-04: Initial Draft

## Status

PROPOSED Partially Implemented

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

### Dependency Injection

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
* module-scoped dependencies via `ModuleKey`s (where each module gets a unique dependency)
* one-per-module dependencies through the `OnePerModuleType` tag interface

Here are some examples of how these would be used in an SDK module:
* `StoreKey` could be a module-scoped dependency which is unique per module
* a module's `AppModule` instance (or the equivalent) could be a `OnePerModuleType`
* CLI commands could be provided with `AutoGroupType`s

### Declarative App Config

In order to compose modules into an app, a declarative app configuration will be used. This configuration is based off
of protobuf and its basic structure is very simple:

```protobuf
package cosmos.app.v1;

message Config {
  repeated ModuleConfig modules = 1;
}

message ModuleConfig {
  string name = 1;
  google.protobuf.Any config = 2;
}
```

The configuration for every module is itself a protobuf message and modules will be identified and loaded based
on the protobuf type URL of their config object (ex. `cosmos.bank.module.v1.Module`). Modules are given a unique short `name`
to share resources across different versions of the same module which might have a different protobuf package
versions (ex. `cosmos.bank.module.v2.Module`).

An example app config in YAML might look like this:
```yaml
modules:
  - name: baseapp
    config:
      "@type": cosmos.baseapp.module.v1.Module
      begin_blockers: [staking, auth, bank]
      end_blockers: [bank, auth, staking]
      init_genesis: [bank, auth, staking]
  - name: auth
    config:
      "@type": cosmos.auth.module.v1.Module
      bech32_prefix: "foo"
  - name: bank
    config:
      "@type": cosmos.bank.module.v1.Module
  - name: staking
    config:
      "@type": cosmos.staking.module.v1.Module
```

In the above example, there is a hypothetical `baseapp` module which contains the information around ordering of
begin blockers, end blockers, and init genesis. Rather than lifting these concerns up to the module config layer,
they are themselves handled by a module which could allow a convenient way of swapping out different versions of
baseapp (for instance to target different versions of tendermint), without needing to change the rest of the config.
The `baseapp` module would then provide to the server framework (which sort of sits outside the ABCI app) an instance
of `abci.Application`.

In this model, an app is *modules all the way down* and the dependency injection/app config layer is very much
protocol-agnostic and can adapt to even major breaking changes at the protocol layer.

### Module & Protobuf Registration

In order for the two components of dependency injection and declarative configuration to work together as described,
we need a way for modules to actually register themselves and provide dependencies to the container.

One additional complexity that needs to be handled at this layer is protobuf registry initialization. Recall that
in both the current SDK `codec` and the proposed [ADR 054: Protobuf Semver Compatible Codegen](https://github.com/cosmos/cosmos-sdk/pull/11802),
protobuf types need to be explicitly registered. Given that the app config itself is based on protobuf and
uses protobuf `Any` types, protobuf registration needs to happen before the app config itself can be decoded. Because
we don't know which protobuf `Any` types will be needed a priori and modules themselves define those types, we need
to decode the app config in separate phases:
1. parse app config JSON/YAML as raw JSON and collect required module type URLs (without doing proto JSON decoding)
2. build a [protobuf type registry](https://pkg.go.dev/google.golang.org/protobuf@v1.28.0/reflect/protoregistry) based
   on file descriptors and types provided by each required module
3. decode the app config as proto JSON using the protobuf type registry

Because in [ADR 054: Protobuf Semver Compatible Codegen](https://github.com/cosmos/cosmos-sdk/pull/11802), each module
should use `internal` generated code which is not registered with the global protobuf registry, this code should provide
an alternate way to register protobuf types with a type registry.

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
