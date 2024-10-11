use ixc_message_api::AccountID;
use simple_time::{Duration, Time};
use crate::encoder::{EncodeError};
use crate::structs::{StructDecodeVisitor, StructEncodeVisitor, StructType};
use crate::value::SchemaValue;
use crate::buffer::{Writer, WriterFactory};
use crate::codec::ValueEncodeVisitor;
use crate::list::ListEncodeVisitor;
use crate::enums::EnumType;
use crate::state_object::ObjectValue;

pub fn encode_value<'a>(value: &dyn ValueEncodeVisitor, writer_factory: &'a dyn WriterFactory) -> Result<&'a [u8], EncodeError> {
    let mut sizer = EncodeSizer { size: 0 };
    value.encode(&mut sizer)?;
    let mut writer = writer_factory.new_reverse(sizer.size)?;
    let mut encoder = Encoder { writer: &mut writer };
    value.encode(&mut encoder)?;
    Ok(writer.finish())
}

pub(crate) struct Encoder<'a, W> {
    pub(crate) writer: &'a mut W,
}

impl<'a, W: Writer> crate::encoder::Encoder for Encoder<'a, W> {
    fn encode_u32(&mut self, x: u32) -> Result<(), EncodeError> {
        self.writer.write(&x.to_le_bytes())
    }

    fn encode_i32(&mut self, x: i32) -> Result<(), EncodeError> {
        self.writer.write(&x.to_le_bytes())
    }

    fn encode_u64(&mut self, x: u64) -> Result<(), EncodeError> {
        self.writer.write(&x.to_le_bytes())
    }

    fn encode_u128(&mut self, x: u128) -> Result<(), EncodeError> {
        self.writer.write(&x.to_le_bytes())
    }

    fn encode_str(&mut self, x: &str) -> Result<(), EncodeError> {
        self.writer.write(x.as_bytes())
    }

    fn encode_list(&mut self, visitor: &dyn ListEncodeVisitor) -> Result<(), EncodeError> {
        let mut sub = Encoder { writer: self.writer };
        let mut inner = InnerEncoder::<W> { outer: &mut sub };
        let size = visitor.encode_reverse(&mut inner)?;
        self.encode_u32(size)?;
        Ok(())
    }

    fn encode_struct(&mut self, visitor: &dyn StructEncodeVisitor, struct_type: &StructType) -> Result<(), EncodeError> {
        let mut i = struct_type.fields.len();
        let mut sub = Encoder { writer: self.writer };
        let mut inner = InnerEncoder::<W> { outer: &mut sub };
        for f in struct_type.fields.iter().rev() {
            i -= 1;
            visitor.encode_field(i, &mut inner)?;
        }
        Ok(())
    }

    fn encode_account_id(&mut self, x: AccountID) -> Result<(), EncodeError> {
        self.encode_u64(x.into())
    }

    fn encode_bool(&mut self, x: bool) -> Result<(), EncodeError> {
        self.encode_u8(if x { 1 } else { 0 })
    }

    fn encode_u8(&mut self, x: u8) -> Result<(), EncodeError> {
        self.writer.write(&[x])
    }

    fn encode_u16(&mut self, x: u16) -> Result<(), EncodeError> {
        self.writer.write(&x.to_le_bytes())
    }

    fn encode_i8(&mut self, x: i8) -> Result<(), EncodeError> {
        self.writer.write(&[x as u8])
    }

    fn encode_i16(&mut self, x: i16) -> Result<(), EncodeError> {
        self.writer.write(&x.to_le_bytes())
    }

    fn encode_i64(&mut self, x: i64) -> Result<(), EncodeError> {
        self.writer.write(&x.to_le_bytes())
    }

    fn encode_i128(&mut self, x: i128) -> Result<(), EncodeError> {
        self.writer.write(&x.to_le_bytes())
    }

    fn encode_bytes(&mut self, x: &[u8]) -> Result<(), EncodeError> {
        self.writer.write(x)
    }

    fn encode_time(&mut self, x: Time) -> Result<(), EncodeError> {
        /// TODO find a more efficient way to encode time
        self.encode_i128(x.unix_nanos())
    }

    fn encode_duration(&mut self, x: Duration) -> Result<(), EncodeError> {
        /// TODO find a more efficient way to encode duration
        self.encode_i128(x.nanos())
    }

    fn encode_option(&mut self, visitor: Option<&dyn ValueEncodeVisitor>) -> Result<(), EncodeError> {
        if let Some(visitor) = visitor {
            visitor.encode(self)
        } else {
            Ok(())
        }
    }
}

pub(crate) struct EncodeSizer {
    pub(crate) size: usize,
}

impl crate::encoder::Encoder for EncodeSizer {
    fn encode_u32(&mut self, x: u32) -> Result<(), EncodeError> {
        self.size += 4;
        Ok(())
    }

    fn encode_i32(&mut self, x: i32) -> Result<(), EncodeError> {
        self.size += 4;
        Ok(())
    }

    fn encode_u64(&mut self, x: u64) -> Result<(), EncodeError> {
        self.size += 8;
        Ok(())
    }

