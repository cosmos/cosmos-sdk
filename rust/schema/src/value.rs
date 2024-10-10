//! This module contains traits that must be implemented by types that can be used in the schema.

use ixc_message_api::handler::Allocator;
use ixc_message_api::packet::MessagePacket;
use ixc_schema::buffer::WriterFactory;
use crate::codec::{decode_value, Codec, ValueDecodeVisitor};
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
    fn visit_decode_state(state: &mut Self::DecodeState, decoder: &mut dyn Decoder<'a>) -> Result<(), DecodeError>;

    /// Finish decoding the value, return it and return the memory handle if needed.
    fn finish_decode_state(state: Self::DecodeState, mem: &'a MemoryManager) -> Result<Self, DecodeError>
    where
        Self: Sized;

    /// Encode the value to the encoder.
    fn encode(&self, encoder: &mut dyn Encoder) -> Result<(), EncodeError>;
}

impl<'a> SchemaValue<'a> for u8 {
    type Type = u8;
    type DecodeState = u8;

    fn visit_decode_state(state: &mut Self::DecodeState, decoder: &mut dyn Decoder<'a>) -> Result<(), DecodeError> {
        *state = decoder.decode_u8()?;
        Ok(())
    }

    fn finish_decode_state(state: Self::DecodeState, mem: &'a MemoryManager) -> Result<Self, DecodeError>
    where
        Self: Sized
    {
        Ok(state)
    }

    fn encode(&self, encoder: &mut dyn Encoder) -> Result<(), EncodeError> {
        encoder.encode_u8(*self)
    }
}
impl<'a> SchemaValue<'a> for u16 {
    type Type = u16;
    type DecodeState = u16;

    fn visit_decode_state(state: &mut Self::DecodeState, decoder: &mut dyn Decoder<'a>) -> Result<(), DecodeError> {
        *state = decoder.decode_u16()?;
        Ok(())
    }

    fn finish_decode_state(state: Self::DecodeState, _: &'a MemoryManager) -> Result<Self, DecodeError> {
        Ok(state)
    }

    fn encode(&self, encoder: &mut dyn Encoder) -> Result<(), EncodeError> {
        encoder.encode_u16(*self)
    }
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

    fn visit_decode_state(state: &mut Self::DecodeState, decoder: &mut dyn Decoder<'a>) -> Result<(), DecodeError> {
        *state = decoder.decode_i8()?;
        Ok(())
    }

    fn finish_decode_state(state: Self::DecodeState, _: &'a MemoryManager) -> Result<Self, DecodeError> {
        Ok(state)
    }

    fn encode(&self, encoder: &mut dyn Encoder) -> Result<(), EncodeError> {
        encoder.encode_i8(*self)
    }
}

impl<'a> SchemaValue<'a> for i16 {
    type Type = i16;
    type DecodeState = i16;

    fn visit_decode_state(state: &mut Self::DecodeState, decoder: &mut dyn Decoder<'a>) -> Result<(), DecodeError> {
        *state = decoder.decode_i16()?;
        Ok(())
    }

    fn finish_decode_state(state: Self::DecodeState, _: &'a MemoryManager) -> Result<Self, DecodeError> {
        Ok(state)
    }

    fn encode(&self, encoder: &mut dyn Encoder) -> Result<(), EncodeError> {
        encoder.encode_i16(*self)
    }
}

impl<'a> SchemaValue<'a> for i32 {
    type Type = i32;
    type DecodeState = i32;

    fn visit_decode_state(state: &mut Self::DecodeState, decoder: &mut dyn Decoder<'a>) -> Result<(), DecodeError> {
        *state = decoder.decode_i32()?;
        Ok(())
    }

    fn finish_decode_state(state: Self::DecodeState, _: &'a MemoryManager) -> Result<Self, DecodeError> {
        Ok(state)
    }

    fn encode(&self, encoder: &mut dyn Encoder) -> Result<(), EncodeError> {
        encoder.encode_i32(*self)
    }
}

impl<'a> SchemaValue<'a> for i64 {
    type Type = i64;
    type DecodeState = i64;

    fn visit_decode_state(state: &mut Self::DecodeState, decoder: &mut dyn Decoder<'a>) -> Result<(), DecodeError> {
        *state = decoder.decode_i64()?;
        Ok(())
    }

    fn finish_decode_state(state: Self::DecodeState, _: &'a MemoryManager) -> Result<Self, DecodeError> {
        Ok(state)
    }

    fn encode(&self, encoder: &mut dyn Encoder) -> Result<(), EncodeError> {
        encoder.encode_i64(*self)
    }
}

impl<'a> SchemaValue<'a> for i128 {
    type Type = IntNT<16>;
    type DecodeState = i128;

    fn visit_decode_state(state: &mut Self::DecodeState, decoder: &mut dyn Decoder<'a>) -> Result<(), DecodeError> {
        *state = decoder.decode_i128()?;
        Ok(())
    }

    fn finish_decode_state(state: Self::DecodeState, _: &'a MemoryManager) -> Result<Self, DecodeError> {
        Ok(state)
    }

    fn encode(&self, encoder: &mut dyn Encoder) -> Result<(), EncodeError> {
        encoder.encode_i128(*self)
    }
}

impl<'a> SchemaValue<'a> for bool {
    type Type = bool;
    type DecodeState = bool;

    fn visit_decode_state(state: &mut Self::DecodeState, decoder: &mut dyn Decoder<'a>) -> Result<(), DecodeError> {
        *state = decoder.decode_bool()?;
        Ok(())
    }

    fn finish_decode_state(state: Self::DecodeState, _: &'a MemoryManager) -> Result<Self, DecodeError> {
        Ok(state)
    }

    fn encode(&self, encoder: &mut dyn Encoder) -> Result<(), EncodeError> {
        encoder.encode_bool(*self)
    }
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

    fn encode(&self, encoder: &mut dyn Encoder) -> Result<(), EncodeError> {
        encoder.encode_str(self)
    }
}

impl<'a> SchemaValue<'a> for simple_time::Time {
    type Type = TimeT;
    type DecodeState = simple_time::Time;

    fn visit_decode_state(state: &mut Self::DecodeState, decoder: &mut dyn Decoder<'a>) -> Result<(), DecodeError> {
        *state = decoder.decode_time()?;
        Ok(())
    }

    fn finish_decode_state(state: Self::DecodeState, _: &'a MemoryManager) -> Result<Self, DecodeError> {
        Ok(state)
    }

    fn encode(&self, encoder: &mut dyn Encoder) -> Result<(), EncodeError> {
        encoder.encode_time(*self)
    }
}

impl<'a> SchemaValue<'a> for simple_time::Duration {
    type Type = DurationT;
    type DecodeState = simple_time::Duration;

    fn visit_decode_state(state: &mut Self::DecodeState, decoder: &mut dyn Decoder<'a>) -> Result<(), DecodeError> {
        *state = decoder.decode_duration()?;
        Ok(())
    }

    fn finish_decode_state(state: Self::DecodeState, _: &'a MemoryManager) -> Result<Self, DecodeError> {
        Ok(state)
    }

    fn encode(&self, encoder: &mut dyn Encoder) -> Result<(), EncodeError> {
        encoder.encode_duration(*self)
    }
}


impl<'a, V: SchemaValue<'a>> SchemaValue<'a> for Option<V> {
    type Type = Option<V::Type>;
    type DecodeState = Option<V::DecodeState>;

    fn visit_decode_state(state: &mut Self::DecodeState, decoder: &mut dyn Decoder<'a>) -> Result<(), DecodeError> {
        struct Visitor<'b, U:SchemaValue<'b>>(U::DecodeState);
        // TODO can we reduce the duplication between this and codec::decode_value?
        impl <'b, U:SchemaValue<'b>> ValueDecodeVisitor<'b> for Visitor<'b, U> {
            fn decode(&mut self, decoder: &mut dyn Decoder<'b>) -> Result<(), DecodeError> {
                U::visit_decode_state(&mut self.0, decoder)
            }
        }
        let mut visitor = Visitor::<V>(V::DecodeState::default());
        if decoder.decode_option(&mut visitor)? {
            *state = Some(visitor.0);
        }
        Ok(())
    }

    fn finish_decode_state(state: Self::DecodeState, mem: &'a MemoryManager) -> Result<Self, DecodeError>
    where
        Self: Sized
    {
        state.map(|state| V::finish_decode_state(state, mem)).transpose()
    }

    fn encode(&self, encoder: &mut dyn Encoder) -> Result<(), EncodeError> {
        match self {
            Some(value) => {
                encoder.encode_option(Some(value))
            }
            None => {
                encoder.encode_option(None)
            }
        }
    }
}

impl<'a> SchemaValue<'a> for &'a [u8] {
    type Type = BytesT;
    type DecodeState = &'a [u8];

    fn visit_decode_state(state: &mut Self::DecodeState, decoder: &mut dyn Decoder<'a>) -> Result<(), DecodeError> {
        *state = decoder.decode_borrowed_bytes()?;
        Ok(())
    }

    fn finish_decode_state(state: Self::DecodeState, mem: &'a MemoryManager) -> Result<Self, DecodeError> {
        Ok(state)
    }

    fn encode(&self, encoder: &mut dyn Encoder) -> Result<(), EncodeError> {
        encoder.encode_bytes(self)
    }
}

impl<'a> SchemaValue<'a> for alloc::vec::Vec<u8> {
    type Type = BytesT;
    type DecodeState = alloc::vec::Vec<u8>;

    fn visit_decode_state(state: &mut Self::DecodeState, decoder: &mut dyn Decoder<'a>) -> Result<(), DecodeError> {
        *state = decoder.decode_owned_bytes()?;
        Ok(())
    }

    fn finish_decode_state(state: Self::DecodeState, mem: &'a MemoryManager) -> Result<Self, DecodeError> {
        Ok(state)
    }

    fn encode(&self, encoder: &mut dyn Encoder) -> Result<(), EncodeError> {
        encoder.encode_bytes(self)
    }
}

/// A trait that must be implemented by value types that can be used as list elements.
pub trait ListElementValue<'a>: SchemaValue<'a>
where
    Self::Type: ListElementType,
{}

impl<'a, V: ListElementValue<'a>> SchemaValue<'a> for &'a [V]
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
            None => Ok(&[]),
            Some(xs) => Ok(mem_handle.unpack_slice(xs))
        }
    }

    fn encode(&self, encoder: &mut dyn Encoder) -> Result<(), EncodeError> {
        encoder.encode_list(self)
    }
}

