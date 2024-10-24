use crate::buffer::{Reader, ReverseSliceWriter, Writer};
use crate::decoder::DecodeError;
use crate::encoder::EncodeError;
use crate::mem::MemoryManager;
use crate::state_object::value_field::{Bytes, ObjectFieldValue};
use crate::Str;

/// This trait is implemented for types that can be used as key fields in state objects.
pub trait KeyFieldValue: ObjectFieldValue {
    /// Encode the key segment as a non-terminal segment.
    fn encode<'a>(key: &Self::In<'a>, writer: &mut ReverseSliceWriter) -> Result<(), EncodeError>;

    /// Encode the key segment as the terminal segment.
    fn encode_terminal<'a>(key: &Self::In<'a>, writer: &mut ReverseSliceWriter) -> Result<(), EncodeError> {
        Self::encode(key, writer)
    }

    /// Decode the key segment as a non-terminal segment.
    fn decode<'a>(reader: &mut &'a [u8], memory_manager: &'a MemoryManager) -> Result<Self::Out<'a>, DecodeError>;

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
    fn encode<'a>(key: &Self::In<'a>, writer: &mut ReverseSliceWriter) -> Result<(), EncodeError> {
        writer.write(&[*key])
    }

    fn decode<'a>(reader: &mut &'a [u8], memory_manager: &'a MemoryManager) -> Result<Self::Out<'a>, DecodeError> {
        Ok(reader.read_bytes(1)?[0])
    }

    fn out_size<'a>(key: &Self::In<'a>) -> usize { 1 }
}

impl KeyFieldValue for u16 {
    fn encode<'a>(key: &Self::In<'a>, writer: &mut ReverseSliceWriter) -> Result<(), EncodeError> {
        writer.write(&key.to_be_bytes())
    }

    fn decode<'a>(reader: &mut &'a [u8], memory_manager: &'a MemoryManager) -> Result<Self::Out<'a>, DecodeError> {
        let bz = reader.read_bytes(2)?;
        Ok(u16::from_be_bytes(bz.try_into().unwrap()))
    }

    fn out_size<'a>(key: &Self::In<'a>) -> usize { 2 }
}

impl KeyFieldValue for u32 {
    fn encode<'a>(key: &Self::In<'a>, writer: &mut ReverseSliceWriter) -> Result<(), EncodeError> {
        writer.write(&key.to_be_bytes())
    }

    fn decode<'a>(reader: &mut &'a [u8], memory_manager: &'a MemoryManager) -> Result<Self::Out<'a>, DecodeError> {
        let bz = reader.read_bytes(4)?;
        Ok(u32::from_be_bytes(bz.try_into().unwrap()))
    }

    fn out_size<'a>(key: &Self::In<'a>) -> usize { 4 }
}

impl KeyFieldValue for u64 {
    fn encode<'a>(key: &Self::In<'a>, writer: &mut ReverseSliceWriter) -> Result<(), EncodeError> {
        writer.write(&key.to_be_bytes())
    }

    fn decode<'a>(reader: &mut &'a [u8], memory_manager: &'a MemoryManager) -> Result<Self::Out<'a>, DecodeError> {
        let bz = reader.read_bytes(8)?;
        Ok(u64::from_be_bytes(bz.try_into().unwrap()))
    }

    fn out_size<'a>(key: &Self::In<'a>) -> usize { 8 }
}

impl KeyFieldValue for u128 {
    fn encode<'a>(key: &Self::In<'a>, writer: &mut ReverseSliceWriter) -> Result<(), EncodeError> {
        writer.write(&key.to_be_bytes())
    }

    fn decode<'a>(reader: &mut &'a [u8], memory_manager: &'a MemoryManager) -> Result<Self::Out<'a>, DecodeError> {
        let bz = reader.read_bytes(16)?;
        Ok(u128::from_be_bytes(bz.try_into().unwrap()))
    }

    fn out_size<'a>(key: &Self::In<'a>) -> usize { 16 }
}

