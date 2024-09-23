//! This module contains traits that must be implemented by types that can be used in the schema.

use alloc::borrow::ToOwned;
use bump_scope::BumpString;
use crate::decoder::{DecodeError, Decoder};
use crate::encoder::{EncodeError, Encoder};
use crate::types::*;

/// Any type used directly as a message function argument or struct field must implement this trait.
/// Unlike [`Value`] it takes a lifetime parameter so value may already be borrowed where it is
/// declared.
pub trait ArgValue<'a>
where
    Self: Sized + 'a,
{
    /// The type of the value.
    type Type: Type;

    /// In progress decoding state.
    type DecodeState: Default;

    /// Memory handle type returned if the decoded data borrows data which needed
    /// to be allocated and needs some owner.
    /// This handle is that owner.
    type MemoryHandle;

    /// Decode the value from the decoder.
    fn visit_decode_state<D: Decoder<'a>>(state: &'a mut Self::DecodeState, decoder: &'a mut D) -> Result<(), DecodeError> {
        unimplemented!("decode")
    }

    /// Finish decoding the value, return it and return the memory handle if needed.
    fn finish_decode_state(state: Self::DecodeState) -> Result<(Self, Option<Self::MemoryHandle>), DecodeError> {
        unimplemented!("finish")
    }

    /// Encode the value to the encoder.
    fn encode<E: Encoder>(&self, encoder: &mut E) -> Result<(), EncodeError> {
        unimplemented!("encode")
    }
}

/// This trait describes value types that are to be used as generic parameters
/// where there is no lifetime parameter available.
/// Any types implementing this trait relate themselves to a type implementing [`ArgValue`]
/// so that generic types taking a `Value` type parameter can use a borrowed value if possible.
pub trait Value {
    /// The possibly borrowable value type this type is related to.
    type MaybeBorrowed<'a>: ArgValue<'a>;
}

impl<'a> ArgValue<'a> for u8 {
    type Type = U8T;
    type DecodeState = u8;
    type MemoryHandle = ();
}
impl<'a> ArgValue<'a> for u16 {
    type Type = U16T;
    type DecodeState = u16;
    type MemoryHandle = ();
}
impl<'a> ArgValue<'a> for u32 {
    type Type = U32T;
    type DecodeState = u32;
    type MemoryHandle = ();
}
impl<'a> ArgValue<'a> for u64 {
    type Type = U64T;
    type DecodeState = u64;
    type MemoryHandle = ();
}
impl<'a> ArgValue<'a> for u128 {
    type Type = UIntNT<16>;
    type DecodeState = u128;
    type MemoryHandle = ();
}
impl<'a> ArgValue<'a> for i8 {
    type Type = I8T;
    type DecodeState = i8;
    type MemoryHandle = ();
}
impl<'a> ArgValue<'a> for i16 {
    type Type = I16T;
    type DecodeState = i16;
    type MemoryHandle = ();
}
impl<'a> ArgValue<'a> for i32 {
    type Type = I32T;
    type DecodeState = i32;
    type MemoryHandle = ();

    fn visit_decode_state<D: Decoder<'a>>(state: &'a mut Self::DecodeState, decoder: &'a mut D) -> Result<(), DecodeError> {
        todo!()
    }
}
impl<'a> ArgValue<'a> for i64 {
    type Type = I64T;
    type DecodeState = i64;
    type MemoryHandle = ();
}
impl<'a> ArgValue<'a> for i128 {
    type Type = IntNT<16>;
    type DecodeState = i128;
    type MemoryHandle = ();
}
impl<'a> ArgValue<'a> for bool {
    type Type = Bool;
    type DecodeState = bool;
    type MemoryHandle = ();
}
impl<'a> ArgValue<'a> for &'a str {
    type Type = StrT;
    type DecodeState = Option<Result<&'a str, BumpString<'a, 'a>>>;
    type MemoryHandle = BumpString<'a, 'a>;

    fn visit_decode_state<D: Decoder<'a>>(state: &'a mut Self::DecodeState, decoder: &'a mut D) -> Result<(), DecodeError> {
        todo!()
    }
}

