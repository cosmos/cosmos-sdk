This crate defines the encoding traits and schemas for the Interchain SDK which can be used in message
parameter and return types, serializable structs and enums, and objects in storage.

## Basic Types

The basic types supported intrinsically are:
* [`u8`], [`u16`], [`u32`], [`u64`], [`u128`]
* [`i8`], [`i16`], [`i32`], [`i64`], [`i128`]
* [`bool`]
* [`String`] or [`str`]
* [`interchain_message_api::Address`]
* [`simple_time::Time`] and [`simple_time::Duration`]
* [`Option<T>`] where `T` is any supported type (except for another `Option`)
* [`Vec<T>`] and `&[T]` where `T` is any supported type (except for another `Vec` or `&[T]`)

Custom types may be derived using the [`StructCodec`], [`EnumCodec`] and [`OneOfCodec`] trait
derive macros.

Depending on which context values are declared in, they may or may not need to be borrowed.
For instance, if we are using `str` as a function parameter, we must actually borrow it as `&str`.
The rules for this are as follows:
* if the type is a function parameter or struct field, it must implement [`value::MaybeBorrowed`]
  and can be borrowed as long as there is some implicit or explicit lifetime parameter available
* some generic types will require (such as `interchain_core::Response` and `state_objects::Item`)
  require [`value::Value`] instead and these types can't be borrowed explicitly even if the underlying
  implementation uses borrowing, (ex. we must write `Item<str>` instead of `Item<&str>` even though
  `Item.get()` and `Item.set()` actually take `&str`)