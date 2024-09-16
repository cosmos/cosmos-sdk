//! This module contains the `Value` trait that should be implemented for all types that can be encoded and decoded.
use crate::StructCodec;
use crate::types::*;

/// Value is the trait that should be implemented for all types that can be encoded and decoded.
pub trait Value<'a>
where
    Self: 'a,
{
    /// The type of the value.
    type Type: Type;
}

impl<'a> Value<'a> for u8 {
    type Type = U8T;
}
impl<'a> Value<'a> for u16 {
    type Type = U16T;
}
impl<'a> Value<'a> for u32 {
    type Type = U32T;
}
impl<'a> Value<'a> for u64 {
    type Type = U64T;
}
impl<'a> Value<'a> for u128 {
    type Type = UIntNT<16>;
}
impl<'a> Value<'a> for i8 {
    type Type = I8T;
}
impl<'a> Value<'a> for i16 {
    type Type = I16T;
}
impl<'a> Value<'a> for i32 {
    type Type = I32T;
}
impl<'a> Value<'a> for i64 {
    type Type = I64T;
}
impl<'a> Value<'a> for i128 {
    type Type = IntNT<16>;
}
impl<'a> Value<'a> for bool {
    type Type = Bool;
}
impl<'a> Value<'a> for &'a str {
    type Type = StrT;
}
impl<'a> Value<'a> for simple_time::Time {
    type Type = TimeT;
}
impl<'a> Value<'a> for simple_time::Duration {
    type Type = DurationT;
}
impl<'a, V: Value<'a>> Value<'a> for Option<V> {
    type Type = NullableT<V::Type>;
}
impl<'a, V: Value<'a>> Value<'a> for &'a [V]
where
    V::Type: ListElementType,
{
    type Type = ListT<V::Type>;
}

#[cfg(feature = "address")]
impl<'a> Value<'a> for interchain_message_api::Address {
    type Type = AddressT;
}

#[cfg(feature = "arrayvec")]
impl<'a, T: Type, V: Value<'a, T>, const N: usize> Value<'a, ListT<T>> for arrayvec::ArrayVec<T, N> {}
#[cfg(feature = "arrayvec")]
impl<'a, const N: usize> Value<'a, StrT> for arrayvec::ArrayString<T, N> {}

