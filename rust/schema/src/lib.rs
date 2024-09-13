//! This crate defines the basic traits and types for schema and encoding.

/// StructCodec is the trait that should be derived to encode and decode a struct.
pub trait StructCodec {}

/// Value is the trait that should be implemented for all types that can be encoded and decoded.
pub trait Value<'a>
where
    Self: 'a,
{}
impl<'a> Value<'a> for u8 {}
impl<'a> Value<'a> for u16 {}
impl<'a> Value<'a> for u32 {}
impl<'a> Value<'a> for u64 {}
impl<'a> Value<'a> for u128 {}
impl<'a> Value<'a> for i8 {}
impl<'a> Value<'a> for i16 {}
impl<'a> Value<'a> for i32 {}
impl<'a> Value<'a> for i64 {}
impl<'a> Value<'a> for i128 {}
impl<'a> Value<'a> for bool {}
impl<'a> Value<'a> for &'a str {}
impl<'a, T: Value<'a>> Value<'a> for Option<T> {}
impl<'a> Value<'a> for &'a [u8] {}

#[cfg(feature = "address")]
impl<'a> Value<'a> for interchain_message_api::Address {}
#[cfg(feature = "address")]
impl<'a> Value<'a> for &'a interchain_message_api::Address {}

#[cfg(feature = "arrayvec")]
impl<'a, T: Value<'a>, const N: usize> Value<'a> for arrayvec::ArrayVec<T, N> {}
#[cfg(feature = "arrayvec")]
impl<'a, const N: usize> Value<'a> for arrayvec::ArrayString<T, N> {}

#[cfg(feature = "macros")]
#[allow(unused_imports)]
#[macro_use]
extern crate interchain_schema_macros;
#[cfg(feature = "macros")]
#[doc(inline)]
pub use interchain_schema_macros::*;
