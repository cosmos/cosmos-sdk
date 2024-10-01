use bump_scope::BumpScope;
use ixc_message_api::AccountID;
use crate::encoder::{EncodeError};
use crate::structs::{StructDecodeVisitor, StructEncodeVisitor, StructType};
use crate::value::Value;
use crate::buffer::{Writer, WriterFactory};
use crate::state_object::ObjectValue;

pub fn encode_value<'a, V: Value<'a>, F: WriterFactory>(value: &V, writer_factory: F) -> Result<F::Output, EncodeError> {
    let mut sizer = EncodeSizer { size: 0 };
    <V as Value<'a>>::encode(value, &mut sizer)?;
    let mut writer = writer_factory.new_reverse(sizer.size)?;
    let mut encoder = Encoder { writer: &mut writer };
    <V as Value<'a>>::encode(value, &mut encoder)?;
    writer.finish()
}

// fn encode_object_value<'a, V: ObjectValue, F: WriterFactory>(value: V::In<'a>, writer_factory: &F) -> Result<F::Output, EncodeError> {
//     let mut sizer = EncodeSizer { size: 0 };
//     let mut inner = InnerEncodeSizer { outer: &mut sizer };
//     V::encode_reverse(&value, &mut inner)?;
//     let mut writer = writer_factory.new_reverse(sizer.size)?;
//     let mut encoder = Encoder { writer: &mut writer };
//     let mut inner = InnerEncoder { outer: &mut encoder };
//     V::encode_reverse(&value, &mut inner)?;
//     writer.finish()
// }

struct Encoder<'a, W> {
    writer: &'a mut W,
}

impl<'a, W: Writer> crate::encoder::Encoder for Encoder<'a, W> {
    fn encode_u32(&mut self, x: u32) -> Result<(), EncodeError> {
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

    fn encode_list_slice<'c, V: Value<'c>>(&mut self, x: &[V]) -> Result<(), EncodeError> {
        let mut sub = Encoder { writer: self.writer };
        let mut inner = InnerEncoder::<W> { outer: &mut sub };
        for v in x.iter().rev() {
            <V as Value>::encode(v, &mut inner)?;
        }
        self.encode_u32(x.len() as u32)?;
        Ok(())
    }

    fn encode_struct<V: StructEncodeVisitor>(&mut self, visitor: &V, struct_type: &StructType) -> Result<(), EncodeError>  {
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
        self.encode_u64(x.get())
    }
}

struct EncodeSizer {
    size: usize,
}

impl crate::encoder::Encoder for EncodeSizer {
    fn encode_u32(&mut self, x: u32) -> Result<(), EncodeError> {
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

    fn encode_list_slice<'a, V: Value<'a>>(&mut self, xs: &[V]) -> Result<(), EncodeError> {
        self.size += 4;
        let mut sub = InnerEncodeSizer { outer: self };
        for x in xs.iter() {
            <V as Value>::encode(x, &mut sub)?;
        }
        Ok(())
    }

    fn encode_struct<V: StructEncodeVisitor>(&mut self, visitor: &V, struct_type: &StructType) -> Result<(), EncodeError> {
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
}

struct InnerEncoder<'b, 'a: 'b, W> {
    outer: &'b mut Encoder<'a, W>,
}

impl<'b, 'a: 'b, W: Writer> crate::encoder::Encoder for InnerEncoder<'a, 'b, W> {
    fn encode_u32(&mut self, x: u32) -> Result<(), EncodeError> {
        self.outer.encode_u32(x)
    }

    fn encode_u64(&mut self, x: u64) -> Result<(), EncodeError> { self.outer.encode_u64(x) }

    fn encode_u128(&mut self, x: u128) -> Result<(), EncodeError> {
        self.outer.encode_u128(x)
    }

    fn encode_str(&mut self, x: &str) -> Result<(), EncodeError> {
        self.outer.encode_str(x)?;
        self.encode_u32(x.len() as u32)
    }

    fn encode_list_slice<'c, V: Value<'c>>(&mut self, x: &[V]) -> Result<(), EncodeError> {
        self.outer.encode_list_slice(x)
        // TODO: prefix with length of actual encoded data
    }

    fn encode_struct<V: StructEncodeVisitor>(&mut self, visitor: &V, struct_type: &StructType) -> Result<(), EncodeError> {
        let end_pos = self.outer.writer.pos(); // this is a reverse writer so we start at the end
        self.outer.encode_struct(visitor, struct_type)?;
        let start_pos = self.outer.writer.pos(); // now we know the start position
        let len = (end_pos - start_pos) as u32;
        self.outer.encode_u32(len)
    }

    fn encode_account_id(&mut self, x: AccountID) -> Result<(), EncodeError> {
        self.outer.encode_account_id(x)
    }
}

struct InnerEncodeSizer<'a> {
    outer: &'a mut EncodeSizer,
}

impl<'a> crate::encoder::Encoder for InnerEncodeSizer<'a> {
    fn encode_u32(&mut self, x: u32) -> Result<(), EncodeError> {
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

    fn encode_list_slice<'b, V: Value<'b>>(&mut self, xs: &[V]) -> Result<(), EncodeError> {
        self.outer.size += 4; // for the for bytes size
        self.outer.encode_list_slice(xs)
    }

    fn encode_struct<V: StructEncodeVisitor>(&mut self, visitor: &V, struct_type: &StructType) -> Result<(), EncodeError>  {
        self.outer.size += 4;
        self.outer.encode_struct(visitor, struct_type)
    }

    fn encode_account_id(&mut self, x: AccountID) -> Result<(), EncodeError> {
        self.outer.size += 8;
        Ok(())
    }
}

#[cfg(test)]
mod tests {
    use allocator_api2::alloc::Allocator;
    use bump_scope::Bump;
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
        let res = encode_value(&x, &mem as &dyn Allocator).unwrap();
        assert_eq!(res, &[10, 0, 0, 0]);
    }
}
