use alloc::string::String;
use alloc::vec::Vec;
use ixc_message_api::AccountID;
use simple_time::{Duration, Time};
use crate::codec::ValueDecodeVisitor;
use crate::decoder::{DecodeError};
use crate::list::ListDecodeVisitor;
use crate::mem::MemoryManager;
use crate::state_object::ObjectValue;
use crate::structs::{StructDecodeVisitor, StructType};
use crate::value::SchemaValue;

pub fn decode_value<'a>(input: &'a [u8], memory_manager: &'a MemoryManager, visitor: &mut dyn ValueDecodeVisitor<'a>) -> Result<(), DecodeError> {
    visitor.decode(&mut Decoder { buf: input, scope: memory_manager })
}

pub(crate) struct Decoder<'a> {
    pub(crate) buf: &'a [u8],
    pub(crate) scope: &'a MemoryManager,
}

impl <'a> Decoder<'a> {
    fn read_bytes(&mut self, size: usize) -> Result<&'a [u8], DecodeError> {
        if self.buf.len() < size {
            return Err(DecodeError::OutOfData);
        }
        let bz = &self.buf[0..size];
        self.buf = &self.buf[size..];
        Ok(bz)
    }
}

impl<'a> crate::decoder::Decoder<'a> for Decoder<'a> {
    fn decode_u32(&mut self) -> Result<u32, DecodeError> {
        let bz = self.read_bytes(4)?;
        Ok(u32::from_le_bytes(bz.try_into().unwrap()))
    }

    fn decode_i32(&mut self) -> Result<i32, DecodeError> {
        let bz = self.read_bytes(4)?;
        Ok(i32::from_le_bytes(bz.try_into().unwrap()))
    }

    fn decode_u64(&mut self) -> Result<u64, DecodeError> {
        let bz = self.read_bytes(8)?;
        Ok(u64::from_le_bytes(bz.try_into().unwrap()))
    }

    fn decode_u128(&mut self) -> Result<u128, DecodeError> {
        let bz = self.read_bytes(16)?;
        Ok(u128::from_le_bytes(bz.try_into().unwrap()))
    }

    fn decode_borrowed_str(&mut self) -> Result<&'a str, DecodeError> {
        let bz = self.buf;
        self.buf = &[];
        Ok(core::str::from_utf8(bz).map_err(|_| DecodeError::InvalidData)?)
    }

    fn decode_owned_str(&mut self) -> Result<String, DecodeError> {
        let bz = self.buf;
        self.buf = &[];
        Ok(String::from_utf8(bz.to_vec()).map_err(|_| DecodeError::InvalidData)?)
    }

    fn decode_struct(&mut self, visitor: &mut dyn StructDecodeVisitor<'a>, struct_type: &StructType) -> Result<(), DecodeError> {
        let mut i = 0;
        let mut sub = Decoder { buf: self.buf, scope: self.scope };
        let mut inner = InnerDecoder { outer: &mut sub };
        for _ in struct_type.fields.iter() {
            visitor.decode_field(i, &mut inner)?;
            i += 1;
        }
        Ok(())
    }

    fn decode_list(&mut self, visitor: &mut dyn ListDecodeVisitor<'a>) -> Result<(), DecodeError> {
        let size = self.decode_u32()? as usize;
        visitor.init(size, &self.scope)?;
        let mut sub = Decoder { buf: self.buf, scope: self.scope };
        let mut inner = InnerDecoder { outer: &mut sub };
        for _ in 0..size {
            visitor.next(&mut inner)?;
        }
        Ok(())
    }

    fn decode_account_id(&mut self) -> Result<AccountID, DecodeError> {
        let id = self.decode_u128()?;
        Ok(AccountID::new(id))
    }

    fn mem_manager(&self) -> &'a MemoryManager {
        &self.scope
    }

    fn decode_bool(&mut self) -> Result<bool, DecodeError> {
        let bz = self.read_bytes(1)?;
        Ok(bz[0] != 0)
    }

    fn decode_u8(&mut self) -> Result<u8, DecodeError> {
        let bz = self.read_bytes(1)?;
        Ok(bz[0])
    }

    fn decode_u16(&mut self) -> Result<u16, DecodeError> {
        let bz = self.read_bytes(2)?;
        Ok(u16::from_le_bytes(bz.try_into().unwrap()))
    }

    fn decode_i8(&mut self) -> Result<i8, DecodeError> {
        let bz = self.read_bytes(1)?;
        Ok(bz[0] as i8)
    }

    fn decode_i64(&mut self) -> Result<i64, DecodeError> {
        let bz = self.read_bytes(8)?;
        Ok(i64::from_le_bytes(bz.try_into().unwrap()))
    }

    fn decode_i128(&mut self) -> Result<i128, DecodeError> {
        let bz = self.read_bytes(16)?;
        Ok(i128::from_le_bytes(bz.try_into().unwrap()))
    }

    fn decode_borrowed_bytes(&mut self) -> Result<&'a [u8], DecodeError> {
        let bz = self.buf;
        self.buf = &[];
        Ok(bz)
    }

    fn decode_owned_bytes(&mut self) -> Result<Vec<u8>, DecodeError> {
        let bz = self.buf;
        self.buf = &[];
        Ok(bz.to_vec())
    }

    fn decode_time(&mut self) -> Result<Time, DecodeError> {
        Ok(Time::from_unix_nanos(self.decode_i128()?))
    }

    fn decode_duration(&mut self) -> Result<Duration, DecodeError> {
        Ok(Duration::from_nanos(self.decode_i128()?))
    }

    fn decode_i16(&mut self) -> Result<i16, DecodeError> {
        let bz = self.read_bytes(2)?;
        Ok(i16::from_le_bytes(bz.try_into().unwrap()))
    }

    fn decode_option(&mut self, visitor: &mut dyn ValueDecodeVisitor<'a>) -> Result<bool, DecodeError> {
        if self.buf.is_empty() {
            Ok(false)
        } else {
            visitor.decode(&mut Decoder { buf: self.buf, scope: self.scope })?;
            Ok(true)
        }
    }
}

