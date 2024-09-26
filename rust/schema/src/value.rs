//! This module contains traits that must be implemented by types that can be used in the schema.

use alloc::borrow::ToOwned;
use bump_scope::{BumpString, BumpVec};
use crate::decoder::{DecodeError, Decoder};
use crate::encoder::{EncodeError, Encoder};
use crate::list::SliceState;
use crate::mem::MemoryManager;
use crate::types::*;

/// Any type used directly as a message function argument or struct field must implement this trait.
/// Unlike [`AbstractValue`] it takes a lifetime parameter so value may already be borrowed where it is
/// declared.
pub trait Value<'a>
where
    Self: Sized + 'a,
{
    /// The type of the value.
    type Type: Type;

    /// In progress decoding state.
    type DecodeState: Default;

    /// Decode the value from the decoder.
    fn visit_decode_state<D: Decoder<'a>>(state: &mut Self::DecodeState, decoder: &mut D) -> Result<(), DecodeError> {
        unimplemented!("decode")
    }

    /// Finish decoding the value, return it and return the memory handle if needed.
    fn finish_decode_state(state: Self::DecodeState, mem_handle: &MemoryManager<'a, 'a>) -> Result<Self, DecodeError> {
        unimplemented!("finish")
    }

    /// Encode the value to the encoder.
    fn encode<E: Encoder>(&self, encoder: &mut E) -> Result<(), EncodeError> {
        unimplemented!("encode")
    }
}

/// This trait describes value types that are to be used as generic parameters
/// where there is no lifetime parameter available.
/// Any types implementing this trait relate themselves to a type implementing [`Value`]
/// so that generic types taking a `Value` type parameter can use a borrowed value if possible.
pub trait AbstractValue {
    /// The possibly borrowable value type this type is related to.
    type Value<'a>: Value<'a>;
}

impl<'a> Value<'a> for u8 {
    type Type = U8T;
    type DecodeState = u8;
}
impl<'a> Value<'a> for u16 {
    type Type = U16T;
    type DecodeState = u16;
}
impl<'a> Value<'a> for u32 {
    type Type = U32T;
    type DecodeState = u32;

    fn visit_decode_state<D: Decoder<'a>>(state: &mut Self::DecodeState, decoder: &mut D) -> Result<(), DecodeError> {
        *state = decoder.decode_u32()?;
        Ok(())
    }

    fn finish_decode_state(state: Self::DecodeState, mem_handle: &MemoryManager<'a, 'a>) -> Result<Self, DecodeError>  {
        Ok(state)
    }

    fn encode<E: Encoder>(&self, encoder: &mut E) -> Result<(), EncodeError> {
        encoder.encode_u32(*self)
    }
}
impl<'a> Value<'a> for u64 {
    type Type = U64T;
    type DecodeState = u64;
}
impl<'a> Value<'a> for u128 {
    type Type = UIntNT<16>;
    type DecodeState = u128;

    fn visit_decode_state<D: Decoder<'a>>(state: &mut Self::DecodeState, decoder: &mut D) -> Result<(), DecodeError> {
        *state = decoder.decode_u128()?;
        Ok(())
    }

    fn finish_decode_state(state: Self::DecodeState, mem_handle: &MemoryManager<'a, 'a>) -> Result<Self, DecodeError>  {
        Ok(state)
    }

    fn encode<E: Encoder>(&self, encoder: &mut E) -> Result<(), EncodeError> {
        encoder.encode_u128(*self)
    }
}
impl<'a> Value<'a> for i8 {
    type Type = I8T;
    type DecodeState = i8;
}
impl<'a> Value<'a> for i16 {
    type Type = I16T;
    type DecodeState = i16;
}
impl<'a> Value<'a> for i32 {
    type Type = I32T;
    type DecodeState = i32;
}
impl<'a> Value<'a> for i64 {
    type Type = I64T;
    type DecodeState = i64;
}
impl<'a> Value<'a> for i128 {
    type Type = IntNT<16>;
    type DecodeState = i128;
}
impl<'a> Value<'a> for bool {
    type Type = Bool;
    type DecodeState = bool;
}
impl<'a> Value<'a> for &'a str {
    type Type = StrT;
    type DecodeState = &'a str;

    fn visit_decode_state<D: Decoder<'a>>(state: &mut Self::DecodeState, decoder: &mut D) -> Result<(), DecodeError> {
        *state = decoder.decode_borrowed_str()?;
        Ok(())
    }

    fn finish_decode_state(state: Self::DecodeState, mem_handle: &MemoryManager<'a, 'a>) -> Result<Self, DecodeError> {
        Ok(state)
    }

    fn encode<E: Encoder>(&self, encoder: &mut E) -> Result<(), EncodeError> {
        encoder.encode_str(self)
    }
}

