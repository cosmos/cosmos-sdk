//! This module contains traits that must be implemented by types that can be used in the schema.
use crate::types::*;

/// Any type used directly as a message function argument or struct field must implement this trait.
/// Unlike [`Value`] it takes a lifetime parameter so value may already be borrowed where it is
/// declared.
pub trait MaybeBorrowed<'a>
where
    Self: 'a,
{
    /// The type of the value.
    type Type: Type;
}

/// This trait describes value types that are to be used as generic parameters
/// where there is no lifetime parameter available.
/// Any types implementing this trait relate themselves to a type implementing [`MaybeBorrowed`]
/// so that generic types taking a `Value` type parameter can use a borrowed value if possible.
pub trait Value {
    /// The possibly borrowable value type this type is related to.
    type MaybeBorrowed<'a>: MaybeBorrowed<'a>;
}

impl<'a> MaybeBorrowed<'a> for u8 {
    type Type = U8T;
}
impl<'a> MaybeBorrowed<'a> for u16 {
    type Type = U16T;
}
impl<'a> MaybeBorrowed<'a> for u32 {
    type Type = U32T;
}
impl<'a> MaybeBorrowed<'a> for u64 {
    type Type = U64T;
}
impl<'a> MaybeBorrowed<'a> for u128 {
    type Type = UIntNT<16>;
}
impl<'a> MaybeBorrowed<'a> for i8 {
    type Type = I8T;
}
impl<'a> MaybeBorrowed<'a> for i16 {
    type Type = I16T;
}
impl<'a> MaybeBorrowed<'a> for i32 {
    type Type = I32T;
}
impl<'a> MaybeBorrowed<'a> for i64 {
    type Type = I64T;
}
impl<'a> MaybeBorrowed<'a> for i128 {
    type Type = IntNT<16>;
}
impl<'a> MaybeBorrowed<'a> for bool {
    type Type = Bool;
}
impl<'a> MaybeBorrowed<'a> for &'a str {
    type Type = StrT;
}
impl<'a> MaybeBorrowed<'a> for simple_time::Time {
    type Type = TimeT;
}
impl<'a> MaybeBorrowed<'a> for simple_time::Duration {
    type Type = DurationT;
}
impl<'a, V: MaybeBorrowed<'a>> MaybeBorrowed<'a> for Option<V> {
    type Type = NullableT<V::Type>;
}
impl<'a, V: MaybeBorrowed<'a>> MaybeBorrowed<'a> for &'a [V]
where
    V::Type: ListElementType,
{
    type Type = ListT<V::Type>;
}

#[cfg(feature = "address")]
impl<'a> MaybeBorrowed<'a> for interchain_message_api::Address {
    type Type = AddressT;
}

#[cfg(feature = "arrayvec")]
impl<'a, T: Type, V: MaybeBorrowed<'a, T>, const N: usize> MaybeBorrowed<'a, ListT<T>> for arrayvec::ArrayVec<T, N> {}
#[cfg(feature = "arrayvec")]
impl<'a, const N: usize> MaybeBorrowed<'a, StrT> for arrayvec::ArrayString<T, N> {}

