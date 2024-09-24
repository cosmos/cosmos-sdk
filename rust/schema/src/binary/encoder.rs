use bump_scope::BumpScope;
use crate::encoder::{EncodeError};
use crate::r#struct::{StructDecodeVisitor, StructEncodeVisitor};
use crate::value::Value;
use crate::buffer::{ReverseWriter, ReverseWriterFactory, Writer};

pub fn encode_value<'a, V: Value<'a>, F: ReverseWriterFactory>(value: &V, writer_factory: &F) -> Result<<<F as ReverseWriterFactory>::Writer as ReverseWriter>::Output, EncodeError> {
    let mut sizer = EncodeSizer { size: 0 };
    <V as Value<'a>>::encode(value, &mut sizer)?;
    let mut writer = writer_factory.new(sizer.size);
    let mut encoder = Encoder { writer: &mut writer };
    <V as Value<'a>>::encode(value, &mut encoder)?;
    writer.finish()
}

struct Encoder<'a, W> {
    writer: &'a mut W
}

impl <'a, W: ReverseWriter> crate::encoder::Encoder for Encoder<'a, W> {
    fn encode_u32(&mut self, x: u32) -> Result<(), EncodeError> {
        self.writer.write(&x.to_le_bytes())
    }

    fn encode_u128(&mut self, x: u128) -> Result<(), EncodeError> {
        self.writer.write(&x.to_le_bytes())
    }

    fn encode_str(&mut self, x: &str) -> Result<(), EncodeError> {
        self.writer.write(x.as_bytes())
    }

    fn encode_list_slice<'c, V: Value<'c>>(&mut self, x: &[V]) -> Result<(), EncodeError> {
        for v in x.iter().rev() {
            <V as Value>::encode(v, self)?;
        }
        self.writer.write(&(x.len() as u32).to_le_bytes())?;
        Ok(())
    }

    fn encode_struct<V: StructEncodeVisitor>(&mut self, visitor: &V) -> Result<(), EncodeError> {
        let mut i = V::FIELDS.len();
        let mut sub = Encoder { writer: self.writer };
        let mut inner = InnerEncoder::<W> { outer: &mut sub };
        for f in V::FIELDS.iter().rev() {
            i -= 1;
            visitor.encode_field(i, &mut inner)?;
        }
        Ok(())
    }
}

struct EncodeSizer {
    size: usize
}

impl crate::encoder::Encoder for EncodeSizer {
    fn encode_u32(&mut self, x: u32) -> Result<(), EncodeError> {
        self.size += 4;
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

    fn encode_struct<'a, V: StructEncodeVisitor>(&mut self, visitor: &V) -> Result<(), EncodeError> {
        let mut i = 0;
        let mut sub = InnerEncodeSizer { outer: self };
        for f in V::FIELDS {
            visitor.encode_field(i, &mut sub)?;
            i += 1;
        }
        Ok(())
    }
}

struct InnerEncoder<'b, 'a:'b, W> {
    outer: &'b mut Encoder<'a, W>
}

impl <'b, 'a:'b, W: ReverseWriter> crate::encoder::Encoder for InnerEncoder<'a, 'b, W> {
    fn encode_u32(&mut self, x: u32) -> Result<(), EncodeError> {
        self.outer.encode_u32(x)
    }

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

    fn encode_struct<V: StructEncodeVisitor>(&mut self, visitor: &V) -> Result<(), EncodeError> {
        let end_pos = self.outer.writer.pos(); // this is a reverse writer so we start at the end
        self.outer.encode_struct(visitor)?;
        let start_pos = self.outer.writer.pos(); // now we know the start position
        let len = (end_pos - start_pos) as u32;
        self.outer.encode_u32(len)
    }
}

struct InnerEncodeSizer<'a> {
    outer: &'a mut EncodeSizer
}

impl <'a> crate::encoder::Encoder for InnerEncodeSizer<'a> {
    fn encode_u32(&mut self, x: u32) -> Result<(), EncodeError> {
        self.outer.size += 4;
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

    fn encode_struct<V: StructEncodeVisitor>(&mut self, visitor: &V) -> Result<(), EncodeError> {
        self.outer.size += 4;
        self.outer.encode_struct(visitor)
    }
}

#[cfg(test)]
mod tests {
    use bump_scope::Bump;
    use crate::binary::encoder::encode_value;
    use crate::encoder::Encoder;

    #[test]
    fn test_u32_size() {
        let mut sizer = crate::binary::encoder::EncodeSizer { size: 0 };
        sizer.encode_u32(10).unwrap();
        assert_eq!(sizer.size, 4);
    }

    #[test]
    fn test_u32_encode() {
        let x = 10u32;
        let bump = Bump::new();
        let res = encode_value(&x, bump.as_scope()).unwrap();
        assert_eq!(res.as_slice(), &[10, 0, 0, 0]);
    }
}
