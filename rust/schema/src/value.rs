//! This module contains traits that must be implemented by types that can be used in the schema.

use ixc_message_api::handler::Allocator;
use ixc_message_api::packet::MessagePacket;
use crate::codec::Codec;
use crate::decoder::{DecodeError, Decoder};
use crate::encoder::{EncodeError, Encoder};
use crate::list::AllocatorVecBuilder;
use crate::mem::MemoryManager;
use crate::types::*;

/// Any type used directly as a message function argument or struct field must implement this trait.
/// Unlike [`ObjectFieldValue`] it takes a lifetime parameter so value may already be borrowed where it is
/// declared.
pub trait SchemaValue<'a>
where
    Self: 'a,
{
    /// The type of the value.
    type Type: Type;


    /// In progress decoding state.
    type DecodeState: Default;

    /// Decode the value from the decoder.
    fn visit_decode_state(_state: &mut Self::DecodeState, _decoder: &mut dyn Decoder<'a>) -> Result<(), DecodeError> {
        unimplemented!("decode")
    }

    /// Finish decoding the value, return it and return the memory handle if needed.
    fn finish_decode_state(_state: Self::DecodeState, _mem: &'a MemoryManager) -> Result<Self, DecodeError>
        where Self: Sized
    {
        unimplemented!("finish")
    }

    /// Encode the value to the encoder.
    fn encode(&self, _encoder: &mut dyn Encoder) -> Result<(), EncodeError> {
        unimplemented!("encode")
    }
}

impl<'a> SchemaValue<'a> for u8 {
    type Type = u8;
    type DecodeState = u8;
}
impl<'a> SchemaValue<'a> for u16 {
    type Type = u16;
    type DecodeState = u16;
}

impl<'a> SchemaValue<'a> for u32 {
    type Type = u32;
    type DecodeState = u32;

    fn visit_decode_state(state: &mut Self::DecodeState, decoder: &mut dyn Decoder<'a>) -> Result<(), DecodeError> {
        *state = decoder.decode_u32()?;
        Ok(())
    }

    fn finish_decode_state(state: Self::DecodeState, _: &'a MemoryManager) -> Result<Self, DecodeError> {
        Ok(state)
    }

    fn encode(&self, encoder: &mut dyn Encoder) -> Result<(), EncodeError> {
        encoder.encode_u32(*self)
    }
}

impl<'a> SchemaValue<'a> for u64 {
    type Type = u64;
    type DecodeState = u64;

    fn visit_decode_state(state: &mut Self::DecodeState, decoder: &mut dyn Decoder<'a>) -> Result<(), DecodeError> {
        *state = decoder.decode_u64()?;
        Ok(())
    }

    fn finish_decode_state(state: Self::DecodeState, _: &'a MemoryManager) -> Result<Self, DecodeError> {
        Ok(state)
    }

    fn encode(&self, encoder: &mut dyn Encoder) -> Result<(), EncodeError> {
        encoder.encode_u64(*self)
    }
}

impl<'a> SchemaValue<'a> for u128 {
    type Type = UIntNT<16>;
    type DecodeState = u128;

    fn visit_decode_state(state: &mut Self::DecodeState, decoder: &mut dyn Decoder<'a>) -> Result<(), DecodeError> {
        *state = decoder.decode_u128()?;
        Ok(())
    }

    fn finish_decode_state(state: Self::DecodeState, _: &'a MemoryManager) -> Result<Self, DecodeError> {
        Ok(state)
    }

    fn encode(&self, encoder: &mut dyn Encoder) -> Result<(), EncodeError> {
        encoder.encode_u128(*self)
    }
}

impl<'a> SchemaValue<'a> for i8 {
    type Type = i8;
    type DecodeState = i8;
}
impl<'a> SchemaValue<'a> for i16 {
    type Type = i16;
    type DecodeState = i16;
}
impl<'a> SchemaValue<'a> for i32 {
    type Type = i32;
    type DecodeState = i32;
}
impl<'a> SchemaValue<'a> for i64 {
    type Type = i64;
    type DecodeState = i64;
}
impl<'a> SchemaValue<'a> for i128 {
    type Type = IntNT<16>;
    type DecodeState = i128;
}
impl<'a> SchemaValue<'a> for bool {
    type Type = bool;
    type DecodeState = bool;
}
impl<'a> SchemaValue<'a> for &'a str {
    type Type = StrT;
    type DecodeState = &'a str;

    fn visit_decode_state(state: &mut Self::DecodeState, decoder: &mut dyn Decoder<'a>) -> Result<(), DecodeError> {
        *state = decoder.decode_borrowed_str()?;
        Ok(())
    }

    fn finish_decode_state(state: Self::DecodeState, _: &'a MemoryManager) -> Result<Self, DecodeError> {
        Ok(state)
    }

    fn encode(&self, encoder: &mut dyn Encoder) -> Result<(), EncodeError> {
        encoder.encode_str(self)
    }
}