impl<'a, V: ListElementValue<'a>> SchemaValue<'a> for allocator_api2::vec::Vec<V, &'a dyn allocator_api2::alloc::Allocator>
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

    fn encode(&self, encoder: &mut dyn Encoder) -> Result<(), EncodeError> {
        // encoder.encode_list(self)
        todo!()
    }
}

#[cfg(feature = "std")]
impl<'a, V: ListElementValue<'a>> SchemaValue<'a> for alloc::vec::Vec<V>
where
    V::Type: ListElementType,
{
    type Type = ListT<V::Type>;
    type DecodeState = alloc::vec::Vec<V>;

    fn visit_decode_state(state: &mut Self::DecodeState, decoder: &mut dyn Decoder<'a>) -> Result<(), DecodeError> {
        decoder.decode_list(state)
    }

    fn finish_decode_state(state: Self::DecodeState, mem: &'a MemoryManager) -> Result<Self, DecodeError>
    where
        Self: Sized
    {
        Ok(state)
    }

    fn encode(&self, encoder: &mut dyn Encoder) -> Result<(), EncodeError> {
        encoder.encode_list(self)
    }
}

impl<'a> SchemaValue<'a> for ixc_message_api::AccountID {
    type Type = AccountIdT;
    type DecodeState = ixc_message_api::AccountID;

    fn visit_decode_state(state: &mut Self::DecodeState, decoder: &mut dyn Decoder<'a>) -> Result<(), DecodeError> {
        *state = decoder.decode_account_id()?;
        Ok(())
    }

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

    /// Decode the value.
    fn decode_value(cdc: &dyn Codec, data: &'a [u8], memory_manager: &'a MemoryManager) -> Result<Self::Value, DecodeError>;

    /// Encode the value.
    fn encode_value<'b>(cdc: &dyn Codec, value: &Self::Value, writer_factory: &'b dyn WriterFactory) -> Result<Option<&'b [u8]>, EncodeError>;
}