impl KeyFieldValue for i8 {
    fn encode<'a>(key: &Self::In<'a>, writer: &mut ReverseSliceWriter) -> Result<(), EncodeError> {
        let x = *key as u8;
        // flip first bit for ordering
        writer.write(&[x ^ 0x80])
    }

    fn decode<'a>(reader: &mut &'a [u8], memory_manager: &'a MemoryManager) -> Result<Self::Out<'a>, DecodeError> {
        let x = reader.read_bytes(1)?[0];
        // flip first bit back
        Ok((x ^ 0x80) as i8)
    }

    fn out_size<'a>(key: &Self::In<'a>) -> usize { 1 }
}

impl KeyFieldValue for i16 {
    fn encode<'a>(key: &Self::In<'a>, writer: &mut ReverseSliceWriter) -> Result<(), EncodeError> {
        // flip first bit for ordering
        let x = *key as u16 ^ 0x8000;
        writer.write(&x.to_be_bytes())
    }

    fn decode<'a>(reader: &mut &'a [u8], memory_manager: &'a MemoryManager) -> Result<Self::Out<'a>, DecodeError> {
        let x = u16::from_be_bytes(reader.read_bytes(2)?.try_into().unwrap());
        // flip first bit back
        Ok((x ^ 0x8000) as i16)
    }

    fn out_size<'a>(key: &Self::In<'a>) -> usize { 2 }
}

impl KeyFieldValue for i32 {
    fn encode<'a>(key: &Self::In<'a>, writer: &mut ReverseSliceWriter) -> Result<(), EncodeError> {
        let x = *key as u32 ^ 0x80000000;
        writer.write(&x.to_be_bytes())
    }

    fn decode<'a>(reader: &mut &'a [u8], memory_manager: &'a MemoryManager) -> Result<Self::Out<'a>, DecodeError> {
        let x = u32::from_be_bytes(reader.read_bytes(4)?.try_into().unwrap());
        Ok((x ^ 0x80000000) as i32)
    }


    fn out_size<'a>(key: &Self::In<'a>) -> usize { 4 }
}

impl KeyFieldValue for i64 {
    fn encode<'a>(key: &Self::In<'a>, writer: &mut ReverseSliceWriter) -> Result<(), EncodeError> {
        let x = *key as u64 ^ 0x8000000000000000;
        writer.write(&x.to_be_bytes())
    }

    fn decode<'a>(reader: &mut &'a [u8], memory_manager: &'a MemoryManager) -> Result<Self::Out<'a>, DecodeError> {
        let x = u64::from_be_bytes(reader.read_bytes(8)?.try_into().unwrap());
        Ok((x ^ 0x8000000000000000) as i64)
    }

    fn out_size<'a>(key: &Self::In<'a>) -> usize { 8 }
}

impl KeyFieldValue for i128 {
    fn encode<'a>(key: &Self::In<'a>, writer: &mut ReverseSliceWriter) -> Result<(), EncodeError> {
        let x = *key as u128 ^ 0x80000000000000000000000000000000;
        writer.write(&x.to_be_bytes())
    }

    fn decode<'a>(reader: &mut &'a [u8], memory_manager: &'a MemoryManager) -> Result<Self::Out<'a>, DecodeError> {
        let x = u128::from_be_bytes(reader.read_bytes(16)?.try_into().unwrap());
        Ok((x ^ 0x80000000000000000000000000000000) as i128)
    }

    fn out_size<'a>(key: &Self::In<'a>) -> usize { 16 }
}

impl KeyFieldValue for bool {
    fn encode<'a>(key: &Self::In<'a>, writer: &mut ReverseSliceWriter) -> Result<(), EncodeError> {
        writer.write(&[*key as u8])
    }

    fn decode<'a>(reader: &mut &'a [u8], memory_manager: &'a MemoryManager) -> Result<Self::Out<'a>, DecodeError> {
        Ok(reader.read_bytes(1)?[0] != 0)
    }

    fn out_size<'a>(key: &Self::In<'a>) -> usize { 1 }
}

impl KeyFieldValue for simple_time::Time {
    fn encode<'a>(key: &Self::In<'a>, writer: &mut ReverseSliceWriter) -> Result<(), EncodeError> {
        // TODO we only need 12 bytes max
        <i128 as KeyFieldValue>::encode(&key.unix_nanos(), writer)
    }

    fn decode<'a>(reader: &mut &'a [u8], memory_manager: &'a MemoryManager) -> Result<Self::Out<'a>, DecodeError> {
        <i128 as KeyFieldValue>::decode(reader, memory_manager)
            .map(simple_time::Time::from_unix_nanos)
    }

    fn out_size<'a>(key: &Self::In<'a>) -> usize { 12 }
}

