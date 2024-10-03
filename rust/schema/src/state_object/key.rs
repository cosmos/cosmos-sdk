use crate::buffer::{Reader, Writer, WriterFactory};
use crate::decoder::DecodeError;
use crate::encoder::EncodeError;
use crate::mem::MemoryManager;
use crate::state_object::value::ObjectValue;
use crate::state_object::KeyFieldValue;

/// Encode an object key with the given prefix.
pub fn encode_object_key<'a, 'b, K: ObjectKey>(prefix: &[u8], key: &K::In<'a>, writer_factory: &'b dyn WriterFactory) -> Result<&'b [u8], EncodeError> {
    let out_size = <K as ObjectKey>::out_size(key) + prefix.len();
    let mut writer = writer_factory.new_reverse(out_size)?;
    <K as ObjectKey>::encode(key, &mut writer)?;
    // write the prefix last because we are encoding in reverse order
    writer.write(prefix)?;
    Ok(writer.finish())
}

/// Decode an object key. This function assumes that the input has already had any prefix stripped.
pub fn decode_object_key<'a, K: ObjectKey>(input: &'a [u8], memory_manager: &'a MemoryManager) -> Result<K::Out<'a>, DecodeError> {
    <K as ObjectKey>::decode(input, memory_manager)
}

/// This trait is implemented for types that can be used as keys in state objects.
pub trait ObjectKey: ObjectValue {
    /// Encode the key.
    fn encode<'a, W: Writer>(key: &Self::In<'a>, writer: &mut W) -> Result<(), EncodeError>;

    /// Decode the key.
    fn decode<'a>(input: &'a [u8], memory_manager: &'a MemoryManager) -> Result<Self::Out<'a>, DecodeError>;

    /// Compute the output buffer size for the key.
    fn out_size<'a>(key: &Self::In<'a>) -> usize;
}

impl ObjectKey for () {
    fn encode<'a, W: Writer>(key: &Self::In<'a>, writer: &mut W) -> Result<(), EncodeError> {
        Ok(())
    }

    fn decode<'a>(input: &'a [u8], memory_manager: &'a MemoryManager) -> Result<Self::Out<'a>, DecodeError> {
        Ok(())
    }

    fn out_size<'a>(key: &Self::In<'a>) -> usize { 0 }
}

impl<A: KeyFieldValue> ObjectKey for A {
    fn encode<'a, W: Writer>(key: &Self::In<'a>, writer: &mut W) -> Result<(), EncodeError> {
        A::encode_terminal(key, writer)
    }

    fn decode<'a>(input: &'a [u8], memory_manager: &'a MemoryManager) -> Result<Self::Out<'a>, DecodeError> {
        let mut reader = input;
        let a = A::decode_terminal(&mut reader, memory_manager)?;
        reader.is_done()?;
        Ok(a)
    }

    fn out_size<'a>(key: &Self::In<'a>) -> usize {
        A::out_size_terminal(key)
    }
}

impl<A: KeyFieldValue> ObjectKey for (A,) {
    fn encode<'a, W: Writer>(key: &Self::In<'a>, writer: &mut W) -> Result<(), EncodeError> {
        A::encode(&key.0, writer)
    }

    fn decode<'a>(input: &'a [u8], memory_manager: &'a MemoryManager) -> Result<Self::Out<'a>, DecodeError> {
        let mut reader = input;
        let a = A::decode(&mut reader, memory_manager)?;
        reader.is_done()?;
        Ok((a,))
    }

    fn out_size<'a>(key: &Self::In<'a>) -> usize {
        A::out_size_terminal(&key.0)
    }
}

impl<A: KeyFieldValue, B: KeyFieldValue> ObjectKey for (A, B) {
    fn encode<'a, W: Writer>(key: &Self::In<'a>, writer: &mut W) -> Result<(), EncodeError> {
        B::encode_terminal(&key.1, writer)?;
        A::encode(&key.0, writer)
    }

    fn decode<'a>(input: &'a [u8], memory_manager: &'a MemoryManager) -> Result<Self::Out<'a>, DecodeError> {
        let mut reader = input;
        let a = A::decode(&mut reader, memory_manager)?;
        let b = B::decode_terminal(&mut reader, memory_manager)?;
        reader.is_done()?;
        Ok((a, b))
    }

    fn out_size<'a>(key: &Self::In<'a>) -> usize {
        A::out_size(&key.0) + B::out_size_terminal(&key.1)
    }
}

impl<A: KeyFieldValue, B: KeyFieldValue, C: KeyFieldValue> ObjectKey for (A, B, C) {
    fn encode<'a, W: Writer>(key: &Self::In<'a>, writer: &mut W) -> Result<(), EncodeError> {
        C::encode_terminal(&key.2, writer)?;
        B::encode(&key.1, writer)?;
        A::encode(&key.0, writer)
    }

    fn decode<'a>(input: &'a [u8], memory_manager: &'a MemoryManager) -> Result<Self::Out<'a>, DecodeError> {
        let mut reader = input;
        let a = A::decode(&mut reader, memory_manager)?;
        let b = B::decode(&mut reader, memory_manager)?;
        let c = C::decode_terminal(&mut reader, memory_manager)?;
        reader.is_done()?;
        Ok((a, b, c))
    }

    fn out_size<'a>(key: &Self::In<'a>) -> usize {
        A::out_size(&key.0) + B::out_size(&key.1) + C::out_size_terminal(&key.2)
    }
}

impl<A: KeyFieldValue, B: KeyFieldValue, C: KeyFieldValue, D: KeyFieldValue> ObjectKey for (A, B, C, D) {
    fn encode<'a, W: Writer>(key: &Self::In<'a>, writer: &mut W) -> Result<(), EncodeError> {
        D::encode_terminal(&key.3, writer)?;
        C::encode(&key.2, writer)?;
        B::encode(&key.1, writer)?;
        A::encode(&key.0, writer)
    }

    fn decode<'a>(input: &'a [u8], memory_manager: &'a MemoryManager) -> Result<Self::Out<'a>, DecodeError> {
        let mut reader = input;
        let a = A::decode(&mut reader, memory_manager)?;
        let b = B::decode(&mut reader, memory_manager)?;
        let c = C::decode(&mut reader, memory_manager)?;
        let d = D::decode_terminal(&mut reader, memory_manager)?;
        reader.is_done()?;
        Ok((a, b, c, d))
    }

    fn out_size<'a>(key: &Self::In<'a>) -> usize {
        A::out_size(&key.0) + B::out_size(&key.1) + C::out_size(&key.2) + D::out_size_terminal(&key.3)
    }
}