struct InnerDecoder<'b, 'a: 'b> {
    outer: &'b mut Decoder<'a>,
}
impl<'b, 'a: 'b> crate::decoder::Decoder<'a> for InnerDecoder<'b, 'a> {
    fn decode_u32(&mut self) -> Result<u32, DecodeError> {
        self.outer.decode_u32()
    }

    fn decode_i32(&mut self) -> Result<i32, DecodeError> { self.outer.decode_i32() }

    fn decode_u64(&mut self) -> Result<u64, DecodeError> { self.outer.decode_u64() }

    fn decode_u128(&mut self) -> Result<u128, DecodeError> {
        self.outer.decode_u128()
    }

    fn decode_borrowed_str(&mut self) -> Result<&'a str, DecodeError> {
        let size = self.decode_u32()? as usize;
        let bz = self.outer.read_bytes(size)?;
        Ok(core::str::from_utf8(bz).map_err(|_| DecodeError::InvalidData)?)
    }

    fn decode_owned_str(&mut self) -> Result<String, DecodeError> {
        let size = self.decode_u32()? as usize;
        let bz = self.outer.read_bytes(size)?;
        Ok(String::from_utf8(bz.to_vec()).map_err(|_| DecodeError::InvalidData)?)
    }

    fn decode_struct(&mut self, visitor: &mut dyn StructDecodeVisitor<'a>, struct_type: &StructType) -> Result<(), DecodeError> {
        let size = self.decode_u32()? as usize;
        let bz = self.outer.read_bytes(size)?;
        let mut sub = Decoder { buf: bz, scope: self.outer.scope };
        sub.decode_struct(visitor, struct_type)
    }

    fn decode_list(&mut self, visitor: &mut dyn ListDecodeVisitor<'a>) -> Result<(), DecodeError> {
        let size = self.decode_u32()? as usize;
        let bz = self.outer.read_bytes(size)?;
        let mut sub = Decoder { buf: bz, scope: self.outer.scope };
        sub.decode_list(visitor)
    }

    fn decode_account_id(&mut self) -> Result<AccountID, DecodeError> {
        self.outer.decode_account_id()
    }

    fn mem_manager(&self) -> &'a MemoryManager {
        &self.outer.scope
    }

    fn decode_bool(&mut self) -> Result<bool, DecodeError> {
        self.outer.decode_bool()
    }

    fn decode_u8(&mut self) -> Result<u8, DecodeError> {
        self.outer.decode_u8()
    }

    fn decode_u16(&mut self) -> Result<u16, DecodeError> {
        self.outer.decode_u16()
    }

    fn decode_i8(&mut self) -> Result<i8, DecodeError> {
        self.outer.decode_i8()
    }

    fn decode_i64(&mut self) -> Result<i64, DecodeError> {
        self.outer.decode_i64()
    }

    fn decode_i128(&mut self) -> Result<i128, DecodeError> {
        self.outer.decode_i128()
    }

    fn decode_borrowed_bytes(&mut self) -> Result<&'a [u8], DecodeError> {
        let size = self.decode_u32()? as usize;
        let bz = self.outer.read_bytes(size)?;
        Ok(bz)
    }

    fn decode_owned_bytes(&mut self) -> Result<Vec<u8>, DecodeError> {
        let size = self.decode_u32()? as usize;
        let bz = self.outer.read_bytes(size)?;
        Ok(bz.to_vec())
    }

    fn decode_time(&mut self) -> Result<Time, DecodeError> {
        self.outer.decode_time()
    }

    fn decode_duration(&mut self) -> Result<Duration, DecodeError> {
        self.outer.decode_duration()
    }

    fn decode_i16(&mut self) -> Result<i16, DecodeError> {
        self.outer.decode_i16()
    }

    fn decode_option(&mut self, visitor: &mut dyn ValueDecodeVisitor<'a>) -> Result<bool, DecodeError> {
        let present = self.decode_bool()?;
        if present {
            visitor.decode(self)?;
            Ok(true)
        } else {
            Ok(false)
        }
    }
}

