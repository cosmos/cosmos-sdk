# ADR 023: Protocol Buffer Naming and Style Conventions

## Changelog

- 2020 April 26: Initial Draft

## Status

Proposed

## Context

Protocol Buffers provide a basic [style guide](https://developers.google.com/protocol-buffers/docs/style)
and [Buf](https://buf.build/docs/style-guide) builds upon that. To do the
extent possible we want to follow industry accepted wisdom and guidelines for
the effective use of protobuf, deviating from those only when there is clear
rationale for our use case. 

### Adoption of `Any`

The adoption of `google.protobuf.Any` as the recommended approach for encoding
interface types (as opposed to `oneof`) makes package naming a central part
of the encoding as package names now appear in encoded messages.

### Current Directory Organization

Thus far we have mostly followed [Buf's](https://buf.build) [DEFAULT](https://buf.build/docs/lint-checkers#default)
recommendations, with the minor deviation of disabling [`PACKAGE_DIRECTORY_MATCH`](https://buf.build/docs/lint-checkers#file_layout)
which although being convenient for developing code comes with the warning
from Buf that:
> you will have a very bad time with many Protobuf plugins across various languages if you do not do this

### Adoption of gRPC Queries

In [ADR 021](adr-021-protobuf-query-encoding.md), gRPC was adopted for Protobuf
native queries. The full gRPC service path thus becomes a key part of ABCI query
path. In the future, gRPC queries may be allowed from within persistent scripts
by technologies such as CosmWasm and these query routes would be stored within
script binaries.

## Decision

These guidelines are meant to act as a style guide for both the SDK and
third-party modules.

We should adopt all of the [DEFAULT](https://buf.build/docs/lint-checkers#default)
checkers in [Buf's](https://buf.build) including [`PACKAGE_DIRECTORY_MATCH`](https://buf.build/docs/lint-checkers#file_layout),
except:
* [PACKAGE_VERSION_SUFFIX](https://buf.build/docs/lint-checkers#package_version_suffix)
* [SERICE_SUFFIX](https://buf.build/docs/lint-checkers#service_suffix)

### Principles

#### Concise and Descriptive Names

Names should be descriptive enough to convey their meaning and distinguish
them from other names.

Given that we are using fully-qualifed names within
`google.protobuf.Any` as well as within gRPC query routes, we should aim to
keep names concise, without going overboard. The general rule of thumb should
be if a shorter name would convey more or else the same thing, pick the shorter
name.

For instance, `cosmos.bank.Send` conveys roughly the same information
as `cosmos_sdk.x.bank.v1.MsgSend` but is notably shorter. If we needed version 2 of
bank, we could just write `cosmos.bank2.Send`.

Such conciseness makes names both more pleasant to work with and take up less
space within transactions and on the wire.

In general, names should be **_concise but not cryptic_**.

#### Names are for Clients First

#### Plan for Longevity

### Versioning

#### Don't Allow Breaking Changes in Stable Packages

Always use a breaking change detector such as [Buf](https://buf.build) to prevent
breaking changes in stable (non-alpha or beta) packages. Breaking changes can
break persistent scripts/smart contracts and generally provide a bad UX for
clients. With protobuf, there should usually be ways to extend existing
functionality rather than just breaking it.

#### Use Simple Version Suffixes

Instead of using [Buf's recommended version suffix](https://buf.build/docs/lint-checkers#package_version_suffix),
we can more concisely indicate version 2, 3, etc. of a package with a simple
suffix rather than a sub-package, ex. `cosmos.bank2.Send` or `cosmos.bank.Send2`.
Version 1 should obviously be the non-suffixed version, ex. `cosmos.bank.Send`.

#### Use `alpha` or `beta` to Denote Non-stable Packages

[Buf's recommended version suffix](https://buf.build/docs/lint-checkers#package_version_suffix)
(ex. `v1alpha1`) _should_ be used for non-stable packages. These packages will
likely be excluded from breaking change detection and _should_ generally 
be blacklisted from usage by smart contracts/persistent scripts to prevent them
from breaking.

### Package Naming

#### Adopt a short, unique top-level package name

Top-level packages should adopt a short name that is unlikely to collide with
other names in common usage. In the near future, a registry should be created
to reserve and index top-level package names used within the Cosmos ecosystem.

#### Limit sub-package depth

Sub-package depth should be increased with caution. Generally a single
sub-package is needed for a module or a library. Even though `x` or `modules`
is used in source code to denote modules, this is often unnecessary for .proto
files. Only items which are known to be used infrequently should have 

### Message Naming

#### Use simple verbs for transaction messages

#### Use simple nouns for state

### Service and RPC Naming

#### Use just `Query` or `Q` for the query service

#### Omit `Get` and `Query` from RPC names

## Future Improvements

## Consequences

### Positive

* names will be shorter
* `.proto` file imports can be more standard (without `"third_party/proto" in
the path)

### Negative

* some proto names will diverge from their go names or the go names will have to
change (ex. `MsgSend` vs `Send`)

### Neutral

* `.proto`  files will need to be moved into a top-level `proto/` directory

## References
