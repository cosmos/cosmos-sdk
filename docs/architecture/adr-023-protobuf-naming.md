# ADR 023: Protocol Buffer Naming and Versioning Conventions

## Changelog

- 2020 April 27: Initial Draft

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

For instance, `cosmos.bank.Send` (16 bytes) conveys roughly the same information
as `cosmos_sdk.x.bank.v1.MsgSend` (28 bytes) but is notably shorter.
If we needed version 2 of bank, we could just write `cosmos.bank2.Send`.

Such conciseness makes names both more pleasant to work with and take up less
space within transactions and on the wire.

We should also resist the temptation to make names cryptically short with
abbreviations. For instance, we shouldn't try to reduce `cosmos.bank.Send`
to `csm.bk.Snd` just to save six bytes.

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
from breaking. The SDK _should_ mark any packages as alpha or beta where the
API is likely to change significantly in the near future.

### Package Naming

#### Adopt a short, unique top-level package name

Top-level packages should adopt a short name that is unlikely to collide with
other names in common usage. In the near future, a registry should be created
to reserve and index top-level package names used within the Cosmos ecosystem.

#### Limit sub-package depth

Sub-package depth should be increased with caution. Generally a single
sub-package is needed for a module or a library. Even though `x` or `modules`
is used in source code to denote modules, this is often unnecessary for .proto
files as modules are the primary thing sub-packages are used for. Only items which
are known to be used infrequently should have deep sub-package depths.

### Message Naming

#### Use simple verbs for transaction messages

This is likely one of the most controversial recommendations here as it will
change legacy go struct names.

Historically all transaction messages in the SDK have been prefixed with `Msg`
which is a shortening of messsage. To begin with "message", not being an action
word, does not clearly indicate the concept of transaction or operation. So
understanding that `Msg` indicates transaction likely requires knowledge of
Cosmos SDK conventions. Once if someone understands that `Msg` means transaction,
however, it still is likely redundant information as the type name usually
includes an action verb such as "send", "create" or "submit".

Going forward, for both conciseness and clarity, transaction messages should
simply use a descriptive action verb to indicate that this is something which
performs an action in a transaction. This should also be made clear in the
documentation.

To maintain compatibility with existing names in go code, an alias can be
introduced for the previous type name.

#### Use simple nouns for state types

Nouns should be used for other types which are used in state, queries and as
arguments to some transaction messages. Ex. `Coin`, `Proposal` and `Delegation`.

### Service and RPC Naming

[ADR 021](adr-021-protobuf-query-encoding.md) specifies that modules should
implement a gRPC query service. We should consider the principle of conciseness
for query service and RPC names as these may be called within the state machine
from persistent script modules such as CosmWasm. For example, we can shorten
`/cosmos_sdk.x.bank.v1.QueryService/QueryBalance` to
`/cosmos.bank.Query/Balance` without losing much useful information.

#### Use just `Query` or `Q` for the query service

Instead of [Buf's default service suffix recommendation](https://github.com/cosmos/cosmos-sdk/pull/6033),
we should simply use the shorter `Query` or even consider just `Q` for query
services.

For other types of gRPC services, we should consider sticking with Buf's
default recommendation.

#### Omit `Get` and `Query` from query service RPC names

`Get` and `Query` should be omitted from `Query` service names because they are
redundant in the fully-qualified name. For instance, `/cosmos.bank.Query/QueryBalance`
just says `Query` twice without any new information.

## Future Improvements

## Consequences

### Positive

* names will be more concise and easier to read
* all transactions using `Any` will be at least 12 bytes shorter
(`_sdk.x`, `.v1`, `Msg` will be removed from `sdk.Msg` message names)
* query RPC names also be more concise and easy to read
* `.proto` file imports can be more standard (without `"third_party/proto" in
the path)

### Negative

* some legacy go struct names will change (ex. `MsgSend` vs `Send`). Note that
legacy Amino type names will not change

### Neutral

* `.proto`  files will need to be reorganized and refactored

## References