#[cfg(feature = "std")]
impl<'a> SchemaValue<'a> for alloc::string::String {
    type Type = StrT;
    type DecodeState = alloc::string::String;

    fn visit_decode_state(state: &mut Self::DecodeState, decoder: &mut dyn Decoder<'a>) -> Result<(), DecodeError> {
        *state = decoder.decode_owned_str()?;
        Ok(())
    }

    fn finish_decode_state(state: Self::DecodeState, _: &'a MemoryManager) -> Result<Self, DecodeError> {
        Ok(state)
    }
}

impl<'a> SchemaValue<'a> for simple_time::Time {
    type Type = TimeT;
    type DecodeState = simple_time::Time;
}
impl<'a> SchemaValue<'a> for simple_time::Duration {
    type Type = DurationT;
    type DecodeState = simple_time::Duration;
}
impl<'a, V: SchemaValue<'a>> SchemaValue<'a> for Option<V> {
    type Type = Option<V::Type>;
    type DecodeState = Option<V::DecodeState>;
}

impl<'a> SchemaValue<'a> for &'a [u8] {
    type Type = BytesT;
    type DecodeState = &'a [u8];

    fn visit_decode_state(state: &mut Self::DecodeState, decoder: &mut dyn Decoder<'a>) -> Result<(), DecodeError> {
        todo!()
    }

    fn finish_decode_state(state: Self::DecodeState, mem: &'a MemoryManager) -> Result<Self, DecodeError> {
        todo!()
    }

    fn encode(&self, encoder: &mut dyn Encoder) -> Result<(), EncodeError> {
        todo!()
    }
}

/// A trait that must be implemented by value types that can be used as list elements.
pub trait ListElementValue<'a>: SchemaValue<'a>
where
    Self::Type: ListElementType,
{}

impl<'a, V: ListElementValue<'a>> SchemaValue<'a> for &'a [V]
where V::Type: ListElementType
{
    type Type = ListT<V::Type>;
    type DecodeState = AllocatorVecBuilder<'a, V>;

    fn visit_decode_state(state: &mut Self::DecodeState, decoder: &mut dyn Decoder<'a>) -> Result<(), DecodeError> {
        decoder.decode_list(state)
    }

    fn finish_decode_state(state: Self::DecodeState, mem_handle: &'a MemoryManager) -> Result<Self, DecodeError> {
        match state.xs {
            None => Ok(&[]),
            Some(xs) => Ok(mem_handle.unpack_slice(xs))
        }
    }

    fn encode(&self, encoder: &mut dyn Encoder) -> Result<(), EncodeError> {
        encoder.encode_list(self)
    }
}

