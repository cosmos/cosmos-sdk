# ADR 054: Semver Compatible SDK Modules

## Changelog

* 2022-04-27: First draft
* 2024-07-21: Second draft

## Status

DRAFT

## Abstract

In order to move the Cosmos SDK to a system of decoupled semantically versioned
modules which can be composed in different combinations (ex. staking v3 with
bank v1 and distribution v2), we need to reassess how we organize the API surface
of modules to avoid problems with go semantic import versioning and
circular dependencies. This ADR explores various approaches we can take to
addressing these issues.

## Context

There has been [a fair amount of desire](https://github.com/cosmos/cosmos-sdk/discussions/10162)
in the community for semantic versioning in the SDK and there has been significant
movement to splitting SDK modules into [standalone go modules](https://github.com/cosmos/cosmos-sdk/issues/11899).
Both of these will ideally allow the ecosystem to move faster because we won't
be waiting for all dependencies to update synchronously. For instance, we could
have 3 versions of the core SDK compatible with the latest 2 releases of
CosmWasm as well as 4 different versions of staking . This sort of setup would
allow early adopters to aggressively integrate new versions, while allowing
more conservative users to be selective about which versions they're ready for.

In order to achieve this, we need to solve the following problems:

1. because of the way [go semantic import versioning](https://research.swtch.com/vgo-import) (SIV)
   works, moving to SIV naively will actually make it harder to achieve these goals
2. circular dependencies between modules need to be broken to actually release
   many modules in the SDK independently
3. pernicious minor version incompatibilities introduced through correctly
   [evolving protobuf schemas](https://protobuf.dev/programming-guides/proto3/#updating)
   without correct [unknown field filtering](./adr-020-protobuf-transaction-encoding.md#unknown-field-filtering)

Note that all the following discussion assumes that the proto file versioning and state machine versioning of a module
are distinct in that:

* proto files are maintained in a non-breaking way (using something
  like [buf breaking](https://docs.buf.build/breaking/overview)
  to ensure all changes are backwards compatible)
* proto file versions get bumped much less frequently, i.e. we might maintain `cosmos.bank.v1` through many versions
  of the bank module state machine
* state machine breaking changes are more common and ideally this is what we'd want to semantically version with
  go modules, ex. `x/bank/v2`, `x/bank/v3`, etc.

### Problem 1: Semantic Import Versioning Compatibility

Consider we have a module `foo` which defines the following `MsgDoSomething` and that we've released its state
machine in go module `example.com/foo`:

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
add a new validation rule on `amount` requiring it to be non-zero, and that following go semantic versioning we
release the next state machine version of `foo` as `example.com/foo/v2`.

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
incompatible structs in the go type system.

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
even if both versions of these modules are otherwise compatible with each
other.

#### Naive Mitigation

A naive approach to fixing this would be to not regenerate the protobuf types
in `example.com/foo/v2/types` but instead just update `example.com/foo/types`
to reflect the changes needed for `v2` (adding `condition` and requiring
`amount` to be non-zero). Then we could release a patch of `example.com/foo/types`
with this update and use that for `foo/v2`. But this change is state machine
breaking for `v1`. It requires changing the `ValidateBasic` method to reject
the case where `amount` is zero, and it adds the `condition` field which
should be rejected based
on [ADR 020 unknown field filtering](./adr-020-protobuf-transaction-encoding.md#unknown-field-filtering).
So adding these changes as a patch on `v1` is actually incorrect based on semantic
versioning. Chains that want to stay on `v1` of `foo` should not
be importing these changes because they are incorrect for `v1.`

### Problem 2: Circular dependencies

None of the above approaches allow `foo` and `bar` to be separate modules
if for some reason `foo` and `bar` depend on each other in different ways.
For instance, we can't have `foo` import `bar/types` while `bar` imports
`foo/types`.

We have several cases of circular module dependencies in the SDK
(ex. staking, distribution and slashing) that are legitimate from a state machine
perspective. Without separating the API types out somehow, there would be
no way to independently semantically version these modules without some other
mitigation.

### Problem 3: Handling Minor Version Incompatibilities

Imagine that we solve the first two problems but now have a scenario where
`bar/v2` wants the option to use `MsgDoSomething.condition` which only `foo/v2`
supports. If `bar/v2` works with `foo` `v1` and sets `condition` to some non-nil
value, then `foo` will silently ignore this field resulting in a silent logic
possibly dangerous logic error. If `bar/v2` were able to check whether `foo` was
on `v1` or `v2` and dynamically, it could choose to only use `condition` when
`foo/v2` is available. Even if `bar/v2` were able to perform this check, however,
how do we know that it is always performing the check properly. Without
some sort of
framework-level [unknown field filtering](./adr-020-protobuf-transaction-encoding.md#unknown-field-filtering),
it is hard to know whether these pernicious hard to detect bugs are getting into
our app and a client-server layer such as [ADR 033: Inter-Module Communication](./adr-033-protobuf-inter-module-comm.md)
may be needed to do this.

## Solutions

### Approach A) Separate API and State Machine Modules

One solution (first proposed in https://github.com/cosmos/cosmos-sdk/discussions/10582) is to isolate all protobuf
generated code into a separate module
from the state machine module. This would mean that we could have state machine
go modules `foo` and `foo/v2` which could use a types or API go module say
`foo/api`. This `foo/api` go module would be perpetually on `v1.x` and only
accept non-breaking changes. This would then allow other modules to be
compatible with either `foo` or `foo/v2` as long as the inter-module API only
depends on the types in `foo/api`. It would also allow modules `foo` and `bar`
to depend on each other in that both of them could depend on `foo/api` and
`bar/api` without `foo` directly depending on `bar` and vice versa.

This is similar to the naive mitigation described above except that it separates
the types into separate go modules which in and of itself could be used to
break circular module dependencies. It has the same problems as the naive solution,
otherwise, which we could rectify by:

1. removing all state machine breaking code from the API module (ex. `ValidateBasic` and any other interface methods)
2. embedding the correct file descriptors for unknown field filtering in the binary

#### Migrate all interface methods on API types to handlers

To solve 1), we need to remove all interface implementations from generated
types and instead use a handler approach which essentially means that given
a type `X`, we have some sort of resolver which allows us to resolve interface
implementations for that type (ex. `sdk.Msg` or `authz.Authorization`). For
example:

```go
func (k Keeper) DoSomething(msg MsgDoSomething) error {
	var validateBasicHandler ValidateBasicHandler
	err := k.resolver.Resolve(&validateBasicHandler, msg)
	if err != nil {
		return err
	}   
	
	err = validateBasicHandler.ValidateBasic()
	...
}
```

In the case of some methods on `sdk.Msg`, we could replace them with declarative
annotations. For instance, `GetSigners` can already be replaced by the protobuf
annotation `cosmos.msg.v1.signer`. In the future, we may consider some sort
of protobuf validation framework (like https://github.com/bufbuild/protoc-gen-validate
but more Cosmos-specific) to replace `ValidateBasic`.

#### Pinned FileDescriptor's

To solve 2), state machine modules must be able to specify what the version of
the protobuf files was that they were built against. For instance if the API
module for `foo` upgrades to `foo/v2`, the original `foo` module still needs
a copy of the original protobuf files it was built with so that ADR 020
unknown field filtering will reject `MsgDoSomething` when `condition` is
set.

The simplest way to do this may be to embed the protobuf `FileDescriptor`s into
the module itself so that these `FileDescriptor`s are used at runtime rather
than the ones that are built into the `foo/api` which may be different. Using
[buf build](https://docs.buf.build/build/usage#output-format), [go embed](https://pkg.go.dev/embed),
and a build script we can probably come up with a solution for embedding
`FileDescriptor`s into modules that is fairly straightforward.

#### Potential limitations to generated code

One challenge with this approach is that it places heavy restrictions on what
can go in API modules and requires that most of this is state machine breaking.
All or most of the code in the API module would be generated from protobuf
files, so we can probably control this with how code generation is done, but
it is a risk to be aware of.

For instance, we do code generation for the ORM that in the future could
contain optimizations that are state machine breaking. We
would either need to ensure very carefully that the optimizations aren't
actually state machine breaking in generated code or separate this generated code
out from the API module into the state machine module. Both of these mitigations
are potentially viable but the API module approach does require an extra level
of care to avoid these sorts of issues.

#### Minor Version Incompatibilities

This approach in and of itself does little to address any potential minor
version incompatibilities and the
requisite [unknown field filtering](./adr-020-protobuf-transaction-encoding.md#unknown-field-filtering).
Likely some sort of client-server routing layer which does this check such as
[ADR 033: Inter-Module communication](./adr-033-protobuf-inter-module-comm.md)
is required to make sure that this is done properly. We could then allow
modules to perform a runtime check given a `MsgClient`, ex:

```go
func (k Keeper) CallFoo() error {
	if k.interModuleClient.MinorRevision(k.fooMsgClient) >= 2 {
		k.fooMsgClient.DoSomething(&MsgDoSomething{Condition: ...})
    } else {
        ...
    }
}
```

To do the unknown field filtering itself, the ADR 033 router would need to use
the [protoreflect API](https://pkg.go.dev/google.golang.org/protobuf/reflect/protoreflect)
to ensure that no fields unknown to the receiving module are set. This could
result in an undesirable performance hit depending on how complex this logic is.

#### No New Fields in Existing Protobuf Messages

An alternative to addressing minor version incompatibilities as described above is disallowing new fields in existing protobuf messages. While this is more restrictive, it simplifies versioning and eliminates the need for runtime unknown field checking. In addition, this approach would simplify cross language communication with the proposed [RFC 002: Zero Copy Encoding](../rfc/rfc-002-zero-copy-encoding.md). So, while it is rather restrictive, it has gained a fair amount of support.

Although disallowing new fields may seem overly restrictive, there is a straightforward way to work around it using protobuf `oneof`s. Because `oneof` and `enum` cases must get processed through a `switch` statement, adding new cases is not problematic because any unknown cases can be handled by a `default` clause. The router layer wouldn't need to do unknown field filtering for these because the `switch` statement is a native way to do this. If we needed to add new fields to `MsgDoSomething` from above and retain the possibility of adding more new fields in the future, we could do something like this:

```protobuf
message MsgDoSomethingWithOptions {
  string sender = 1;
  uint64 amount = 2;
  repeated MsgDoSomethingOption options = 3;
}

message MsgDoSomethingOption {
  oneof option {
    Condition condition = 1;
  }
}
```

New `oneof` cases can be added to `MsgDoSomethingOption` and this has a similar effect as adding new fields to `MsgDoSomethingWithOptions` but no new fields are needed. A similar strategy is recommended for adding variadic options to golang functions in https://go.dev/blog/module-compatibility and expanded upon in https://commandcenter.blogspot.com/2014/01/self-referential-functions-and-design.html.

### Approach B) Changes to Generated Code to a Getter/Setter API

An alternate approach to solving the versioning problem is to change how protobuf code is generated and move modules
mostly or completely in the direction of inter-module communication as described
in [ADR 033](./adr-033-protobuf-inter-module-comm.md).
In this paradigm, a module could generate all the types it needs internally - including the API types of other modules -
and talk to other modules via a client-server boundary. For instance, if `bar` needs to talk to `foo`, it could
generate its own version of `MsgDoSomething` as `bar/internal/foo/v1.MsgDoSomething` and just pass this to the
inter-module router which would somehow convert it to the version which foo needs (ex. `foo/internal.MsgDoSomething`).

Currently, two generated structs for the same protobuf type cannot exist in the same go binary without special
build flags (see https://protobuf.dev/reference/go/faq/#fix-namespace-conflict).
A relatively simple mitigation to this issue would be to set up the protobuf code to not register protobuf types
globally if they are generated in an `internal/` package. This will require modules to register their types manually
with the app-level level protobuf registry, this is similar to what modules already do with the `InterfaceRegistry`
and amino codec.

If modules _only_ do ADR 033 message passing then a naive and non-performant solution for
converting `bar/internal/foo/v1.MsgDoSomething`
to `foo/internal.MsgDoSomething` would be marshaling and unmarshaling in the ADR 033 router. This would break down if
we needed to expose protobuf types in `Keeper` interfaces because the whole point is to try to keep these types
`internal/` so that we don't end up with all the import version incompatibilities we've described above. However,
because of the issue with minor version incompatibilities and the need
for [unknown field filtering](./adr-020-protobuf-transaction-encoding.md#unknown-field-filtering),
sticking with the `Keeper` paradigm instead of ADR 033 may be unviable to begin with.

A more performant solution (that could maybe be adapted to work with `Keeper` interfaces) would be to only expose
getters and setters for generated types and internally store data in memory buffers which could be passed from
one implementation to another in a zero-copy way.

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

func NewMsgSend() MsgSend { return &msgSendImpl{memoryBuffers: ...} }
```

Under the hood, `MsgSend` could be implemented based on some raw memory buffer in the same way
that [Cap'n Proto](https://capnproto.org)
and [FlatBuffers](https://google.github.io/flatbuffers/) so that we could convert between one version of `MsgSend`
and another without serialization (i.e. zero-copy). This approach would have the added benefits of allowing zero-copy
message passing to modules written in other languages such as Rust and accessed through a VM or FFI. It could also make
unknown field filtering in inter-module communication simpler if we require that all new fields are added in sequential
order, ex. just checking that no field `> 5` is set.

Also, we wouldn't have any issues with state machine breaking code on generated types because all the generated
code used in the state machine would actually live in the state machine module itself. Depending on how interface
types and protobuf `Any`s are used in other languages, however, it may still be desirable to take the handler
approach described in approach A. Either way, types implementing interfaces would still need to be registered
with an `InterfaceRegistry` as they are now because there would be no way to retrieve them via the global registry.

In order to simplify access to other modules using ADR 033, a public API module (maybe even one
[remotely generated by Buf](https://docs.buf.build/bsr/remote-generation/go)) could be used by client modules instead
of requiring to generate all client types internally.

The big downsides of this approach are that it requires big changes to how people use protobuf types and would be a
substantial rewrite of the protobuf code generator. This new generated code, however, could still be made compatible
with
the [`google.golang.org/protobuf/reflect/protoreflect`](https://pkg.go.dev/google.golang.org/protobuf/reflect/protoreflect)
API in order to work with all standard golang protobuf tooling.

It is possible that the naive approach of marshaling/unmarshaling in the ADR 033 router is an acceptable intermediate
solution if the changes to the code generator are seen as too complex. However, since all modules would likely need
to migrate to ADR 033 anyway with this approach, it might be better to do this all at once.

### Approach C) Don't address these issues

If the above solutions are seen as too complex, we can also decide not to do anything explicit to enable better module
version compatibility, and break circular dependencies.

In this case, when developers are confronted with the issues described above they can require dependencies to update in
sync (what we do now) or attempt some ad-hoc potentially hacky solution.

One approach is to ditch go semantic import versioning (SIV) altogether. Some people have commented that go's SIV
(i.e. changing the import path to `foo/v2`, `foo/v3`, etc.) is too restrictive and that it should be optional. The
golang maintainers disagree and only officially support semantic import versioning. We could, however, take the
contrarian perspective and get more flexibility by using 0.x-based versioning basically forever.

Module version compatibility could then be achieved using go.mod replace directives to pin dependencies to specific
compatible 0.x versions. For instance if we knew `foo` 0.2 and 0.3 were both compatible with `bar` 0.3 and 0.4, we
could use replace directives in our go.mod to stick to the versions of `foo` and `bar` we want. This would work as
long as the authors of `foo` and `bar` avoid incompatible breaking changes between these modules.

Or, if developers choose to use semantic import versioning, they can attempt the naive solution described above
and would also need to use special tags and replace directives to make sure that modules are pinned to the correct
versions.

Note, however, that all of these ad-hoc approaches, would be vulnerable to the minor version compatibility issues
described above unless [unknown field filtering](./adr-020-protobuf-transaction-encoding.md#unknown-field-filtering)
is properly addressed.

### Approach D) Avoid protobuf generated code in public APIs

An alternative approach would be to avoid protobuf generated code in public module APIs. This would help avoid the
discrepancy between state machine versions and client API versions at the module to module boundaries. It would mean
that we wouldn't do inter-module message passing based on ADR 033, but rather stick to the existing keeper approach
and take it one step further by avoiding any protobuf generated code in the keeper interface methods.

Using this approach, our `foo.Keeper.DoSomething` method wouldn't have the generated `MsgDoSomething` struct (which
comes from the protobuf API), but instead positional parameters. Then in order for `foo/v2` to support the `foo/v1`
keeper it would simply need to implement both the v1 and v2 keeper APIs. The `DoSomething` method in v2 could have the
additional `condition` parameter, but this wouldn't be present in v1 at all so there would be no danger of a client
accidentally setting this when it isn't available. 

So this approach would avoid the challenge around minor version incompatibilities because the existing module keeper
API would not get new fields when they are added to protobuf files.

Taking this approach, however, would likely require making all protobuf generated code internal in order to prevent
it from leaking into the keeper API. This means we would still need to modify the protobuf code generator to not
register `internal/` code with the global registry, and we would still need to manually register protobuf
`FileDescriptor`s (this is probably true in all scenarios). It may, however, be possible to avoid needing to refactor
interface methods on generated types to handlers.

Also, this approach doesn't address what would be done in scenarios where modules still want to use the message router.
Either way, we probably still want a way to pass messages from one module to another router safely even if it's just for
use cases like `x/gov`, `x/authz`, CosmWasm, etc. That would still require most of the things outlined in approach (B),
although we could advise modules to prefer keepers for communicating with other modules.

The biggest downside of this approach is probably that it requires a strict refactoring of keeper interfaces to avoid
generated code leaking into the API. This may result in cases where we need to duplicate types that are already defined
in proto files and then write methods for converting between the golang and protobuf version. This may end up in a lot
of unnecessary boilerplate and that may discourage modules from actually adopting it and achieving effective version
compatibility. Approaches (A) and (B), although heavy handed initially, aim to provide a system which once adopted
more or less gives the developer version compatibility for free with minimal boilerplate. Approach (D) may not be able
to provide such a straightforward system since it requires a golang API to be defined alongside a protobuf API in a
way that requires duplication and differing sets of design principles (protobuf APIs encourage additive changes
while golang APIs would forbid it).

Other downsides to this approach are:

* no clear roadmap to supporting modules in other languages like Rust
* doesn't get us any closer to proper object capability security (one of the goals of ADR 033)
* ADR 033 needs to be done properly anyway for the set of use cases which do need it

### Approach E) Use Structural Typing in Inter-module APIs, Avoid New Fields on Messages

The current non-router based approach for inter-module communication is for a module to define a `Keeper` interface and for a consumer module to define an expected keeper interface with a subset of the keeper's methods. Such an interface can allow one module to avoid a direct dependency on another module if no concrete types need to be imported from the other module. For instance, if we had a method `DoSomething(context.Context, string, uint64)` as in `foo.v1`, then a module calling `DoSomething` would not need to import `foo` directly. If, however, we had a struct parameter such as `Condition` and the new method were `DoSomethingV2(context.Context, string, uint64, foo.Condition)` then a calling module would generally need to import foo just to get a reference to the `foo.Condition` struct.

Golang, however, supports both structural and nominal typing. Nominal typing means that two types equivalent if and only if they have the same name. Structural typing in golang means that two types are equivalent if they have the same structure and are unnamed. So if we defined `Condition` nominally it might look like `type Condition struct { Field1 string }` and if we defined it structurally it would look like `type Condition = struct { Field1 string }`. If `Condition` were defined structurally we could use the expected keeper approach and a calling module _would not_ need to import `foo` at all to define the `DoSomethingV2` method in its expected keeper interface. Structural typing avoids the dependency problems described above.

We could actually extend this structural typing paradigm to protobuf generated code _if_ we disallow adding new fields to existing protobuf messages. This would be required because two struct types are only identical if their fields are identical. If even the order of fields in a struct or the struct tags change, then golang considers the structs as different types. While this is fairly restrictive, it is under consideration for approach A) and [RFC 002](../rfc/rfc-002-zero-copy-encoding.md) as well and has gained a fair amount of support.

Small modifications to the existing pulsar code generator could potentially generate code that uses structural typing and moves the implementation of protobuf interfaces to wrapper types because unnamed structs can't define methods. Here's an example of what this might look like:

```go
type MsgDoSomething = struct {
	Sender string
    Amount uint64
}

type MsgDoSomething_Message(*MsgDoSomething)

// MsgDoSomething_Message would actually implement protobuf methods
var _ proto.Message = (*MsgDoSomething_Message)(nil)
```

At least at the message layer, such an API wouldn't pose a problem because the transaction decoder does message decoding and modules wouldn't need to interact with the `proto.Message` interface directly.

In order to avoid problems with the global protobuf registry, the structural typing generated code would only register message descriptors with the global registry but not register message types. This would allow two modules to generate the same protobuf types in different packages without causing a conflict. Because the types are defined structurally, they would actually be the _same_ types but no direct import would be required. In order to ensure compatibility, when message descriptors are registered at startup, a check would be required to ensure that messages are identical (i.e. no new fields).

With this approach, a module `bar` calling module `foo` could either import `foo` directly to get its types or generate its own set of compatible types for `foo`s API. The dependency problem is essentially solved with this approach without needing any sort of special discipline around a separate API module. The main discipline would be around versioning protobuf APIs correctly and not adding new fields.

Taking this approach one step further, we could potentially even define APIs that unwrap the request and response types, i.e. `DoSomething(context.Context, string, uint64) error` vs `DoSomething(context.Context, MsgDoSomething) (MsgDoSomethingResponse, error)`. Or, APIs could even be defined directly in golang and the `.proto` files plus marshaling code could be generated via `go generate`. Ex:

```go
package foo 

import "context"

//go:generate go run github.com/cosmos/cosmos-proto/cmd/structproto

type Msg interface {
	DoSomething(context.Context, string, uint64) error
}
```

While having the limitation of not allowing new fields to be added to existing structs, approach E) has the following benefits:
* unlike approach A), api types can be generated in the same go module, but direct imports can always be avoided
* generated client/server code could look more like regular go interfaces (without needing a set of intermediate structs)
* keeper interface defined in `.go` files could be turned into protobuf APIs (rather than needing to write `.proto` files)
* SDK modules could adopt go semantic versioning without any of the issues described above, achieving the initially stated goals of this ADR

## Decision

There has been no decision yet, and the SDK has more or less been following approach C) and official adoption of [0ver](https://0ver.org) as a policy has been discussed. The issue of decoupling modules, properly versioning protobuf types, avoiding breakage, and adopting semver continue to arise from time to time. The most serious alternatives under consideration currently are approaches A) and E). The remainder of this ADR has been left blank and will be filled in when and if there is further convergence on a solution.

## Consequences

### Backwards Compatibility

### Positive

### Negative

### Neutral

## Further Discussions

## References

* https://github.com/cosmos/cosmos-sdk/discussions/10162
* https://github.com/cosmos/cosmos-sdk/discussions/10582
* https://github.com/cosmos/cosmos-sdk/discussions/10368
* https://github.com/cosmos/cosmos-sdk/pull/11340
* https://github.com/cosmos/cosmos-sdk/issues/11899
* [ADR 020](./adr-020-protobuf-transaction-encoding.md)
* [ADR 033](./adr-033-protobuf-inter-module-comm.md)
