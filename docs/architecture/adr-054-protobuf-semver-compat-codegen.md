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

### Problem 1: Version Compatibility

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

#### Scenario A: Backward Compatibility: Newer Foo, Older Bar

Imagine we have a chain which uses both `foo` and `bar` and wants to upgrade to
`foo/v2`, but the `bar` module has not upgraded to `foo/v2`.

In this case, the chain will not be able to upgrade to `foo/v2` until `bar` 
has upgraded its references to `example.com/foo/types.MsgDoSomething` to 
`example.com/foo/v2/types.MsgDoSomething`.

Even if `bar`'s usage of `MsgDoSomething` has not changed at all, the upgrade
will be impossible without this change because `example.com/foo/types.MsgDoSomething`
and `example.com/foo/v2/types.MsgDoSomething` are fundamentally different
incompatible types in the go type system.

#### Scenario B: Forward Compatibility: Older Foo, Newer Bar

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

#### Naive Mitigation

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

### Problem 2: Circular dependencies

None of the above approaches allow `foo` and `bar` to be separate modules
if for some reason `foo` and `bar` depend on each other in different ways.
We have several cases of circular module dependencies in the SDK
(ex. staking, distribution and slashing) that are legitimate from a state machine
perspective but that would make it impossible to independently semantically
version these modules without some other mitigation.

### Solutions

### Approach A) Separate API and State Machine Modules

One solution (first proposed in https://github.com/cosmos/cosmos-sdk/discussions/10582) is to isolate all protobuf generated code into a separate module
from the state machine module. This would mean that we could have state machine
go modules `foo` and `foo/v2` which could use a types or API go module say
`foo/api`. This `foo/api` go module would be perpetually on `v1.x` and only
accept non-breaking changes. This would then allow other modules to be
compatible with either `foo` or `foo/v2` as long as the inter-module API only
depends on the types in `foo/api`. It would also allow modules `foo` and `bar`
to depend on each other in that both of them could depend on `foo/api` and
`bar/api` without `foo` directly depending on `bar` and vice versa.

This is similar to the naive mitigation described above except that it separates
the types into separate go modules which in and of itself could break circular
dependencies. Otherwise, it would have the same problems as that solution which
we could rectify by:
1. removing all state machine breaking code from the API module (ex. `ValidateBasic` and any other interface methods)
2. embedding the correct file descriptors to be used for unknown field filtering in the binary.

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
the protobuf files was that they were built against. For instance if the API
module for `foo` upgrades to `foo/v2`, the original `foo` module still needs
a copy of the original protobuf files it was built with so that ADR 020
unknown field filtering will reject `MsgDoSomething` when `condition` is
set.

The simplest way to do this may be to embed the protobuf `FileDescriptor`s into
the module itself so that these `FileDescriptor`s are used at runtime rather 
than the ones that are built into the `foo/api` which may be different. This
would involve the following steps:
1. a build step executed during `make proto-gen` that runs `buf build` to pack the `FileDescriptor`s into an image file
2. a go embed directive to embed the image file

#### Potential limitations to generated code

One challenge with this approach is that it places heavy restrictions on what
can go in API modules and how generated code can affect state machine logic.

For instance, we do code generation for the ORM that in the future could
potentially contain optimizations that are even state machine breaking. We
would either need to ensure very carefully that the optimizations aren't
state machine breaking in generated code or separate this generated code
out from the API module into the state machine module. Both of these mitigations
are potentially viable but the API module approach does require an extra level
of care to avoid these sorts of issues.

### Approach B) Changes to Generated Code

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

For example, imagine this protobuf API with only getters and setters is exposed for `MsgSend`:
```go
type MsgSend interface {
	proto.Message
	GetFromAddress() string
	GetToAddress() string
	GetAmount() []v1beta1.Coin
    SetFromAddress(string)
    SetToAddress(string)
    SetAmount([]v1beta1.Coin)
}

func NewMsgSend() MsgSend { ... }
```

Under the hood, `MsgSend` could be implemented based on raw `[]byte` memory buffers, ex:
```go
type msgSendImpl struct {
	memoryBuffers []interface{}
}

func (m *msgSendImpl) GetFromAddress() {
	return m.memoryBuffers[0].(string)
}

func (m *msgSendImpl) GetToAddress() {
    return m.memoryBuffers[1].(string)
}
```

This approach would have the added benefits of allowing zero-copy message passing to modules written in other languages
such as Rust. It could also make unknown field filtering in inter-module communication simpler if we assume that all
new fields are added in sequential order.

Also, we wouldn't have any issues with state machine breaking code on generated types because all the generated
code used in the state machine would actually live in the state machine module. Other modules using ADR 033 or
keeper methods could still use a public API module to reference other types. In fact, it may be best to combine
the API module approach with this approach - basically public keeper APIs would use public API module types whereas
internal logic would use internal generated types. This, however, may be a bit complex for module authors to manage.

Other big downsides of this approach are that it requires big changes to how people use protobuf types and would be a
substantial rewrite of the protobuf code generator. These types, however, could still be made compatible with
the `google.golang.org/protobuf/reflect/protoreflect` API in order to work with all standard golang protobuf tooling.

It is possible that this second approach could be adopted later on as an optimization for multi-language message
on top of the first API module approach.

### Approach C) Use 0.x based versioning and replace directives

Some people have commented that go's semantic import versioning (i.e. changing the import path to `foo/v2`, `foo/v3`,
etc.) is too restrictive and that it should be optional. The golang maintainers disagree and only officially support
semantic import versioning, although we could take the contrary perspective and get more flexibility by using 0.x-based
versioning basically forever.

Module version compatibility could then be achieved using go.mod replace directives to pin dependencies to specific
compatible 0.x versions. For instance if we knew `foo` 0.2 and 0.3 were both compatible with `bar` 0.3 and 0.4, we
could use replace directives in our go.mod to stick to the versions of `foo` and `bar` we want. This would work as
long as the authors of `foo` and `bar` avoid incompatible breaking changes between these modules.

### Approach D) Use semantic versioning and don't address these issues



## Decision


## Consequences

### Backwards Compatibility

TODO

### Positive

TODO

### Negative

TODO

### Neutral

## Further Discussions

TODO

## References

* https://github.com/cosmos/cosmos-sdk/discussions/10162
* https://github.com/cosmos/cosmos-sdk/discussions/10582
* https://github.com/cosmos/cosmos-sdk/discussions/10368
* https://github.com/cosmos/cosmos-sdk/pull/11340