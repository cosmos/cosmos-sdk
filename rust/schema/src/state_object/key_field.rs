use crate::buffer::{Reader, ReverseSliceWriter, Writer};
use crate::decoder::DecodeError;
use crate::encoder::EncodeError;
use crate::mem::MemoryManager;
use crate::state_object::value_field::ObjectFieldValue;

/// This trait is implemented for types that can be used as key fields in state objects.
pub trait KeyFieldValue: ObjectFieldValue {
    /// Encode the key segment as a non-terminal segment.
    fn encode<'a>(key: &Self::In<'a>, writer: &mut ReverseSliceWriter) -> Result<(), EncodeError> {
        unimplemented!("encode")
    }

    /// Encode the key segment as the terminal segment.
    fn encode_terminal<'a>(key: &Self::In<'a>, writer: &mut ReverseSliceWriter) -> Result<(), EncodeError> {
        Self::encode(key, writer)
    }

    /// Decode the key segment as a non-terminal segment.
    fn decode<'a>(reader: &mut &'a [u8], memory_manager: &'a MemoryManager) -> Result<Self::Out<'a>, DecodeError> {
        unimplemented!("decode")
    }

    /// Decode the key segment as the terminal segment.
    fn decode_terminal<'a>(reader: &mut &'a [u8], memory_manager: &'a MemoryManager) -> Result<Self::Out<'a>, DecodeError> {
        Self::decode(reader, memory_manager)
    }

    /// Get the size of the key segment as a non-terminal segment.
    fn out_size<'a>(key: &Self::In<'a>) -> usize;

    /// Get the size of the key segment as the terminal segment.
    fn out_size_terminal<'a>(key: &Self::In<'a>) -> usize {
        Self::out_size(key)
    }
}

impl KeyFieldValue for u8 {
    fn out_size<'a>(key: &Self::In<'a>) -> usize { 1 }
}

impl KeyFieldValue for u16 {
    fn out_size<'a>(key: &Self::In<'a>) -> usize { 2 }
}

impl KeyFieldValue for u32 {
    fn out_size<'a>(key: &Self::In<'a>) -> usize { 4 }
}

impl KeyFieldValue for u64 {
    fn out_size<'a>(key: &Self::In<'a>) -> usize { 8 }
    fn encode<'a>(key: &Self::In<'a>, writer: &mut ReverseSliceWriter) -> Result<(), EncodeError> {
        writer.write(&key.to_be_bytes())
    }

    fn decode<'a>(reader: &mut &'a [u8], memory_manager: &'a MemoryManager) -> Result<Self::Out<'a>, DecodeError> {
        let bz = reader.read_bytes(8)?;
        Ok(u64::from_be_bytes(bz.try_into().unwrap()))
    }
}

impl KeyFieldValue for u128 {
    fn out_size<'a>(key: &Self::In<'a>) -> usize { 16 }
}

impl KeyFieldValue for i8 {
    fn out_size<'a>(key: &Self::In<'a>) -> usize { 1 }
}

impl KeyFieldValue for i16 {
    fn out_size<'a>(key: &Self::In<'a>) -> usize { 2 }
}

impl KeyFieldValue for i32 {
    fn out_size<'a>(key: &Self::In<'a>) -> usize { 4 }
}

impl KeyFieldValue for i64 {
    fn out_size<'a>(key: &Self::In<'a>) -> usize { 8 }
}

impl KeyFieldValue for i128 {
    fn out_size<'a>(key: &Self::In<'a>) -> usize { 16 }
}

impl KeyFieldValue for bool {
    fn out_size<'a>(key: &Self::In<'a>) -> usize { 1 }
}

impl KeyFieldValue for simple_time::Time {
    fn out_size<'a>(key: &Self::In<'a>) -> usize { 12 }
}

impl KeyFieldValue for simple_time::Duration {
    fn out_size<'a>(key: &Self::In<'a>) -> usize { 12 }
}

impl KeyFieldValue for ixc_message_api::AccountID {
    fn out_size<'a>(key: &Self::In<'a>) -> usize { 8 }
    fn encode<'a>(key: &Self::In<'a>, writer: &mut ReverseSliceWriter) -> Result<(), EncodeError> {
        writer.write(&key.get().to_be_bytes())
    }
    fn decode<'a>(reader: &mut &'a [u8], memory_manager: &'a MemoryManager) -> Result<Self::Out<'a>, DecodeError> {
        let bz = reader.read_bytes(8)?;
        Ok(ixc_message_api::AccountID::new(u64::from_be_bytes(bz.try_into().unwrap())))
    }
}

impl KeyFieldValue for str {
    fn out_size<'a>(key: &Self::In<'a>) -> usize { key.len() + 4 }
    fn out_size_terminal<'a>(key: &Self::In<'a>) -> usize { key.len() }
}

impl KeyFieldValue for [u8] {
    fn out_size<'a>(key: &Self::In<'a>) -> usize { key.len() + 4 }
    fn out_size_terminal<'a>(key: &Self::In<'a>) -> usize { key.len() }
}
