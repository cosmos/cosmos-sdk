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
communication was considered again. And in the discussions
around [ADR 054: Semver Compatible SDK Modules](./../architecture/adr-054-semver-compatible-modules.md),
it was determined that multi-language/VM support in the SDK is a near term priority.

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

### New Protobuf Linting and Breaking Change Rules

This zero-copy encoding places some additional requirements on the definition and maintenance of protobuf schemas.

#### No New Fields Can Be Added To Existing Messages

The biggest change is that it will be invalid to add a new field to an existing message and a breaking change detector
will need to be created which augments [buf breaking](https://docs.buf.build/breaking/overview) to detect this.

The reasons for this are two-fold:

1) from an API compatibility perspective, adding a new field to an existing message is actually a state machine breaking
   change which in [ADR 020](../architecture/adr-020-protobuf-transaction-encoding.md) required us to add an unknown
   field
   detector. Furthermore, in [ADR 054](../architecture/adr-054-semver-compatible-modules.md) this "feature" of protobuf
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
  continuous ascending order - this makes it possible to quickly check for unknown values, but also Rust enums don't
  allow specifying discriminant values manually
* all `oneof` field numbers must be `<= 255`. Any `oneof` which needs more field cases is probably doing something very
  wrong.

These requirements make the encoding and generated code simpler.

### Encoding

The encoding of protobuf scalar and composite types is described below.

#### Scalars

* `bool`s are encoded as 1 byte - `0` or `1`
* `uint32`, `int32`, `sint32`, `fixed32`, `sfixed32` are encoded as 4 bytes by default
* `uint64`, `int64`, `sint64`, `fixed64`, `sfixed64` are encoded as 8 bytes by default
* `enum`s are encoded as 1 byte and values *MUST* be in the range of `0` to `255`.
* all scalars declared as `optional` are prefixed with 1 additional whose value is `0` or `1` to indicate presence

#### Messages

By default, messages field are encoded inline as structs. Meaning that if a message struct takes 8 bytes then its inline
field in another struct will add 8 bytes to that struct size.

`optional` message fields will be prefixed by 1 byte to indicate presence. (Alternatively, we could encode optional
message fields as relative pointers (see below) if the desire is to save memory when they are rarely used needed.)

#### Oneof’s

`oneof`s are encoded as a combination of a `uint8` discriminant field and memory that is as large as the largest member
field. `oneof` field numbers *MUST* be between `1` and `255`.

```protobuf
message Foo {
  oneof sum {
    bool x = 1;
    int32 y = 2;
  }
}
```

A discriminant of `0` indicates that the field is not set.

#### Relative Pointer Types: Bytes and Strings and Repeated fields

A relative pointer is an integer offset from the start of the current memory location to another memory location.
Almost all messages should be smaller than 64kb so a 16-bit relative pointer should be sufficient. An amendment to this
specification may choose to describe how to deal with larger messages. For now, protocols that need to use a larger
amount of language should be implemented as native Cosmos SDK go modules. The only known use case for messages larger
than 64kb in a blockchain application is storing executable byte code for a VM, so this seems like a reasonable
limitation.

`bytes`, `string`s and repeated fields are encoded as a relative pointer to a memory location that is prefixed with the
length of the `bytes`, `string` or repeated field value. The length should be encoded as the size of the relative
pointer - either 16 or 32-bits. The relative pointer `0` indicates that a field is not defined.

#### `Any`s

`Any`s are encoded as a relative pointer to the type URL string and a relative pointer to the start of the message
specified by the type URL.

#### Maps

Maps are not supported.

#### Extended Encoding Options

We may choose to allow customizing the encoding of fields so that they take up less space.

For example, we could allow 8-bit or 16-bit integers:
`int32 x = 1 [(cosmos_proto.int16) = true]` would indicate that the field only needs 2 bytes

Or we could allow `string` or `bytes` fields to have a fixed size rather than being encoding as
variable-length relative pointers:
`string y = 2 [(cosmos_proto.fixed_size) = 3]` could indicate that this is a fixed width 3 byte string

If we choose to enable these encoding options, changing these options would be a breaking change that needs to be
prevented by the breaking change detector.

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

In golang, the generated code would not expose any exported struct fields, but rather getters and setters as an
interface
or struct methods, ex:

```go
type Foo interface {
X() int32
SetX(int32) Foo
Y() zpb.Option[uint32]
SetY(zpb.Option[uint32]) Foo
Z() string
SetZ(string) Foo
Bar() Bar
Bars() zpb.Array[Bar]
}

type Bar interface {
Abc() ABC
SetAbc(ABC) Bar
Baz() Baz
Xs() zpb.ScalarArray[uint32]
}

type Baz interface {
Case() Baz_case
GetX() uint32
SetX(uint32)
GetY() string
SetY(string)
}

type Baz_case int32
const (
Baz_X Baz_case = 0
Baz_Y Baz_case = 1
)

type ABC int32
const (
ABC_A ABC = 0
ABC_B ABC = 1
ABC_C ABC = 2
ABC_D ABC = 3
)
```

