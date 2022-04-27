# ADR 054: Protobuf Semver Compatible Codegen

## Changelog

* 2022-04-27: First draft

## Status

PROPOSED

## Abstract

> "If you can't explain it simply, you don't understand it well enough." Provide a simplified and layman-accessible explanation of the ADR.
> A short (~200 word) description of the issue being addressed.

## Context

There has been [a fair amount of desire](https://github.com/cosmos/cosmos-sdk/discussions/10162)
in the community for semantic versioning in the SDK. How this interacts
with protobuf generated code is [more complex](https://github.com/cosmos/cosmos-sdk/discussions/10162#discussioncomment-1363034)
than it seems at first glance.

### Problem

Consider we have a Cosmos SDK module `foo` at a go module v1.x semantic version,
and in that go module there is some generated protobuf code in the protobuf
package `foo.v1` in that module’s  `types` package that forms part of the public
API, for instance as part of `Keeper` methods, ex:
```go
// foo/keeper.go

import "foo/types"

type Keeper interface {
  DoSomething(context.Context, *[]types.Coin)
}
```

Now consider that the developer wants to make a breaking change in `foo` that
*does not affect* the `Keeper` interface above at all. So now a new go module
`foo/v2` must be created to follow go semantic versioning.

The most obvious choice is to move all the code to `foo/v2` including the keeper
so we now have:

```go
// foo/v2/keeper.go

import v2types "foo/v2/types"

type Keeper interface {
  DoSomething(context.Context, *[]v2types.Coin)
}
```

Note that in this scenario `foo/v2/types` *still* refers to the protobuf package
`foo.v1`!

Now consider, another go module `bar` with a state machine also at v1.x that
depends on `foo.Keeper`. Let’s say that `bar` doesn’t have any changes but an
app wants to use `foo/v2` together with `bar`.

Now `bar` can’t be instantiated together with `foo/v2` because it requires the 
`foo` v1.x keeper which depends on the generated code in `foo` v1. Even if the
`foo/v2` `Keeper` interface is basically identical to v1 interface, we must
refactor `bar` to work with the new `foo/v2` `Keeper`. And now consumers of
`bar` will also be forced to upgrade to `foo/v2` to get any `bar` updates.
That’s not really a good outcome because it could cause all sorts of downstream
compatibility problems

To avoid this, we could consider an alternative where `foo/v2` imports `foo` v1
and exposes the `foo` v1 keeper interface to be compatible. Then `foo/v2` will
need to import the `foo` v1 go module then we have these problems:
* there will be a protobuf namespace conflict because we have the protobuf
package `foo.v1` generated into both `foo/types` and `foo/v2/types` in the
same binary. The startup panic can be disabled with a build flag, but that is a
rather hacky solution.
* `foo/v2` will need to implement a wrapper which converts the `foo/types`
structs to `foo/v2/types` structs which is just unnecessary because they both 
represent the *same* protobuf types in `foo.v1`.

One alternative at this point is to simply tell people not to use go semantic
versioning at all and to keep their packages on `v0.x` forever. This solution,
however, would likely be highly unpopular.

### API Module Approach

The first proposed solution was to generate a separate API go module (see
https://github.com/cosmos/cosmos-sdk/discussions/10582). This solution alone,
however, introduces other complexities:

#### Proto File Versioning

Versioning of proto files vs the state machine is complicated. The idea with the `api` module is that it stays
forever on `v1.x` and independent proto packages are versioned with package suffixes `v1beta1`, `v1`, `v2`.

This introduces a bit of a complex versioning scenario where we have a `v1` go module with sub-packages marked alpha or
beta that could be changed or removed. We can deal with this by allowing an explicit exception for packages marked alpha
and beta where those can change independent of the go module version.

Also it requires state machines to solidify a v1 in a beta or alpha package and then a migration of 1) all the proto
files to v1 and 2) all the state machine code to now reference v1 in two steps at the end of development. If we
call `api` v1 and we have a sub-package `foo/v1` then we really can't break `foo/v1` anymore even if the `foo/v1` state
machine isn't ready.

**Essentially this makes state machine development more complicated because it prematurely leaks WIP designs to client
libraries.**

#### Pinned proto images

As is described in  https://github.com/cosmos/cosmos-sdk/discussions/10582), we
would need to pin proto images for state machines in the codec registry
to ensure proper unknown field filtering. This isn't a deal breaker, but it is a bit complex for people to understand
and requires both:

* an additional build step to configure which may be hard to understand
* fairly complex linting of pinned proto images vs runtime generated
  code: https://github.com/cosmos/cosmos-sdk/blob/fdd3d07a28f662c2fabd6007a4a1f64be3a373b3/proto/cosmos/app/v1alpha1/module.proto#L48-L83

#### Complex refactoring of interfaces

As described in https://github.com/cosmos/cosmos-sdk/discussions/10368, all
interface methods on generated code would need to be removed and refactored into
a handler approach. It is a complex refactoring especially of `ValidateBasic`
and alignment on an ideal replacement pattern hasn't been reached yet
(https://github.com/cosmos/cosmos-sdk/pull/11340).

#### Potential limitations to generated code

In doing benchmarking of ORM generated code several potential optimizations have emerged and obviously the most
performant optimization would be to generate as much code as possible. But because ORM table descriptors may evolve from
one version to the next with new indexes and fields being added, this makes generated code state machine breaking.
Recall that the API module paradigm requires no state machine breaking logic in generated code.

## Decision

To address these issues, we will adopt a hybrid approach which requires changes
to the https://github.com/cosmos/cosmos-proto code generator.

For client facing code, the API module pattern is still useful given that
care is taken to only publish stable tags when packages are actually stable
and in production.

For state machine code, the recommended solution is to generate protobuf
code used by the module in an `internal` package and modify the code generator
to *skip global registration* for generated code in internal packages.

Because there are plenty of cases where we want one module to essentially
be a client of another module, either in `Keeper` method or as proposed in
[ADR 033](./adr-033-protobuf-inter-module-comm.md), modules will need
to use different generated code - for instance from an API module - than
the code used in the state machine itself.

To address this, we will also modify the code generator to generate interfaces
which allow different concrete structs to be passed into keeper methods and
ADR 033 than those used internally by the module.

For instance, for a struct `MsgSend`, getters like `GetAmount`, `GetFromAddress`, are already defined so codegen could
create an interface:

```go
type MsgSendI interface {
  protoreflect.ProtoMessage
  GetFromAddress() string
  GetToAddress() string
  GetAmount() []v1beta1.CoinI
  
  IsCosmosBankV1MsgSend() // this method distinguishes this interface from other proto types which may otherwise implement the same fields.
}
```

The gRPC generated server and client could look like this, with the client having
a way to detect the server revision (as described
in https://github.com/cosmos/cosmos-sdk/blob/fdd3d07a28f662c2fabd6007a4a1f64be3a373b3/proto/cosmos/app/v1alpha1/module.proto#L55)
to see which features the server supports if the client is newer:

```go
type MsgServer interface {
  Send(context.Context, MsgSendI) (MsgSendResponseI, error)
}

type MsgClient interface {
  Send(ctx context.Context, in MsgSendI, opts …grpc.CallOption) (MsgSendResponseI, error)
  // note that for MsgSendResponseI we would need to have a generated response wrapper type to deal with the case where
  // the client is newer than the server

  GetServerRevision() uint64
}
```

This would allow inter-module clients to use different generated code built against potentially different revisions of
the same protobuf package without doing marshaling/unmarshaling between different structs.

The structs themselves and the ORM could also use these interfaces directly so that the there is no need to do copying
from interfaces to concrete types when dealing with different generated code, ex:

```go
type MsgSend struct {
  FromAddress string
  ToAddress string
  Amount []v1beta1.CoinI
}

// for the ORM:
type BalanceTable interface {
	Insert(ctx context.Context, balance BalanceI) error
  ...
}
```

## Consequences

### Backwards Compatibility

This approach is more backwards compatible with the existing state machine
code than before, but some changes will be necessitated - first migrating
to the new code generator and secondly modifying `Keeper` methods to use
interfaces rather than concrete structs.

### Positive

* no need for pinned proto images in state machines - the codegen would have the correct image
* interface methods can be defined on generated code
* codegen can include other state machine breaking logic, such ORM optimizations

### Negative

* file descriptors will need to be manually registered (which is already sort of the case with `RegisterInterfaces`)
* significant changes to the code generator will be required

### Neutral


## Further Discussions

Further discussions can take place in https://github.com/cosmos/cosmos-sdk/discussions/10582 and within
the Cosmos SDK Framework Working Group.

## References

* https://github.com/cosmos/cosmos-sdk/discussions/10162
* https://github.com/cosmos/cosmos-sdk/discussions/10582
* https://github.com/cosmos/cosmos-sdk/discussions/10368
* https://github.com/cosmos/cosmos-sdk/pull/11340