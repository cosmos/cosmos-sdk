use alloc::string::String;
use bump_scope::{BumpScope, BumpString};
use crate::decoder::DecodeError;
use crate::list::ListVisitor;
use crate::r#struct::StructDecodeVisitor;
use crate::value::ArgValue;

pub fn decode_value<'b, 'a: 'b, V: ArgValue<'a>>(input: &'a [u8], scope: &'b mut BumpScope<'a>) -> Result<(V, Option<V::MemoryHandle>), DecodeError> {
    let mut decoder = Decoder { buf: input, scope };
    let mut decode_state = V::DecodeState::default();
    V::visit_decode_state(&mut decode_state, &mut decoder)?;
    V::finish_decode_state(decode_state)
}

struct Decoder<'b, 'a: 'b> {
    buf: &'a [u8],
    scope: &'b mut BumpScope<'a>,
}

impl<'b, 'a: 'b> Decoder<'b, 'a> {
    fn read_bytes(&mut self, size: usize) -> Result<&'a [u8], DecodeError> {
        if self.buf.len() < size {
            return Err(DecodeError::OutOfData);
        }
        let bz = &self.buf[0..size];
        self.buf = &self.buf[size..];
        Ok(bz)
    }
}

impl<'b, 'a: 'b> crate::decoder::Decoder<'a> for Decoder<'b, 'a> {
    fn decode_u32(&mut self) -> Result<u32, DecodeError> {
        let bz = self.read_bytes(4)?;
        Ok(u32::from_le_bytes(bz.try_into().unwrap()))
    }

    fn decode_u128(&mut self) -> Result<u128, DecodeError> {
        let bz = self.read_bytes(16)?;
        Ok(u128::from_le_bytes(bz.try_into().unwrap()))
    }

    fn decode_borrowed_str(&mut self) -> Result<Result<&'a str, BumpString<'a, 'a>>, DecodeError> {
        let bz = self.buf;
        self.buf = &[];
        Ok(Ok(core::str::from_utf8(bz).map_err(|_| DecodeError::InvalidData)?))
    }

    fn decode_owned_str(&mut self) -> Result<String, DecodeError> {
        let bz = self.buf;
        self.buf = &[];
        Ok(String::from_utf8(bz.to_vec()).map_err(|_| DecodeError::InvalidData)?)
    }

    fn decode_struct<V: StructDecodeVisitor<'a>>(&mut self, visitor: &mut V) -> Result<(), DecodeError> {
        let size = self.decode_u32()? as usize;
        let bz = self.read_bytes(size)?;
        let mut sub = Decoder { buf: bz, scope: self.scope };
        let mut inner = InnerDecoder { outer: &mut sub };
        inner.decode_struct(visitor)
    }

    fn decode_list<T, V: ListVisitor<'a, T>>(&mut self, visitor: &mut V) -> Result<(), DecodeError> {
        // let size = self.decode_u32()? as usize;
        // visitor.init(size, &mut self.scope)?;
        // let mut inner = InnerDecoder { outer: self };
        // for _ in 0..size {
        //     visitor.next(&mut inner)?;
        // }
        todo!();
        Ok(())
    }

    fn scope(&self) -> &'a BumpScope<'a> {
        // &self.scope
        todo!();
    }
}

struct InnerDecoder<'b, 'a: 'b> {
    outer: &'b mut Decoder<'b, 'a>,
}
impl<'b, 'a: 'b> crate::decoder::Decoder<'a> for InnerDecoder<'b, 'a> {
    fn decode_u32(&mut self) -> Result<u32, DecodeError> {
        self.outer.decode_u32()
    }

    fn decode_u128(&mut self) -> Result<u128, DecodeError> {
        self.outer.decode_u128()
    }

    fn decode_borrowed_str(&mut self) -> Result<Result<&'a str, BumpString<'a, 'a>>, DecodeError> {
        let size = self.decode_u32()? as usize;
        let bz = self.outer.read_bytes(size)?;
        Ok(Ok(core::str::from_utf8(bz).map_err(|_| DecodeError::InvalidData)?))
    }

    fn decode_owned_str(&mut self) -> Result<String, DecodeError> {
        let size = self.decode_u32()? as usize;
        let bz = self.outer.read_bytes(size)?;
        Ok(String::from_utf8(bz.to_vec()).map_err(|_| DecodeError::InvalidData)?)
    }

    fn decode_struct<V: StructDecodeVisitor<'a>>(&mut self, visitor: &mut V) -> Result<(), DecodeError> {
        // let size = self.decode_u32()? as usize;
        // let bz = self.outer.read_bytes(size)?;
        // let dec = Decoder { buf: bz, scope: self.outer.scope };
        todo!()
    }

    fn decode_list<T, V: ListVisitor<'a, T>>(&mut self, visitor: &mut V) -> Result<(), DecodeError> {
        todo!()
    }

    fn scope(&self) -> &'a BumpScope<'a> {
        todo!()
    }
}

#[cfg(test)]
mod tests {
    use bump_scope::Bump;
    use crate::binary::decoder::decode_value;
    use crate::decoder::Decoder;
    use crate::encoder::Encoder;

    #[test]
    fn test_u32_decode() {
        let buf: [u8; 4] = [10, 0, 0, 0];
        let mut bump = Bump::new();
        let mut scope = bump.as_mut_scope();
        let (x, _) = decode_value::<u32>(&buf, &mut scope).unwrap();
        assert_eq!(x, 10);
    }

    #[test]
    fn test_decode_string() {}
}