#[cfg(feature = "std")]
impl<'a> ArgValue<'a> for alloc::string::String {
    type Type = StrT;
    type DecodeState = alloc::string::String;
    type MemoryHandle = ();
}

impl<'a> ArgValue<'a> for simple_time::Time {
    type Type = TimeT;
    type DecodeState = simple_time::Time;
    type MemoryHandle = ();
}
impl<'a> ArgValue<'a> for simple_time::Duration {
    type Type = DurationT;
    type DecodeState = simple_time::Duration;
    type MemoryHandle = ();
}
impl<'a, V: ArgValue<'a>> ArgValue<'a> for Option<V> {
    type Type = NullableT<V::Type>;
    type DecodeState = Option<V::DecodeState>;
    type MemoryHandle = V::MemoryHandle;
}
impl<'a, V: ArgValue<'a>> ArgValue<'a> for &'a [V]
where
    V::Type: ListElementType,
{
    type Type = ListT<V::Type>;
    type DecodeState = ();
    type MemoryHandle = ();
}

#[cfg(feature = "std")]
impl<'a, V: ArgValue<'a>> ArgValue<'a> for alloc::vec::Vec<V>
where
    V::Type: ListElementType,
{
    type Type = ListT<V::Type>;
    type DecodeState = ();
    type MemoryHandle = ();
}

#[cfg(feature = "address")]
impl<'a> ArgValue<'a> for ixc_message_api::Address {
    type Type = AddressT;
    type DecodeState = ixc_message_api::Address;
    type MemoryHandle = ();
}

#[cfg(feature = "arrayvec")]
impl<'a, T: Type, V: ArgValue<'a, T>, const N: usize> ArgValue<'a, ListT<T>> for arrayvec::ArrayVec<T, N> {}
#[cfg(feature = "arrayvec")]
impl<'a, const N: usize> ArgValue<'a, StrT> for arrayvec::ArrayString<T, N> {}

impl Value for u8 {
    type MaybeBorrowed<'a> = u8;
}
impl Value for u16 {
    type MaybeBorrowed<'a> = u16;
}
impl Value for u32 {
    type MaybeBorrowed<'a> = u32;
}
impl Value for u64 {
    type MaybeBorrowed<'a> = u64;
}
impl Value for u128 {
    type MaybeBorrowed<'a> = u128;
}
impl Value for i8 {
    type MaybeBorrowed<'a> = i8;
}
impl Value for i16 {
    type MaybeBorrowed<'a> = i16;
}
impl Value for i32 {
    type MaybeBorrowed<'a> = i32;
}
impl Value for i64 {
    type MaybeBorrowed<'a> = i64;
}
impl Value for i128 {
    type MaybeBorrowed<'a> = i128;
}
impl Value for bool {
    type MaybeBorrowed<'a> = bool;
}
impl Value for str {
    type MaybeBorrowed<'a> = &'a str;
}
impl Value for simple_time::Time {
    type MaybeBorrowed<'a> = simple_time::Time;
}
impl Value for simple_time::Duration {
    type MaybeBorrowed<'a> = simple_time::Duration;
}
impl Value for ixc_message_api::Address {
    type MaybeBorrowed<'a> = ixc_message_api::Address;
}
impl<V: Value> Value for Option<V> {
    type MaybeBorrowed<'a> = Option<V::MaybeBorrowed<'a>>;
}
impl<V: Value> Value for [V]
where
        for<'a> <<V as Value>::MaybeBorrowed<'a> as ArgValue<'a>>::Type: ListElementType,
{
    type MaybeBorrowed<'a> = &'a [V::MaybeBorrowed<'a>];
}

/// ResponseValue is a trait that must be implemented by types that can be used as the return value.
pub trait ResponseValue {
    /// The type that might be borrowed.
    #[cfg(feature = "std")]
    type MaybeBorrowed<'a>: ToOwned;
    #[cfg(not(feature = "std"))]
    type MaybeBorrowed<'a>;
}
impl ResponseValue for () {
    type MaybeBorrowed<'a> = ();
}
impl<V: Value> ResponseValue for V
where
        for<'a> V::MaybeBorrowed<'a>: ToOwned,
{
    type MaybeBorrowed<'a> = V::MaybeBorrowed<'a>;
}

