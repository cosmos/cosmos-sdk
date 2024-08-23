# RFC 002: Zero-Copy Encoding

## Changelog

* 2022-03-08: Initial draft

## Background

When the SDK originally migrated to [protobuf state encoding](./../architecture/adr-019-protobuf-state-encoding.md),
zero-copy encodings such as [Cap'n Proto](https://capnproto.org/)
and [FlatBuffers](https://google.github.io/flatbuffers/)
were considered. We considered how a zero-copy encoding could be beneficial for interoperability with modules
and scripts in other languages and VMs. However, protobuf was still chosen because the maturity of its ecosystem and
tooling was much higher and the client experience and performance were considered the highest priorities.

In [ADR 033: Protobuf-based Inter-Module Communication](./../architecture/adr-033-protobuf-inter-module-comm.md), the
idea of cross-language/VM inter-module
communication was considered again. And in the discussions surrounding [ADR 054: Semver Compatible SDK Modules](./../architecture/adr-054-semver-compatible-modules.md),
it was determined that multi-language/VM support in the SDK is a near term priority.

While we could do cross-language/VM inter-module communication with protobuf binary or even JSON, the performance
overhead is deemed to be too high because:
* we are proposing replacing keeper calls with inter-module message calls and the overhead of even the inter-module
routing checks has come into question by some SDK users without even considering the possible overhead of encoding.
Effectively we would be replacing function calls with encoding. One of the SDK's primary objectives currently is
improving performance, and we want to avoid inter-module calls from becoming a big step backward.
* we want Rust code to be able to operate in highly resource constrained virtual machines so whatever we can do to
reduce performance overhead as well as the size of generated code will make it easier and more feasible to deploy
first-class integrations with these virtual machines.

Thus, the agreement when the [ADR 054](./../architecture/adr-054-semver-compatible-modules.md) working group concluded
was to pursue a performant zero-copy encoding which is suitable for usage in highly resource constrained environments.

## Proposal

This RFC proposes a zero-copy encoding that is derived from the schema definitions defined in .proto files in the SDK
and all app chains. This would result in a new code generator for that supports both this zero-copy encoding as well as
the existing protobuf binary and JSON encodings as well as the google.golang.org/protobuf API. To make this zero-copy
encoding work, a number of changes are needed to how we manage the versioning of protobuf messages that should
address other concerns raised in [ADR 054](./../architecture/adr-054-semver-compatible-modules.md). The API for using
protobuf in golang would also change and this will be described in the [code generation](#generated-code) section
along with a proposed Rust code generator.

An alternative approach to building a zero-copy encoding based on protobuf schemas would be to switch to FlatBuffers
or Cap'n Proto directly. However, this would require a complete rewrite of the SDK and all app chains. Places this
burden on the ecosystem would not be a wise choice when creating a zero-copy encoding compatible with all our
existing types and schemas is feasible. In the future, we may consider a native schema language for this encoding
that is more natural and succinct for its rules, but for now we are assuming that it is best to continue supporting
the existing protobuf based workflow.

Also, we are not proposing a new encoding for transactions or gRPC query servers. From a client API perspective nothing
would change. The SDK would be capable of marshaling any message to and from protobuf binary and this zero-copy encoding
as needed.

Furthermore, migrating to the **new golang generated code would be 100% opt-in** because the inter-module router will
simply marshal existing gogo proto generated types to/from the zero-copy encoding when needed. So migrating to the new
code generator would provide a performance benefit, but would not be required.

In addition to supporting first-class Cosmos SDK modules defined in other languages and VMs, this encoding is intended
to be useful for user-defined code executing in a VM. To satisfy this, this encoding is designed to enable proper bounds
checking on all memory access at the expense of introducing some error return values in generated code.

### New Protobuf Linting and Breaking Change Rules

This zero-copy encoding places some additional requirements on the definition and maintenance of protobuf schemas.

#### No New Fields Can Be Added To Existing Messages

The biggest change is that it will be invalid to add a new field to an existing message and a breaking change detector
will need to be created which augments [buf breaking](https://docs.buf.build/breaking/overview) to detect this.

The reasons for this are two-fold:

1) from an API compatibility perspective, adding a new field to an existing message is actually a state machine breaking
   change which in [ADR 020](../architecture/adr-020-protobuf-transaction-encoding.md) required us to add an unknown
   field detector. Furthermore, in [ADR 054](../architecture/adr-054-semver-compatible-modules.md) this "feature" of protobuf
   poses one of the biggest problems for correct forward compatibility between different versions of the same module.
2) not allowing new fields in existing messages makes the generated code in languages like Rust (which is currently our
   highest priority target), much simpler and more performant because we can assume a fixed size struct gets allocated.
   If new fields can be added to existing messages, we need to encode the number of fields into the message and then
   do runtime checks. So this both increases memory layers and requires another layout of indirection. With the encoding
   proposed below, "plain old Rust structs" (used with some special field types) can be used.

Instead of adding new fields to existing messages, APIs can add new messages to existing packages or create new packages
with new versions of the messages. Also, we are not restricting the addition of cases to `oneof`s or values to `enum`s.
All of these cases are easier to detect at runtime with standard `switch` statements than the addition of new fields.

#### Additional Linting Rules

The following additional rules will be enforced by a linter that
complements [buf lint](https://docs.buf.build/lint/overview):

* all message fields must be specified in continuous ascending order starting from `1`
* all enums must be specified in continuous ascending order starting from `0` - otherwise it is too complex to check at
  runtime whether an enum value is unknown. An alternative would be to make adding new values to existing enums breaking
* all enum values must be `<= 255`. Any enum in a blockchain application which needs more than 256 values is probably
  doing something very wrong.
* all oneof's must be the *only* element in their containing message and must start at field number `1` and be added in
  continuous ascending order - this makes it possible to quickly check for unknown values
* all `oneof` field numbers must be `<= 255`. Any `oneof` which needs more field cases is probably doing something very
  wrong.

These requirements make the encoding and generated code simpler.

### Encoding

#### Buffers and Memory Management

By default, this encoding attempts to use a single fixed size encoding buffer of 64kb. This imposes a limit on the
maximum size of a message that can be encoded. In the context of a message passing protocol for blockchains, this
is generally a reasonable limit and the only known valid use case for exceeding it is to store user-uploaded byte
code for execution in VMs. To accommodate this, large `string` and `bytes` values can be encoded in additional
standalone buffers if needed. Still, the body of a message included all scalar and message fields
must fit inside the 64kb buffer.

While this design decision greatly simplifies the encoding and decoding logic, as well as the complexity of
generated code, it does mean that APIs will need to do proper bounds checking when writing data that is not fixed
size and return errors.

#### Alignment

In order to allow for zero-copy casting to native structs, all fields are aligned to their natural alignment.
This means that `int32` fields are aligned to 4 bytes, `int64` fields are aligned to 8 bytes, etc.

#### Scalar Encoding

* `bool`s are encoded as 1 byte - `0` or `1`
* `uint32`, `int32`, `sint32`, `fixed32`, `sfixed32` are encoded as 4 bytes by default
* `uint64`, `int64`, `sint64`, `fixed64`, `sfixed64` are encoded as 8 bytes by default
* `enum`s are encoded `int32`s
* all scalars declared as `optional` are prefixed with 1 additional byte whose value is `0` or `1` to indicate presence

All multibyte integers are encoded as little-endian which is by far the most common native byte order for modern
CPUs. Signed integers always use two's complement encoding.

#### Message Encoding

By default, messages field are encoded inline as structs. Meaning that if a message struct takes 8 bytes then its inline
field in another struct will add 8 bytes to that struct size. Structs must be aligned to their most aligned field.

`optional` message fields are encoded as pointers (see below) with a `0` pointer indicating that the field is not set.

#### Oneof’s

`oneof`s are encoded as a combination of a `int32` discriminant field and either the field value if it can fit in
8 bytes or a pointer to the field value.
A discriminant of `0` indicates that the field is not set.

#### Pointer Types: Bytes and Strings and Repeated fields

A pointer type is represented as 64-bits representing the offset from the current location with the limit of using
at most 31-bits of the offset.
This allows for the remaining bits to be used by implementations for other purposes.
If an implementation uses 64kb buffer aligned to 64kb, then the pointer can be left as is and bounds checking can
be performed on demand.
The root of a 64kb buffer aligned 64kb can be found by masking the current offset with
`0xFFFF0000` and then determining the end by adding the length to the offset.
If implementations don't want to have this restraint of buffers as it requires special or cumbersome wrapper types,
then they can implement a pointer fix-up step where the 64bit pointer is rewritten to contain both this offset
and the distant from the offset to the actual end of the buffer as a second 31-bit value.
Implementations can take this a step further and use pointer tagging,
whereby the remaining 2-bits in the pointer can be used as a pointer tag.
This tag can be used to indicate that the pointer is either a buffer offset or a native pointer to some memory
allocated elsewhere. This can be more convenient when we are building a message using struct initialization.
In Rust, we can even distinguish between borrowed and owned pointers using this tag like the native `Cow` type.
In most cases, it is probably more ergonomic to have a quick fix-up step which will lazily fix pointers in each piece
of data referenced in a message.
For example, say we have a pointer to a list of messages.
We can first fix up the pointer to the list.
Then, when we actually deference each element in the list, we can quickly fix any pointers it contains
using the known buffer offset or return an error.

#### `Any`s

`Any`s are encoded as a pointer to the type URL string and a pointer to the start of the message
specified by the type URL.

#### Maps

Maps are not supported.

### Generated Code

We will describe the generated Go and Rust code using this example protobuf file:

```protobuf
message Foo {
  int32 x = 1;
  optional uint32 y = 2;
  string z = 3;
  Bar bar = 4;
  repeated Bar bars = 5;
}


message Bar {
  ABC abc = 1;
  Baz baz = 2;
  repeated uint32 xs = 3;
}

message Baz {
  oneof sum {
    uint32 x = 1;
    string y = 2;
  }
}

enum ABC {
  A = 0;
  B = 1;
  C = 2;
  D = 3;
}
```

#### Go

In Golang, we will simply unmarshal data to 

#### Rust

This encoding should allow generating native structs in Rust that are annotated with `#[repr(C)]`.
Lifetime parameters are added to allow for borrowed strings and slices to be used in the structs.

Here is some example code that uses library types `Option`, `Enum`, `String`, `OneOf` and `Repeated`.

```rust!
#[repr(C)]
struct Foo<'a> {
    x: i32,
    y: u21,
    z: cosmos_proto::String<'a, // String wraps an offset pointer or a native owned or borrowed string
    bar: Bar
}

#[repr(C)]
struct Bar {
    abc: cosmos_proto::Enum<ABC>, // the Enum wrapper allows us to distinguish undefined and defined values of ABC at runtime. 3 is specified as the max value of ABC.
    baz: cosmos_proto::OneOf<Baz<'a>>, // the OneOf wrapper allows distinguished undefined values of Baz at runtime
    xs: cosmos_proto::Repeated<'a, u32> // Repeated wraps a pointer to repeated fields - the data can point to a buffer offset, native owned vector or borrowed slice
}

#[repr(u8)]
enum ABC {
    A = 0,
    B = 1,
    C = 2,
    D = 3,
}

#[repr(C, u8)]
enum Baz<'a> {
    Empty, // all oneof's have a case for Empty if they are unset
    X(u32),
    Y(cosmos_proto::String<'a>)
}
```

Example usage (which does the exact same thing as the go example above) would be:

```rust!
let mut foo = Foo{
  x: 1,
  y: 2,
  z: "hello".into(),
  bar: Bar{
    abc: ABC::A,
    baz: Baz::Empty,
    xs: vec![0, 2].into()
  }
}
```

## Abandoned Ideas (Optional)

## References

## Discussion
