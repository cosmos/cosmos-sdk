use crate::StructCodec;
use crate::types::*;

/// Value is the trait that should be implemented for all types that can be encoded and decoded.
pub trait Value<'a, T: Type>
where
    Self: 'a,
{}
impl<'a> Value<'a, U8T> for u8 {}
impl<'a> Value<'a, U16T> for u16 {}
impl<'a> Value<'a, U32> for u32 {}
impl<'a> Value<'a, U64T> for u64 {}
impl<'a> Value<'a, U128T> for u128 {}
impl<'a> Value<'a, I8T> for i8 {}
impl<'a> Value<'a, I16T> for i16 {}
impl<'a> Value<'a, I32T> for i32 {}
impl<'a> Value<'a, I64T> for i64 {}
impl<'a> Value<'a, I128T> for i128 {}
impl<'a> Value<'a, Bool> for bool {}
impl<'a> Value<'a, StrT> for &'a str {}
impl<'a> Value<'a, TimeT> for simple_time::Time {}
impl<'a> Value<'a, DurationT> for simple_time::Duration {}
impl<'a, T: Type, V: Value<'a, T>> Value<'a, NullableT<T>> for Option<T> {}
impl<'a, S: StructCodec> Value<'a, StructT<S>> for S {}
impl<'a, T: Type, V: Value<'a, T>> Value<'a, ListT<T>> for &'a [T] {}

#[cfg(feature = "address")]
impl<'a> Value<'a, AddressT> for interchain_message_api::Address {}
#[cfg(feature = "address")]
impl<'a> Value<'a, AddressT> for &'a interchain_message_api::Address {}

#[cfg(feature = "arrayvec")]
impl<'a, T: Type, V: Value<'a, T>, const N: usize> Value<'a, ListT<T>> for arrayvec::ArrayVec<T, N> {}
#[cfg(feature = "arrayvec")]
impl<'a, const N: usize> Value<'a, StrT> for arrayvec::ArrayString<T, N> {}