#[cfg(test)]
mod tests {
    extern crate std;

    use alloc::vec;
    use allocator_api2::alloc::Allocator;
    use crate::binary::decoder::decode_value;
    use crate::binary::encoder::encode_value;
    use crate::decoder::{DecodeError, Decoder};
    use crate::encoder::{EncodeError, Encoder};
    use crate::field::Field;
    use crate::mem::MemoryManager;
    use crate::schema::SchemaType;
    use crate::state_object::ObjectFieldValue;
    use crate::structs::{StructDecodeVisitor, StructEncodeVisitor, StructSchema, StructType};
    use crate::types::{to_field, ReferenceableType, StrT, StructT, UIntNT};
    use crate::value::{ListElementValue, SchemaValue};

    extern crate ixc_schema_macros;
    use ixc_schema_macros::*;
    use crate::binary::NativeBinaryCodec;

    #[test]
    fn test_u32_decode() {
        let buf: [u8; 4] = [10, 0, 0, 0];
        let mut mem = MemoryManager::new();
        let x = crate::codec::decode_value::<u32>(&NativeBinaryCodec, &buf, &mem).unwrap();
        assert_eq!(x, 10);
    }

    #[test]
    fn test_decode_borrowed_string() {
        let str = "hello";
        let mut mem = MemoryManager::new();
        let x = crate::codec::decode_value::<&str>(&NativeBinaryCodec, str.as_bytes(), &mem).unwrap();
        assert_eq!(x, "hello");
    }

    #[test]
    fn test_decode_owned_string() {
        let str = "hello";
        let mut mem = MemoryManager::new();
        let x = crate::codec::decode_value::<alloc::string::String>(&NativeBinaryCodec, str.as_bytes(), &mem).unwrap();
        assert_eq!(x, "hello");
    }

    #[derive(Debug, PartialEq, SchemaValue)]
    #[sealed]
    struct Coin<'b> {
        denom: &'b str,
        amount: u128,
    }

    impl <'a> Drop for Coin<'a> {
        fn drop(&mut self) {
            std::println!("drop Coin");
        }
    }

    #[test]
    fn test_coin() {
        let coin = Coin {
            denom: "uatom",
            amount: 1234567890,
        };
        let mem = MemoryManager::new();
        let res = encode_value(&coin, &mem).unwrap();
        let decoded = crate::codec::decode_value::<Coin>(&NativeBinaryCodec, res, &mem).unwrap();
        assert_eq!(decoded, coin);
    }

    #[test]
    fn test_coins() {
        let coins = vec![Coin {
            denom: "uatom",
            amount: 1234567890,
        }, Coin {
            denom: "foo",
            amount: 9876543210,
        }];
        let mem = MemoryManager::new();
        let res = encode_value(&coins, &mem).unwrap();
        let decoded = crate::codec::decode_value::<&[Coin]>(&NativeBinaryCodec, &res, &mem).unwrap();
        assert_eq!(decoded, coins);
    }
}
