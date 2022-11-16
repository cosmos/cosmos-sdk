# ADR 054: Protobuf Semver Compatible Codegen

## Changelog

* 2022-04-27: First draft

## Status

PROPOSED

## Abstract

TODO

## Context

There has been [a fair amount of desire](https://github.com/cosmos/cosmos-sdk/discussions/10162)
in the community for semantic versioning in the SDK. How this interacts
with protobuf generated code is [more complex](https://github.com/cosmos/cosmos-sdk/discussions/10162#discussioncomment-1363034)
than it seems at first glance.

### Problem

Consider we have a module `foo` which defines the following `MsgDoSomething` and that we've released it under
the go module `example.com/foo`:

```protobuf
package foo.v1;

message MsgDoSomething {
  string sender = 1;
  uint64 amount = 2;
}

service Msg {
  DoSomething(MsgDoSomething) returns (MsgDoSomethingResponse);
}
```

Now consider that we make a revision to this module and add a new `condition` field to `MsgDoSomething` and also
add a new validation rule on `amount` requiring it to be non-zero, and that we follow go semantic versioning and
want to release the next state machine version of `foo` as `example.com/foo/v2`.

```protobuf
// Revision 1
package foo.v1;

message MsgDoSomething {
  string sender = 1;
  
  // amount must be a non-zero integer.
  uint64 amount = 2;
  
  // condition is an optional condition on doing the thing.
  //
  // Since: Revision 1
  Condition condition = 3;
}
```

Approaching this naively, we would generate the protobuf types for the initial
version of `foo` in `example.com/foo/types` and we would generate the protobuf
types for the second version in `example.com/foo/v2/types`.

Now let's say we have a module `bar` which talks to `foo` using this keeper
interface which `foo` provides:
```go
type FooKeeper interface {
	DoSomething(MsgDoSomething) error
}
```

### Scenario A: Backward Compatibility: Newer Foo, Older Bar

Imagine we have a chain which uses both `foo` and `bar` and wants to upgrade to
`foo/v2`, but the `bar` module has not upgraded to `foo/v2`.

In this case, the chain will not be able to upgrade to `foo/v2` until `bar` 
has upgraded its references to `example.com/foo/types.MsgDoSomething` to 
`example.com/foo/v2/types.MsgDoSomething`.

Even if `bar`'s usage of `MsgDoSomething` has not changed at all, the upgrade
will be impossible without this change because `example.com/foo/types.MsgDoSomething`
and `example.com/foo/v2/types.MsgDoSomething` are fundamentally different
incompatible types in the go type system.

### Scenario B: Forward Compatibility: Older Foo, Newer Bar

Now let's consider the reverse scenario, where `bar` upgrades to `foo/v2`
by changing the `MsgDoSomething` reference to `example.com/foo/v2/types.MsgDoSomething`
and releases that as `bar/v2` with some other changes that a chain wants.
The chain, however, has decided that it thinks the changes in `foo/v2` are too
risky and that it'd prefer to stay on the initial version of `foo`.

In this scenario, it is impossible to upgrade to `bar/v2` without upgrading
to `foo/v2` even if `bar/v2` would have worked 100% fine with `foo` other
than changing the import path to `MsgDoSomething` (meaning that `bar/v2`
doesn't actually use any new features of `foo/v2`).

Now because of the way go semantic import versioning works, we are locked
into either using `foo` and `bar` OR `foo/v2` and `bar/v2`. We cannot have
`foo` + `bar/v2` OR `foo/v2` + `bar`. The go type system doesn't allow this
even if both versions of these modules are otherwise compatible with each other.

### Naive Mitigation

A naive approach at fixing this would be to not regenerate the protobuf types
in `example.com/foo/v2/types` but instead just update `example.com/foo/types`
to reflect the changes needed for `v2` (adding `condition` and requiring
`amount` to be non-zero). Then we could release a patch of `example.com/foo/types`
with this update and use that for `foo/v2`. But this change is state machine
breaking for `v1`. It requires changing the `ValidateBasic` method to reject
the case where `amount` is zero, and it adds the `condition` field which
should be rejected based on ADR 020 unknown field filtering. So adding these
changes as a patch on `v1` is actually an incorrect semantic versioning patch
to that release line. Chains that want to stay on `v1` of `foo` should not
be importing these changes because they are incorrect for `v1.`

### Solutions

### A) API Module Approach

One solution (first proposed in https://github.com/cosmos/cosmos-sdk/discussions/10582) is to isolate all protobuf generated code into a separate module
from the state machine module. This would mean that we could have state machine
go modules `foo` and `foo/v2` which could use a types or API go module say
`foo/api`. This `foo/api` go module would be perpetually on `v1.x` and only
accept non-breaking changes. This would then allow other modules to be
compatible with either `foo` or `foo/v2` as long as the inter-module API only
depends on the types in `foo/api`.

This approach introduces two complexities which need to be dealt with:
1. if `foo/api` includes any state machine breaking code then this situation is unworkable
because it means that changes to the state machine logic bleed across two go modules
in a hard to reason about manner. Currently, interfaces (such as `sdk.Msg`
or `authz.Authorization` are implemented directly on generated types and these
almost always include state machine breaking logic.
2. the version the protobuf files in `foo/api` at runtime may be different from
the version `foo` or `foo/v2` were built with which could introduce subtle bugs.
For instance, if a field is added to a message in `foo/api` for `foo/v2` but
the original `foo` module doesn't use it, then there may be important data in
that new field which silently gets ignored.

#### Migrate all interface methods on API types to handlers 

To solve 1), we need to remove all interface implementations from generated
types and start using a handler approach which essentially means that given
a type `X`, we have some sort of router which allows us to resolve some interface
implementation for that type (say `sdk.Msg` or `authz.Authorization`).

In the case of some methods on `sdk.Msg`, we can replace them with declarative
annotations. For instance, `GetSigners` can already be replaced by the protobuf
annotation `cosmos.msg.v1.signer`. In the future, we may consider some sort
of protobuf validation framework (like https://github.com/bufbuild/protoc-gen-validate
but more Cosmos-specific) to replace `ValidateBasic`.

#### Pinning State Machine API Compatibility

To solve 2), state machine modules must be able to specify what the version of
the protobuf files was that they were built against. The simplest way to
do this may be to embed the protobuf `FileDescriptor`s into the module itself
so that these `FileDescriptor`s are used at runtime rather than the ones that
are built into the `foo/api` which may be different. This would involve the
following steps:
1. a build step executed during `make proto-gen` that runs `buf build` to pack the `FileDescriptor`s into an image file
2. a go embed directive to embed the image file

One consequence of bundling protobuf types separate from state machine logic is how it affects API forwards
compatibility. By default, protobuf as we're using it is forwards compatible - meaning that newer clients can talk to
older state machines. This can cause problems, however, if fields are added to older messages and clients try to use
these new fields against older state machines.

For example, say someone adds a field to a message to set an optional expiration time on some operation. If the newer
client sends a message an older state machine with this new expiration field set, the state machine will reject it
based on [ADR 020 Unknown Field Filtering](https://github.com/cosmos/cosmos-sdk/blob/master/docs/architecture/adr-020-protobuf-transaction-encoding.md#unknown-field-filtering).

This will break down, however, if we package API types separately because an app developer may use an API version that's
newer that the state machine version and then clients can send a message with an expiration time, the state machine will
accept it but ignore it, which is a bug. This isn't a problem now because the protobuf types are codegen'ed directly
into the state machine code so there can be no discrepancy.

If we migrate to an API module, we will need to pin the API compatibility version in the state machine. Here are two
potential ways to do this:

##### "Since" Annotations

In the [Protobuf package versioning discussion](https://github.com/cosmos/cosmos-sdk/discussions/10406) we agreed to
take [approach 2: annotations using "Since" comments](https://github.com/cosmos/cosmos-sdk/pull/10434) to help clients
deal with forwards compatibility gracefully by only enabling newer API features when they know they're talking to a
newer chain. We could use these same annotations in state machines and instruct unknown field filtering to reject
fields in newer API versions if they are present in protobuf generated code.


##### Embed raw proto files in state machines

This would involve using [golang embed](https://pkg.go.dev/embed) to force state machines to embed the proto files they
intended to use in the binary and registering them with the `InterfaceRegistry` which will use them for unknown field
filtering. By using the `embed.FS` type we can force module developers to copy .proto files into the module path and
update them when they want to support newer APIs. A warning could be emitted if an older API verison is registered and
a newer one is available in generated code, alerting modular developers and/or apps to upgrade.

This second approach is probably less fragile than the "Since" annotations because it could be easy to mess up API
versions with annotations. While these annotations are a good solution for clients like CosmJs which want to support
multiple chains, we need things to be a little stricter inside state machines.

### B) Changes to Generated Code

An alternate approach to solving the versioning problem with generated code is to change the generated code itself.
Currently, we face two issues with how protobuf generated code works:
1. there can be only one generated type per protobuf type in the global registry (this can be overridden with special build flags)
2. if there are two generated types for one protobuf type, then they are not compatible

To solve the global registry problem, we can adapt the code generator to not registry protobuf types with the global
registry if the generated code is placed in an `internal/` package. This will require modules to register their types
with the app-level level protobuf registry manually. This is similar to what modules already do with registering types
with the amino codec and `InterfaceRegistry`.

Dealing with incompatible types is more difficult. Imagine we have an ADR 033 module client which is using the type
`github.com/cosmos/cosmos-sdk/x/bank/types.MsgSend`, but the bank module itself is using
`github.com/cosmos/cosmos-sdk/x/bank/v2/internal/types.MsgSend` in its `MsgServer` implementation. It won't be possible
to directly convert one `MsgSend` to the other the way protobuf types are currently generated.

If we changed protobuf generated types to only expose interfaces and then implemented the storage of the types using
some set of zero-copy memory buffers, then we could simply pass the memory buffers from one implementation of the
types to another.

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