**WARNING: This is an API preview! Most code won't work or even type check properly!**

This crate defines the encoding traits and schemas for the Interchain SDK which can be used in message
parameter and return types, serializable structs and enums, and objects in storage.

## Basic Types

The basic types supported intrinsically are:
* [`u8`], [`u16`], [`u32`], [`u64`], [`u128`]
* [`i8`], [`i16`], [`i32`], [`i64`], [`i128`]
* [`bool`]
* [`String`](alloc::string::String), [`&str`] or [`Cow<str>`](alloc::borrow::Cow)
* [`ixc_message_api::Address`]
* [`simple_time::Time`] and [`simple_time::Duration`]
* [`Option<T>`] where `T` is any supported type (except for another `Option`)
* [`Vec<T>`](alloc::vec::Vec), [`&[T]`](slice) or [`Cow<[T]>`](alloc::borrow::Cow) where `T` is any supported type (except for another `Vec` or `&[T]`)

Custom types may be derived using the [`StructCodec`], [`EnumCodec`] and [`OneOfCodec`] trait
derive macros.

Depending on which context values are declared in, they may or may not need to be borrowed.
For instance, if we are using `str` as a function parameter, we must actually borrow it as `&str`.
The rules for this are as follows:
* if the type is a function parameter or struct field, it must implement [`value::Value`]
  and can be borrowed as long as there is some implicit or explicit lifetime parameter available
* some generic types will require (such as `interchain_core::Response` and `state_objects::Item`)
  require [`value::AbstractValue`] instead and these types can't be borrowed explicitly even if the underlying
  implementation uses borrowing, (ex. we must write `Item<str>` instead of `Item<&str>` even though
  `Item.get()` and `Item.set()` actually take `&str`)

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
  
## Memory Management

If you read the list of supported types carefully, you may notice that borrowed strings
and slices are both supported.
This means that we can define types with lifetimes and deserialize borrowed data, just like in the
popular [Serde](https://serde.rs/lifetimes.html) framework.
Ex:

```rust
pub struct Coin<'a> {
  pub denom: &'a str,
  pub amount: u128,  
}
```

We improve upon Serde's approach with a memory management system to properly deal with
cases where we simply cannot borrow directly from the input data.
In Serde, if we deserialize the above `Coin` structure from JSON input, it wil
work some of the time but will fail with an error whenever the input data contains
JSON escape characters (ex. `"foo\tbar"`).
This is because a borrowed data structure must have some owner, but in Serde the only
possible owner is the input itself.

In `ixc_schema` we use a [bump allocator](https://en.wikipedia.org/wiki/Region-based_memory_management) under the hood
to hold onto any intermediate allocation that must take place to be able
to borrow not just strings, but slices of any sort of data.
When data is decoded there is a "memory manager" that holds onto
the original input data and any temporary allocations and then
deallocates them in mass when we're done with the data.
Because decoding is a phase-oriented operation, this model works well
and is very efficient.
Because of this property, `ixc_schema` can support complex data structures
and manage memory correctly, without needing to interact with the general
purpose Rust global allocator.
With this model, in the future it will likely be possible to build
crates targeting virtual machines such as WebAssembly in `no_std` mode
without any sort of global allocator.

To make the best use of the built-in bump allocator, it is recommended
to avoid uses of [`String`](alloc::string::String) and [`Vec<T>`](alloc::vec::Vec) in favor of
[`&str`] and [`&[T]`](slice) respectively.
You can also use [`Cow<str>`](alloc::borrow::Cow) and [`Cow<[T]>`](alloc::borrow::Cow)
to have the optionality to use borrowed or owned data.

NOTE: The bump allocator system is currently tightly tied to the [bump-scope](https://docs.rs/bump-scope) crate.
This crate, while relatively new, has several important improvements over the
more well-known [bumpalo](https://docs.rs/bumpalo) crate, such as support
for properly calling drop when deallocating and allowing the underlying allocator
to be configured.
It also appears to be quite well tested and researched, although ideally we
could also customize the underlying chunk size.
Before `ixc_schema` is stable, we will need to decide whether to continue using
`bump-scope` or create a custom implementation to avoid the risk of the third-party dependency.