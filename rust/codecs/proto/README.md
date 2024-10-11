## Protobuf Mapping

The following table shows how Rust types are mapped to Protobuf types. Note that some
types will have both a default and alternate mapping.

| Rust Type | Protobuf Type(s)            | Notes                                       |
|-----------|-----------------------------|---------------------------------------------|
| `u8`      | `uint32`                    |                                             |
| `u16`     | `uint32`                    |                                             |
| `u32`     | `uint32`                    |                                             |
| `u64`     | `uint64`                    |                                             |
| `u128`    | `string`                    |                                             |
| `i8`      | `int32`                     |                                             |
| `i16`     | `int32`                     |                                             |
| `i32`     | `int32`                     |                                             |
| `i64`     | `int64`                     |                                             |
| `i128`    | `string`                    |                                             |
| `bool`    | `bool`                      |                                             |
| `str`     | `string`                    |                                             |
| `String`  | `string`                    |                                             |
| `Address` | `string`, alternate `bytes` | uses an address codec for string conversion |
| `Time`    | `google.protobuf.Timestamp` |                                             |
| `Duration`| `google.protobuf.Duration`  |                                             |

## Protobuf Compatibility

Most existing [Cosmos SDK](https://github.com/cosmos/cosmos-sdk) applications use [Protocol Buffers](https://protobuf.dev) encoding.
So far, there has been no mention of it, and all of the examples here use plain old Rust
functions and structs so you may be wondering what's going on.

The [`ixc_schema`] crate defines a rich set of types that can be used in messages that aims to be
richer and more appropriate to this use case than the set of types provided by Protobuf.
Currently, many Cosmos SDK messages use Protobuf `string`s everywhere to represent other data
types.
Sometimes these are annotated with `cosmos_proto.scalar` to indicate which data type we actually mean.
In Rust, we encourage you to use the [`Address`] type for addresses and sized integer types, such as
`u128`, instead of strings or byte arrays.
For all of these types, there is a configurable mapping to Protobuf encoding that is described in more detail
in the [`ixc_schema`] crate documentation.
If the default mapping does not work, an alternate one may be available and can be annotated with optional
`#[proto]` and `#[schema]` attributes.
Ex:

```rust
#[derive(StructCodec)]
#[schema(name="cosmos.base.v1beta1.Coin")]
pub struct Coin {
    pub denom: String,
    #[proto(string, tag=2)] // tag could actually be inferred from field order, but shown for demonstration
    pub amount: u128,
}
```

Eventually, code generators may be implemented that take `.proto` files and generate this Rust code,
but for now we recommend following the Cosmos SDK's "expected keeper" pattern and just re-defining
types in Rust where needed.
Keeping client definitions close to where they're used avoids any of the pernicious issues
with versioning and dependencies that plague the Cosmos SDK in Golang.
Tooling is being developed to statically check the compatibility of types following the [`ixc_schema`] model
(see the `cosmosdk.io/schema/diff` package) so that type definitions from different Rust packages can be
compared for compatibility.

`#[proto]` and `#[schema]` annotations can also be used on arguments to handler functions
to configure the message names handlers correspond to.
(Note that even though the Cosmos SDK uses `service` definitions,
messages are actually routed by message name, not service method name.)
Ex:
```rust
#[module_api]
trait BankMsg {
   #[schema(name="cosmos.bank.v1beta1.MsgSend")]
   fn send(&self, ctx: &Context, 
           #[proto(string, msgv1_signer=true)] from: &Address,
           #[proto(string)] to: &Address,
           coins: &[Coin]) -> Response<()>;
}
```
