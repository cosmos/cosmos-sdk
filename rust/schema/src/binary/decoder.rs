use alloc::string::String;
use ixc_message_api::AccountID;
use crate::decoder::{decode, DecodeError};
use crate::list::ListDecodeVisitor;
use crate::mem::MemoryManager;
use crate::state_object::ObjectValue;
use crate::structs::{StructDecodeVisitor, StructType};
use crate::value::SchemaValue;

pub fn decode_value<'a, V: SchemaValue<'a>>(input: &'a [u8], memory_manager: &'a MemoryManager) -> Result<V, DecodeError> {
    let mut decoder = Decoder { buf: input, scope: memory_manager };
    decode(&mut decoder)
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
        let id = self.decode_u64()?;
        Ok(AccountID::new(id))
    }

    fn mem_manager(&self) -> &'a MemoryManager {
        &self.scope
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

    #[test]
    fn test_u32_decode() {
        let buf: [u8; 4] = [10, 0, 0, 0];
        let mut mem = MemoryManager::new();
        let x = decode_value::<u32>(&buf, &mut mem).unwrap();
        assert_eq!(x, 10);
    }

    #[test]
    fn test_decode_borrowed_string() {
        let str = "hello";
        let mut mem = MemoryManager::new();
        let x = decode_value::<&str>(str.as_bytes(), &mut mem).unwrap();
        assert_eq!(x, "hello");
    }

    #[test]
    fn test_decode_owned_string() {
        let str = "hello";
        let mut mem = MemoryManager::new();
        let x = decode_value::<alloc::string::String>(str.as_bytes(), &mut mem).unwrap();
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
        let res = encode_value(&coin, &mem as &dyn Allocator).unwrap();
        let decoded = decode_value::<Coin>(res, &mem).unwrap();
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
        let res = encode_value(&coins, &mem as &dyn Allocator).unwrap();
        let decoded = decode_value::<&[Coin]>(&res, &mem).unwrap();
        assert_eq!(decoded, coins);
    }
}