#[cfg(feature = "std")]
impl<'a> Value<'a> for alloc::string::String {
    type Type = StrT;
    type DecodeState = alloc::string::String;

    fn visit_decode_state<D: Decoder<'a>>(state: &mut Self::DecodeState, decoder: &mut D) -> Result<(), DecodeError> {
        *state = decoder.decode_owned_str()?;
        Ok(())
    }

    fn finish_decode_state(state: Self::DecodeState, mem_handle: &MemoryManager<'a, 'a>) -> Result<Self, DecodeError> {
        Ok(state)
    }
}

impl<'a> Value<'a> for simple_time::Time {
    type Type = TimeT;
    type DecodeState = simple_time::Time;
}
impl<'a> Value<'a> for simple_time::Duration {
    type Type = DurationT;
    type DecodeState = simple_time::Duration;
}
impl<'a, V: Value<'a>> Value<'a> for Option<V> {
    type Type = NullableT<V::Type>;
    type DecodeState = Option<V::DecodeState>;
}
impl<'a, V: Value<'a>> Value<'a> for &'a [V]
where
    V::Type: ListElementType,
{
    type Type = ListT<V::Type>;
    type DecodeState = SliceState<'a, V>;

    fn visit_decode_state<D: Decoder<'a>>(state: &mut Self::DecodeState, decoder: &mut D) -> Result<(), DecodeError> {
        decoder.decode_list(state)
    }

    fn finish_decode_state(state: Self::DecodeState, mem_handle: &MemoryManager<'a, 'a>) -> Result<Self, DecodeError> {
        match state.xs {
            None => Ok(&[]),
            Some(xs) => Ok(mem_handle.unpack_slice(xs))
        }
    }
}

#[cfg(feature = "std")]
impl<'a, V: Value<'a>> Value<'a> for alloc::vec::Vec<V>
where
    V::Type: ListElementType,
{
    type Type = ListT<V::Type>;
    type DecodeState = alloc::vec::Vec<V>;

    fn encode<E: Encoder>(&self, encoder: &mut E) -> Result<(), EncodeError> {
        encoder.encode_list_slice(self.as_slice())
    }
}

#[cfg(feature = "address")]
impl<'a> Value<'a> for ixc_message_api::AccountID {
    type Type = AccountIDT;
    type DecodeState = ixc_message_api::AccountID;
}

#[cfg(feature = "arrayvec")]
impl<'a, T: Type, V: Value<'a, T>, const N: usize> Value<'a, ListT<T>> for arrayvec::ArrayVec<T, N> {}
#[cfg(feature = "arrayvec")]
impl<'a, const N: usize> Value<'a, StrT> for arrayvec::ArrayString<T, N> {}

impl AbstractValue for u8 {
    type Value<'a> = u8;
}
impl AbstractValue for u16 {
    type Value<'a> = u16;
}
impl AbstractValue for u32 {
    type Value<'a> = u32;
}
impl AbstractValue for u64 {
    type Value<'a> = u64;
}
impl AbstractValue for u128 {
    type Value<'a> = u128;
}
impl AbstractValue for i8 {
    type Value<'a> = i8;
}
impl AbstractValue for i16 {
    type Value<'a> = i16;
}
impl AbstractValue for i32 {
    type Value<'a> = i32;
}
impl AbstractValue for i64 {
    type Value<'a> = i64;
}
impl AbstractValue for i128 {
    type Value<'a> = i128;
}
impl AbstractValue for bool {
    type Value<'a> = bool;
}
impl AbstractValue for str {
    type Value<'a> = &'a str;
}
impl AbstractValue for simple_time::Time {
    type Value<'a> = simple_time::Time;
}
impl AbstractValue for simple_time::Duration {
    type Value<'a> = simple_time::Duration;
}
impl AbstractValue for ixc_message_api::AccountID {
    type Value<'a> = ixc_message_api::AccountID;
}
impl<V: AbstractValue> AbstractValue for Option<V> {
    type Value<'a> = Option<V::Value<'a>>;
}
impl<V: AbstractValue> AbstractValue for [V]
where
        for<'a> <<V as AbstractValue>::Value<'a> as Value<'a>>::Type: ListElementType,
{
    type Value<'a> = &'a [V::Value<'a>];
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
impl<V: AbstractValue> ResponseValue for V
where
        for<'a> V::Value<'a>: ToOwned,
{
    type MaybeBorrowed<'a> = V::Value<'a>;
}
