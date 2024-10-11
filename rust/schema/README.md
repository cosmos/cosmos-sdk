**WARNING: This is an API preview! Most code won't work or even type check properly!**

This crate defines the encoding traits and schemas for the Interchain SDK which can be used in message
parameter and return types, serializable structs and enums, and objects in storage.

## Basic Types

The basic types supported intrinsically are:
* [`u8`], [`u16`], [`u32`], [`u64`], [`u128`]
* [`i8`], [`i16`], [`i32`], [`i64`], [`i128`]
* [`bool`]
* [`String`](alloc::string::String) or [`&str`]
* [`ixc_message_api::AccountID`]
* [`simple_time::Time`] and [`simple_time::Duration`]
* [`Option<T>`] where `T` is any supported type (except for another `Option`)
* [`Vec<T>`](alloc::vec::Vec) or [`&[T]`](slice)

Custom types may be derived using the [`SchemaValue`] derive macro.
As a general rule, if a type implements [`SchemaValue`] it can be 
used as a function or struct parameter.

## Supported Encodings

Similar to the [Serde](https://serde.rs) framework, `ixc_schema` types can support multiple encodings
through different implementations of the [`encoder::Encoder`] and [`decoder::Decoder`] traits.
Unlike Serde, these traits are object safe and dynamic dispatch is used wherever possible
so that code size does not expand due to [Rust's mono-morphization of generics](https://rustwasm.github.io/book/reference/code-size.html#use-trait-objects-instead-of-generic-type-parameters)

Currently only a simple [native binary encoding](binary::NativeBinaryCodec) is supported but
protobuf, JSON, and encodings to other popular VM formats are planned in the future.

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
When data is decoded there is a ["memory manager"](mem::MemoryManager) that holds onto
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