    fn encode_u128(&mut self, x: u128) -> Result<(), EncodeError> {
        self.size += 16;
        Ok(())
    }

    fn encode_str(&mut self, x: &str) -> Result<(), EncodeError> {
        self.size += x.len();
        Ok(())
    }

    fn encode_list(&mut self, visitor: &dyn ListEncodeVisitor) -> Result<(), EncodeError> {
        self.size += 4;
        let mut sub = InnerEncodeSizer { outer: self };
        visitor.encode_reverse(&mut sub)?;
        Ok(())
    }

    fn encode_struct(&mut self, visitor: &dyn StructEncodeVisitor, struct_type: &StructType) -> Result<(), EncodeError> {
        let mut i = 0;
        let mut sub = InnerEncodeSizer { outer: self };
        for f in struct_type.fields {
            visitor.encode_field(i, &mut sub)?;
            i += 1;
        }
        Ok(())
    }

    fn encode_account_id(&mut self, x: AccountID) -> Result<(), EncodeError> {
        self.size += 8;
        Ok(())
    }

    fn encode_enum(&mut self, x: i32, enum_type: &EnumType) -> Result<(), EncodeError> {
        self.encode_i32(x)
    }

    fn encode_bool(&mut self, x: bool) -> Result<(), EncodeError> {
        self.size += 1;
        Ok(())
    }

    fn encode_u8(&mut self, x: u8) -> Result<(), EncodeError> {
        self.size += 1;
        Ok(())
    }

    fn encode_u16(&mut self, x: u16) -> Result<(), EncodeError> {
        self.size += 2;
        Ok(())
    }

    fn encode_i8(&mut self, x: i8) -> Result<(), EncodeError> {
        self.size += 1;
        Ok(())
    }

    fn encode_i16(&mut self, x: i16) -> Result<(), EncodeError> {
        self.size += 2;
        Ok(())
    }

    fn encode_i64(&mut self, x: i64) -> Result<(), EncodeError> {
        self.size += 8;
        Ok(())
    }

    fn encode_i128(&mut self, x: i128) -> Result<(), EncodeError> {
        self.size += 16;
        Ok(())
    }

    fn encode_bytes(&mut self, x: &[u8]) -> Result<(), EncodeError> {
        self.size += x.len();
        Ok(())
    }

    fn encode_time(&mut self, x: Time) -> Result<(), EncodeError> {
        self.encode_i128(x.unix_nanos())
    }

    fn encode_duration(&mut self, x: Duration) -> Result<(), EncodeError> {
        self.encode_i128(x.nanos())
    }

    fn encode_option(&mut self, visitor: Option<&dyn ValueEncodeVisitor>) -> Result<(), EncodeError> {
        if let Some(visitor) = visitor {
            visitor.encode(self)
        } else {
            Ok(())
        }
    }
}

pub(crate) struct InnerEncoder<'b, 'a: 'b, W> {
    pub(crate) outer: &'b mut Encoder<'a, W>,
}

impl<'b, 'a: 'b, W: Writer> crate::encoder::Encoder for InnerEncoder<'a, 'b, W> {
    fn encode_u32(&mut self, x: u32) -> Result<(), EncodeError> {
        self.outer.encode_u32(x)
    }

    fn encode_i32(&mut self, x: i32) -> Result<(), EncodeError> { self.outer.encode_i32(x) }

    fn encode_u64(&mut self, x: u64) -> Result<(), EncodeError> { self.outer.encode_u64(x) }

    fn encode_u128(&mut self, x: u128) -> Result<(), EncodeError> {
        self.outer.encode_u128(x)
    }

    fn encode_str(&mut self, x: &str) -> Result<(), EncodeError> {
        self.outer.encode_str(x)?;
        self.encode_u32(x.len() as u32)
    }

    fn encode_list(&mut self, visitor: &dyn ListEncodeVisitor) -> Result<(), EncodeError> {
        let end_pos = self.outer.writer.pos(); // this is a reverse writer so we start at the end
        self.outer.encode_list(visitor)?;
        let start_pos = self.outer.writer.pos(); // now we know the start position
        let size = (end_pos - start_pos) as u32;
        self.outer.encode_u32(size)
    }

    fn encode_struct(&mut self, visitor: &dyn StructEncodeVisitor, struct_type: &StructType) -> Result<(), EncodeError> {
        let end_pos = self.outer.writer.pos(); // this is a reverse writer so we start at the end
        self.outer.encode_struct(visitor, struct_type)?;
        let start_pos = self.outer.writer.pos(); // now we know the start position
        let size = (end_pos - start_pos) as u32;
        self.outer.encode_u32(size)
    }

    fn encode_account_id(&mut self, x: AccountID) -> Result<(), EncodeError> {
        self.outer.encode_account_id(x)
    }

    fn encode_bool(&mut self, x: bool) -> Result<(), EncodeError> {
        self.outer.encode_bool(x)
    }

    fn encode_u8(&mut self, x: u8) -> Result<(), EncodeError> {
        self.outer.encode_u8(x)
    }

    fn encode_u16(&mut self, x: u16) -> Result<(), EncodeError> {
        self.outer.encode_u16(x)
    }

    fn encode_i8(&mut self, x: i8) -> Result<(), EncodeError> {
        self.outer.encode_i8(x)
    }

    fn encode_i16(&mut self, x: i16) -> Result<(), EncodeError> {
        self.outer.encode_i16(x)
    }

    fn encode_i64(&mut self, x: i64) -> Result<(), EncodeError> {
        self.outer.encode_i64(x)
    }

    fn encode_i128(&mut self, x: i128) -> Result<(), EncodeError> {
        self.outer.encode_i128(x)
    }

    fn encode_bytes(&mut self, x: &[u8]) -> Result<(), EncodeError> {
        self.outer.encode_bytes(x)?;
        self.encode_u32(x.len() as u32)
    }

    fn encode_time(&mut self, x: Time) -> Result<(), EncodeError> {
        self.encode_i128(x.unix_nanos())
    }

    fn encode_duration(&mut self, x: Duration) -> Result<(), EncodeError> {
        self.encode_i128(x.nanos())
    }

    fn encode_option(&mut self, visitor: Option<&dyn ValueEncodeVisitor>) -> Result<(), EncodeError> {
        if let Some(visitor) = visitor {
            visitor.encode(self)?;
            self.encode_bool(true)?;
        } else {
            self.encode_bool(false)?;
        }
        Ok(())
    }
}