impl KeyFieldValue for simple_time::Duration {
    fn encode<'a>(key: &Self::In<'a>, writer: &mut ReverseSliceWriter) -> Result<(), EncodeError> {
        // TODO we only need 12 bytes max
        <i128 as KeyFieldValue>::encode(&key.nanos(), writer)
    }

    fn decode<'a>(reader: &mut &'a [u8], memory_manager: &'a MemoryManager) -> Result<Self::Out<'a>, DecodeError> {
        <i128 as KeyFieldValue>::decode(reader, memory_manager)
            .map(simple_time::Duration::from_nanos)
    }

    fn out_size<'a>(key: &Self::In<'a>) -> usize { 12 }
}

impl KeyFieldValue for ixc_message_api::AccountID {
    fn encode<'a>(key: &Self::In<'a>, writer: &mut ReverseSliceWriter) -> Result<(), EncodeError> {
        let id: u128 = (*key).into();
        writer.write(&id.to_be_bytes())
    }

    fn decode<'a>(reader: &mut &'a [u8], _memory_manager: &'a MemoryManager) -> Result<Self::Out<'a>, DecodeError> {
        let bz = reader.read_bytes(8)?;
        Ok(ixc_message_api::AccountID::new(u128::from_be_bytes(bz.try_into().unwrap())))
    }

    fn out_size<'a>(_key: &Self::In<'a>) -> usize { 16 }
}

impl KeyFieldValue for Str {
    fn encode<'a>(key: &Self::In<'a>, writer: &mut ReverseSliceWriter) -> Result<(), EncodeError> {
        // write null terminator
        writer.write(&[0]);
        writer.write(key.as_bytes())
    }

    fn encode_terminal<'a>(key: &Self::In<'a>, writer: &mut ReverseSliceWriter) -> Result<(), EncodeError> {
        // no null terminator needed
        writer.write(key.as_bytes())
    }

    fn decode<'a>(reader: &mut &'a [u8], memory_manager: &'a MemoryManager) -> Result<Self::Out<'a>, DecodeError> {
        let mut i = 0;
        while reader[i] != 0 {
            i += 1;
        }
        let s = core::str::from_utf8(&reader[..i])
            .map_err(|_| DecodeError::InvalidUtf8)?;
        *reader = &reader[i + 1..];
        Ok(s)
    }

    fn decode_terminal<'a>(reader: &mut &'a [u8], memory_manager: &'a MemoryManager) -> Result<Self::Out<'a>, DecodeError> {
        let s = core::str::from_utf8(reader)
            .map_err(|_| DecodeError::InvalidUtf8)?;
        *reader = &reader[s.len()..];
        Ok(s)
    }

    fn out_size<'a>(key: &Self::In<'a>) -> usize { key.len() + 1 }

    fn out_size_terminal<'a>(key: &Self::In<'a>) -> usize { key.len() }
}

impl KeyFieldValue for Bytes {
    fn encode<'a>(key: &Self::In<'a>, writer: &mut ReverseSliceWriter) -> Result<(), EncodeError> {
        writer.write(key)?;
        // write length
        writer.write(&(key.len() as u32).to_be_bytes())
    }

    fn encode_terminal<'a>(key: &Self::In<'a>, writer: &mut ReverseSliceWriter) -> Result<(), EncodeError> {
        writer.write(key)
    }

    fn decode<'a>(reader: &mut &'a [u8], memory_manager: &'a MemoryManager) -> Result<Self::Out<'a>, DecodeError> {
        let len = u32::from_be_bytes(reader.read_bytes(4)?.try_into().unwrap()) as usize;
        let key = reader.read_bytes(len)?;
        Ok(key)
    }

    fn decode_terminal<'a>(reader: &mut &'a [u8], memory_manager: &'a MemoryManager) -> Result<Self::Out<'a>, DecodeError> {
        Ok(reader)
    }

    fn out_size<'a>(key: &Self::In<'a>) -> usize { key.len() + 4 }
    fn out_size_terminal<'a>(key: &Self::In<'a>) -> usize { key.len() }
}