Special types `zpb.Option`, `zpb.Array` and `zpb.ScalarArray` are used to represent `optional` and repeated fields
respectively. These types would be included in the runtime library (called `zpb` here for zero-copy protobuf) and would
have an API like this:

```go
type Option[T] interface {
IsSet() bool
Value() T
}

type Array[T] interface {
InitWithLength(int)
Len() int
Get(int) T
}

type ScalarArray[T] interface {
Array[T]
Set(int, T)
}
```

Arrays in particular would not be resizable, but would be initialized with a fixed length. This is to ensure that arrays
can be written to the underlying buffer in a linear way.

In golang, buffers would be managed transparently under the hood, and usage of this generated code might look like this
with method chaining used on setters:

```go
foo := NewFoo()
foo.SetX(1)
.SetY(zpb.NewOption[uint32](2))
.SetZ("hello")

bar := foo.Bar()
bar.Baz().SetX(3)

xs := bar.Xs()
xs.InitWithLength(2)
xs.Set(0, 0)
xs.Set(1, 2)

bars := foo.Bars()
bars.InitWithLength(3)
bars.Get(0).Baz().SetY("hello")
bars.Get(1).SetAbc(ABC_B)
bars.Get(2).Baz().SetX(4)
```

Under the hood the generated code would manage memory buffers on its own. The usage of `oneof`s is a bit easier than
the existing go generated code (as with `bar.Baz()` above). And rather than using setters on embedded messages, we
simply get the field (already allocated) and set its fields (as in the case of `foo.Bar()` above or the repeated
field `foo.Bars()`).

#### Rust

This encoding should allow generating native structs in Rust that are annotated with `#[repr(C, align(1))]`. It should
be fairly natural to use from Rust with a key difference that memory buffers (called `Root`s) must be manually allocated
and passed into any relative pointer type.

Here is some example code that uses library types `Option`, `Enum`, `String`, `OneOf` and `Repeated`:

```rust!
#[repr(C, align(1))]
struct Foo {
    x: i32,
    y: cosmos_proto::Option<u32>,
    z: cosmos_proto::String, // String wraps a relative pointer to a string
    bar: Bar
}

#[repr(C, align(1))]
struct Bar {
    abc: cosmos_proto::Enum<ABC, 3>, // the Enum wrapper allows us to distinguish undefined and defined values of ABC at runtime. 3 is specified as the max value of ABC.
    baz: cosmos_proto::OneOf<Baz, 2>, // the OneOf wrapper allows distinguished undefined values of Baz at runtime. 2 is specified as the max field value of Baz.
    xs: cosmos_proto::Repeated<u32> // Repeated wraps a relative pointer to repeated fields
}

#[repr(u8)]
enum ABC {
    A = 0,
    B = 1,
    C = 2,
    D = 3,
}

#[repr(C, u8)]
enum Baz {
    Empty, // all oneof's have a case for Empty if they are unset
    X(u32),
    Y(cosmos_proto::String)
}
```

Example usage (which does the exact same thing as the go example above) would be:
```rust
let mut root = Root<Foo>::new();
let mut foo = root.get_mut();
foo.x = 1;
foo.y = Some(2);
foo.z.set(root, "hello")?; // can return an error if the buffer is too small

foo.bar.baz = Baz::X(3);

foo.bar.xs.init_with_size(root, 2)?; // can return an error if the buffer is too small
foo.bar.xs[0] = 0;
foo.bar.xs[1] = 2;

foo.bars.init_with_size(root, 3)?; // can return an error if the buffer is too small
foo.bars[0].baz = Baz::Y(cosmos_proto::String::new(root, "hello")?); // can return an error if the buffer is too small
foo.bars[1].abc = ABC::B;
foo.bars[2].baz = Baz::X(4);
```

#### Memory Buffer Management

One potential issue to this approach is that it will be impossible to grow the memory buffer after its initial allocation
if we are writing structs like this. To deal with this, a fixed-size 64kb buffer should be allocated as the default which
should be sufficient for all normal blockchain use cases (besides storing VM byte code). Allocation methods will return
an error if the memory buffer size is exceeded.

In the future, we may consider ways to use different sized memory buffers, maybe specified at the message level with
annotations, but this is not a priority for now because applications that need larger memory buffers be implemented as
native Cosmos SDK go modules. The main use case for this would be to support user-defined code to run in VMs and this
is deemed an acceptable trade-off for now.

NOTE: in golang, it is feasible to grow the memory buffer at runtime if needed because of how the golang code is
generated to use getters and setters rather than struct fields. One key motivation for the Rust design above is to
minimize the amount of generated code because we expect that it is more likely that Rust code will run in resource
constrained environments. Thus, for Rust generated code there are slightly higher performance requirements and
expectations than for Go code.

## Abandoned Ideas (Optional)

## References

## Discussion