impl<'a> OptionalValue<'a> for () {
    type Value = ();

    fn decode_value(cdc: &dyn Codec, data: &'a [u8], memory_manager: &'a MemoryManager) -> Result<Self::Value, DecodeError> {
        Ok(())
    }

    fn encode_value<'b>(cdc: &dyn Codec, value: &Self::Value, writer_factory: &'b dyn WriterFactory) -> Result<Option<&'b [u8]>, EncodeError> {
        Ok(None)
    }
}

impl<'a, V: SchemaValue<'a>> OptionalValue<'a> for V
{
    type Value = V;

    fn decode_value(cdc: &dyn Codec, data: &'a [u8], memory_manager: &'a MemoryManager) -> Result<Self::Value, DecodeError> {
        unsafe { decode_value(cdc, data, memory_manager) }
    }

    fn encode_value<'b>(cdc: &dyn Codec, value: &Self::Value, writer_factory: &'b dyn WriterFactory) -> Result<Option<&'b [u8]>, EncodeError> {
        Ok(Some(cdc.encode_value(value, writer_factory)?))
    }
}

impl <'a> ListElementValue<'a> for u16 {}
impl <'a> ListElementValue<'a> for u32 {}
impl <'a> ListElementValue<'a> for u64 {}
impl <'a> ListElementValue<'a> for u128 {}
impl <'a> ListElementValue<'a> for i8 {}
impl <'a> ListElementValue<'a> for i16 {}
impl <'a> ListElementValue<'a> for i32 {}
impl <'a> ListElementValue<'a> for i64 {}
impl <'a> ListElementValue<'a> for i128 {}
impl <'a> ListElementValue<'a> for bool {}
impl <'a> ListElementValue<'a> for &'a str {}
#[cfg(feature = "std")]
impl <'a> ListElementValue<'a> for alloc::string::String {}
impl <'a> ListElementValue<'a> for &'a [u8] {}
#[cfg(feature = "std")]
impl <'a> ListElementValue<'a> for alloc::vec::Vec<u8> {}
impl <'a> ListElementValue<'a> for simple_time::Time {}
impl <'a> ListElementValue<'a> for simple_time::Duration {}