pub(crate) struct InnerEncodeSizer<'a> {
    pub(crate) outer: &'a mut EncodeSizer,
}

impl<'a> crate::encoder::Encoder for InnerEncodeSizer<'a> {
    fn encode_u32(&mut self, x: u32) -> Result<(), EncodeError> {
        self.outer.size += 4;
        Ok(())
    }

    fn encode_i32(&mut self, x: i32) -> Result<(), EncodeError> {
        self.outer.size += 4;
        Ok(())
    }

    fn encode_u64(&mut self, x: u64) -> Result<(), EncodeError> {
        self.outer.size += 8;
        Ok(())
    }

    fn encode_u128(&mut self, x: u128) -> Result<(), EncodeError> {
        self.outer.encode_u128(x)
    }

    fn encode_str(&mut self, x: &str) -> Result<(), EncodeError> {
        self.outer.size += 4;
        self.outer.encode_str(x)
    }

    fn encode_list(&mut self, visitor: &dyn ListEncodeVisitor) -> Result<(), EncodeError> {
        self.outer.size += 4; // for the for bytes size
        self.outer.encode_list(visitor)
    }

    fn encode_struct(&mut self, visitor: &dyn StructEncodeVisitor, struct_type: &StructType) -> Result<(), EncodeError> {
        self.outer.size += 4;
        self.outer.encode_struct(visitor, struct_type)
    }

    fn encode_account_id(&mut self, x: AccountID) -> Result<(), EncodeError> {
        self.outer.size += 8;
        Ok(())
    }

    fn encode_bool(&mut self, x: bool) -> Result<(), EncodeError> {
        self.outer.size += 1;
        Ok(())
    }

    fn encode_u8(&mut self, x: u8) -> Result<(), EncodeError> {
        self.outer.size += 1;
        Ok(())
    }

    fn encode_u16(&mut self, x: u16) -> Result<(), EncodeError> {
        self.outer.size += 2;
        Ok(())
    }

    fn encode_i8(&mut self, x: i8) -> Result<(), EncodeError> {
        self.outer.size += 1;
        Ok(())
    }

    fn encode_i16(&mut self, x: i16) -> Result<(), EncodeError> {
        self.outer.size += 2;
        Ok(())
    }

    fn encode_i64(&mut self, x: i64) -> Result<(), EncodeError> {
        self.outer.size += 8;
        Ok(())
    }

    fn encode_i128(&mut self, x: i128) -> Result<(), EncodeError> {
        self.outer.size += 16;
        Ok(())
    }

    fn encode_bytes(&mut self, x: &[u8]) -> Result<(), EncodeError> {
        self.outer.size += 4;
        self.outer.encode_bytes(x)
    }

    fn encode_time(&mut self, x: Time) -> Result<(), EncodeError> {
        self.encode_i128(x.unix_nanos())
    }

    fn encode_duration(&mut self, x: Duration) -> Result<(), EncodeError> {
        self.encode_i128(x.nanos())
    }

    fn encode_option(&mut self, visitor: Option<&dyn ValueEncodeVisitor>) -> Result<(), EncodeError> {
        self.outer.size += 1;
        if let Some(visitor) = visitor {
            visitor.encode(self)?;
        }
        Ok(())
    }
}

#[cfg(test)]
mod tests {
    use crate::binary::encoder::encode_value;
    use crate::encoder::Encoder;
    use crate::mem::MemoryManager;

    #[test]
    fn test_u32_size() {
        let mut sizer = crate::binary::encoder::EncodeSizer { size: 0 };
        sizer.encode_u32(10).unwrap();
        assert_eq!(sizer.size, 4);
    }

    #[test]
    fn test_u32_encode() {
        let x = 10u32;
        let mem = MemoryManager::new();
        let res = encode_value(&x, &mem).unwrap();
        assert_eq!(res, &[10, 0, 0, 0]);
    }
}