impl<'a, V: SchemaValue<'a>> SchemaValue<'a> for allocator_api2::vec::Vec<V, &'a dyn allocator_api2::alloc::Allocator>
where
    V::Type: ListElementType,
{
    type Type = ListT<V::Type>;
    type DecodeState = AllocatorVecBuilder<'a, V>;

    fn visit_decode_state(state: &mut Self::DecodeState, decoder: &mut dyn Decoder<'a>) -> Result<(), DecodeError> {
        decoder.decode_list(state)
    }

    fn finish_decode_state(state: Self::DecodeState, mem_handle: &'a MemoryManager) -> Result<Self, DecodeError> {
        match state.xs {
            None => Ok(allocator_api2::vec::Vec::new_in(mem_handle)),
            Some(xs) => Ok(xs)
        }
    }
}

#[cfg(feature = "std")]
impl<'a, V: SchemaValue<'a>> SchemaValue<'a> for alloc::vec::Vec<V>
where
    V::Type: ListElementType,
{
    type Type = ListT<V::Type>;
    type DecodeState = alloc::vec::Vec<V>;

    fn encode(&self, encoder: &mut dyn Encoder) -> Result<(), EncodeError> {
        encoder.encode_list(self)
    }
}

impl<'a> SchemaValue<'a> for ixc_message_api::AccountID {
    type Type = AccountIdT;
    type DecodeState = ixc_message_api::AccountID;

    // fn visit_decode_state<D: Decoder<'a>>(state: &mut Self::DecodeState, decoder: &mut D) -> Result<(), DecodeError> {
    //     *state = decoder.decode_account_id()?;
    //     Ok(())
    // }

    fn finish_decode_state(state: Self::DecodeState, mem: &'a MemoryManager) -> Result<Self, DecodeError> {
        Ok(state)
    }

    fn encode(&self, encoder: &mut dyn Encoder) -> Result<(), EncodeError> {
        encoder.encode_account_id(*self)
    }
}

#[cfg(feature = "arrayvec")]
impl<'a, T: Type, V: SchemaValue<'a, T>, const N: usize> SchemaValue<'a, ListT<T>> for arrayvec::ArrayVec<T, N> {}
#[cfg(feature = "arrayvec")]
impl<'a, const N: usize> SchemaValue<'a, StrT> for arrayvec::ArrayString<T, N> {}


/// OptionalValue is a trait that must be implemented by types that can be used as the return value
/// or anywhere else where a value may or may not be necessary.
/// The unit type `()` is used to represent the absence of a value.
pub trait OptionalValue<'a> {
    /// The value type that is returned.
    type Value;

    /// Decode the value from the input.
    fn decode_value<C: Codec>(message_packet: &'a MessagePacket, memory_manager: &'a MemoryManager) -> Result<Self::Value, DecodeError>;

    /// Encode the value to the message packet.
    fn encode_value<C: Codec>(value: &Self::Value, message_packet: &'a mut MessagePacket, allocator: &'a dyn Allocator) -> Result<(), EncodeError>;
}

impl <'a> OptionalValue<'a> for () {
    type Value = ();

    fn decode_value<C: Codec>(message_packet: &'a MessagePacket, memory_manager: &'a MemoryManager) -> Result<Self::Value, DecodeError> {
        Ok(())
    }

    fn encode_value<C: Codec>(value: &Self::Value, message_packet: &'a mut MessagePacket, allocator: &'a dyn Allocator) -> Result<(), EncodeError> {
        Ok(())
    }
}

impl<'a, V: SchemaValue<'a>> OptionalValue<'a> for V
{
    type Value = V;

    fn decode_value<C: Codec>(message_packet: &'a MessagePacket, memory_manager: &'a MemoryManager) -> Result<Self::Value, DecodeError> {
        unsafe { C::decode_value(message_packet.header().out_pointer1.get(message_packet), memory_manager) }
    }

    fn encode_value<C: Codec>(value: &Self::Value, message_packet: &'a mut MessagePacket, allocator: &'a dyn Allocator) -> Result<(), EncodeError> {
        let res = C::encode_value(value, allocator)?;
        unsafe { message_packet.header_mut().out_pointer1.set_slice(res); }
        Ok(())
    }
}
