use crate::buffer::{Reader, Writer};
use crate::decoder::DecodeError;
use crate::encoder::EncodeError;
use crate::mem::MemoryManager;
use crate::stateobject::key_field::KeyFieldValue;
use crate::stateobject::value::ObjectValue;

/// This trait is implemented for types that can be used as keys in state objects.
pub trait ObjectKey: ObjectValue {
    /// Encode the key.
    fn encode<'a, W: Writer>(key: Self::In<'a>, writer: &mut W) -> Result<(), EncodeError>;

    fn decode<'a>(input: &'a [u8], memory_manager: &'a MemoryManager) -> Result<Self::Out<'a>, DecodeError>;
}

impl ObjectKey for () {
    fn encode<'a, W: Writer>(key: Self::In<'a>, writer: &mut W) -> Result<(), EncodeError> {
        Ok(())
    }

    fn decode<'a>(input: &'a [u8], memory_manager: &'a MemoryManager) -> Result<Self::Out<'a>, DecodeError> {
        Ok(())
    }
}

impl<A: KeyFieldValue> ObjectKey for A {
    fn encode<'a, W: Writer>(key: Self::In<'a>, writer: &mut W) -> Result<(), EncodeError> {
        A::encode_terminal(key, writer)
    }

    fn decode<'a>(input: &'a [u8], memory_manager: &'a MemoryManager) -> Result<Self::Out<'a>, DecodeError> {
        let mut reader = input;
        let a = A::decode_terminal(&mut reader, memory_manager)?;
        reader.done()?;
        Ok(a)
    }
}
impl<A: KeyFieldValue> ObjectKey for (A,) {
    fn encode<'a, W: Writer>(key: Self::In<'a>, writer: &mut W) -> Result<(), EncodeError> {
        A::encode(key.0, writer)
    }

    fn decode<'a>(input: &'a [u8], memory_manager: &'a MemoryManager) -> Result<Self::Out<'a>, DecodeError> {
        let mut reader = input;
        let a = A::decode(&mut reader, memory_manager)?;
        reader.done()?;
        Ok((a,))
    }
}

impl<A: KeyFieldValue, B: KeyFieldValue> ObjectKey for (A, B) {
    fn encode<'a, W: Writer>(key: Self::In<'a>, writer: &mut W) -> Result<(), EncodeError> {
        A::encode(key.0, writer)?;
        B::encode_terminal(key.1, writer)
    }

    fn decode<'a>(input: &'a [u8], memory_manager: &'a MemoryManager) -> Result<Self::Out<'a>, DecodeError> {
        let mut reader = input;
        let a = A::decode(&mut reader, memory_manager)?;
        let b = B::decode_terminal(&mut reader, memory_manager)?;
        reader.done()?;
        Ok((a, b))
    }
}

impl<A: KeyFieldValue, B: KeyFieldValue, C: KeyFieldValue> ObjectKey for (A, B, C) {
    fn encode<'a, W: Writer>(key: Self::In<'a>, writer: &mut W) -> Result<(), EncodeError> {
        A::encode(key.0, writer)?;
        B::encode(key.1, writer)?;
        C::encode_terminal(key.2, writer)
    }

    fn decode<'a>(input: &'a [u8], memory_manager: &'a MemoryManager) -> Result<Self::Out<'a>, DecodeError> {
        let mut reader = input;
        let a = A::decode(&mut reader, memory_manager)?;
        let b = B::decode(&mut reader, memory_manager)?;
        let c = C::decode_terminal(&mut reader, memory_manager)?;
        reader.done()?;
        Ok((a, b, c))
    }
}

impl<A: KeyFieldValue, B: KeyFieldValue, C: KeyFieldValue, D: KeyFieldValue> ObjectKey for (A, B, C, D) {
    fn encode<'a, W: Writer>(key: Self::In<'a>, writer: &mut W) -> Result<(), EncodeError> {
        A::encode(key.0, writer)?;
        B::encode(key.1, writer)?;
        C::encode(key.2, writer)?;
        D::encode_terminal(key.3, writer)
    }

    fn decode<'a>(input: &'a [u8], memory_manager: &'a MemoryManager) -> Result<Self::Out<'a>, DecodeError> {
        let mut reader = input;
        let a = A::decode(&mut reader, memory_manager)?;
        let b = B::decode(&mut reader, memory_manager)?;
        let c = C::decode(&mut reader, memory_manager)?;
        let d = D::decode_terminal(&mut reader, memory_manager)?;
        reader.done()?;
        Ok((a, b, c, d))
    }
}

