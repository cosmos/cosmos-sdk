# ADR 023: Protocol Buffer Naming and Versioning Conventions

## Changelog

- 2020 April 27: Initial Draft

## Status

Proposed

## Context

Protocol Buffers provide a basic [style guide](https://developers.google.com/protocol-buffers/docs/style)
and [Buf](https://buf.build/docs/style-guide) builds upon that. To the
extent possible, we want to follow industry accepted guidelines and wisdom for
the effective usage of protobuf, deviating from those only when there is clear
rationale for our use case.

### Adoption of `Any`

The adoption of `google.protobuf.Any` as the recommended approach for encoding
interface types (as opposed to `oneof`) makes package naming a central part
of the encoding as fully-qualified message names now appear in encoded
messages.

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

The goal of this ADR is to provide thoughtful naming conventions that:

* encourage a good user experience for when users interact directly with
.proto files and fully-qualified protobuf names
* balance conciseness against the possibility of either over-optimizing (making
names too short and cryptic) or under-optimizing (just accepting bloated names
with lots of redundant information)

These guidelines are meant to act as a style guide for both the SDK and
third-party modules.

As a starting point, we should adopt all of the [DEFAULT](https://buf.build/docs/lint-checkers#default)
checkers in [Buf's](https://buf.build) including [`PACKAGE_DIRECTORY_MATCH`](https://buf.build/docs/lint-checkers#file_layout),
except:
* [PACKAGE_VERSION_SUFFIX](https://buf.build/docs/lint-checkers#package_version_suffix)
* [SERVICE_SUFFIX](https://buf.build/docs/lint-checkers#service_suffix)

Further guidelines to be described below.

### Principles

#### Concise and Descriptive Names

Names should be descriptive enough to convey their meaning and distinguish
them from other names.

Given that we are using fully-qualifed names within
`google.protobuf.Any` as well as within gRPC query routes, we should aim to
keep names concise, without going overboard. The general rule of thumb should
be if a shorter name would convey more or else the same thing, pick the shorter
name.

For instance, `cosmos.bank.MsgSend` (19 bytes) conveys roughly the same information
as `cosmos_sdk.x.bank.v1.MsgSend` (28 bytes) but is more concise.

Such conciseness makes names both more pleasant to work with and take up less
space within transactions and on the wire.

We should also resist the temptation to over-optimize, by making names
cryptically short with abbreviations. For instance, we shouldn't try to
reduce `cosmos.bank.MsgSend` to `csm.bk.MSnd` just to save a few bytes.

The goal is to make names **_concise but not cryptic_**.

#### Names are for Clients First

Package and type names should be chosen for the benefit of users, not
necessarily because of legacy concerns related to the go code-base.

#### Plan for Longevity

In the interests of long-term support, we should plan on the names we do
choose to be in usage for a long time, so now is the opportunity to make
the best choices for the future.

### Versioning

#### Don't Allow Breaking Changes in Stable Packages

Always use a breaking change detector such as [Buf](https://buf.build) to prevent
breaking changes in stable (non-alpha or beta) packages. Breaking changes can
break smart contracts/persistent scripts and generally provide a bad UX for
clients. With protobuf, there should usually be ways to extend existing
functionality instead of just breaking it.

#### Omit v1 suffix

Instead of using [Buf's recommended version suffix](https://buf.build/docs/lint-checkers#package_version_suffix),
we can omit `v1` for packages that don't actually have a second version. This
allows for more concise names for common use cases like `cosmos.bank.Send`.
Packages that do have a second or third version can indicate that with `.v2`
or `.v3`.

#### Use `alpha` or `beta` to Denote Non-stable Packages

[Buf's recommended version suffix](https://buf.build/docs/lint-checkers#package_version_suffix)
(ex. `v1alpha1`) _should_ be used for non-stable packages. These packages should
likely be excluded from breaking change detection and _should_ generally 
be blocked from usage by smart contracts/persistent scripts to prevent them
from breaking. The SDK _should_ mark any packages as alpha or beta where the
API is likely to change significantly in the near future.

### Package Naming

#### Adopt a short, unique top-level package name

Top-level packages should adopt a short name that is known to not collide with
other names in common usage within the Cosmos ecosystem. In the near future, a
registry should be created to reserve and index top-level package names used
within the Cosmos ecosystem. Because the Cosmos SDK is intended to provide
the top-level types for the Cosmos project, the top-level package name `cosmos`
is recommended for usage within the Cosmos SDK instead of the longer `cosmos_sdk`.
[ICS](https://github.com/cosmos/ics) specifications could consider a
short top-level package like `ics23` based upon the standard number.

#### Limit sub-package depth

Sub-package depth should be increased with caution. Generally a single
sub-package is needed for a module or a library. Even though `x` or `modules`
is used in source code to denote modules, this is often unnecessary for .proto
files as modules are the primary thing sub-packages are used for. Only items which
are known to be used infrequently should have deep sub-package depths.

For the Cosmos SDK, it is recommended that that we simply write `cosmos.bank`,
`cosmos.gov`, etc. rather than `cosmos.x.bank`. In practice, most non-module
types can go straight in the `cosmos` package or we can introduce a
`cosmos.base` package if needed. Note that this naming _will not_ change
go package names, i.e. the `cosmos.bank` protobuf package will still live in
`x/bank`.

### Message Naming

Message type names should be as concise possible without losing clarity. `sdk.Msg`
types which are used in transactions will retain the `Msg` prefix as that provides
helpful context.

### Service and RPC Naming

[ADR 021](adr-021-protobuf-query-encoding.md) specifies that modules should
implement a gRPC query service. We should consider the principle of conciseness
for query service and RPC names as these may be called from persistent script
modules such as CosmWasm. Also, users may use these query paths from tools like
[gRPCurl](https://github.com/fullstorydev/grpcurl). As an example, we can shorten
`/cosmos_sdk.x.bank.v1.QueryService/QueryBalance` to
`/cosmos.bank.Query/Balance` without losing much useful information.

RPC request and response types _should_ follow the `ServiceNameMethodNameRequest`/
`ServiceNameMethodNameResponse` naming convention. i.e. for an RPC method named `Balance`
on the `Query` service, the request and response types would be `QueryBalanceRequest`
and `QueryBalanceResponse`. This will be more self-explanatory than `BalanceRequest`
and `BalanceResponse`.

#### Use just `Query` for the query service

Instead of [Buf's default service suffix recommendation](https://github.com/cosmos/cosmos-sdk/pull/6033),
we should simply use the shorter `Query` for query services.

For other types of gRPC services, we should consider sticking with Buf's
default recommendation.

#### Omit `Get` and `Query` from query service RPC names

`Get` and `Query` should be omitted from `Query` service names because they are
redundant in the fully-qualified name. For instance, `/cosmos.bank.Query/QueryBalance`
just says `Query` twice without any new information.

## Future Improvements

A registry of top-level package names should be created to coordinate naming
across the ecosystem, prevent collisions, and also help developers discover
useful schemas. A simple starting point would be a git repository with
community-based governance.

## Consequences

### Positive

* names will be more concise and easier to read and type
* all transactions using `Any` will be at shorter (`_sdk.x` and `.v1` will be removed)
* `.proto` file imports will be more standard (without `"third_party/proto"` in
the path)
* code generation will be easier for clients because .proto files will be
in a single `proto/` directory which can be copied rather than scattered
throughout the SDK

### Negative

### Neutral

* `.proto`  files will need to be reorganized and refactored
* some modules may need to be marked as alpha or beta

## References
